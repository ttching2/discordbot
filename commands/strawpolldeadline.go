package commands

import (
	"discordbot/strawpoll"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/andersfylling/disgord"
)

type strawpollDeadlineCommandFactory struct {
	strawpollClient *strawpoll.Client
	repo            StrawpollDeadlineRepository
	session         DiscordSession
}

func (c *strawpollDeadlineCommandFactory) PrintHelp() string {
	return CommandPrefix + StrawPollDeadlineString + "{strawpoll_url} {channel_name} {role_name} - Ping role in given channel when deadline is met and announce results."
}

func NewCommandFactory(session DiscordSession, strawpollClient *strawpoll.Client, repo StrawpollDeadlineRepository) *strawpollDeadlineCommandFactory {
	return &strawpollDeadlineCommandFactory{
		strawpollClient: strawpollClient,
		repo:            repo,
		session:         session,
	}
}

func (c *strawpollDeadlineCommandFactory) CreateRequest(data *disgord.MessageCreate, user *Users) interface{} {
	return &strawpollDeadlineCommand {
		strawpollDeadlineCommandFactory: c,
		data: data,
		user: user,
	}
}

type strawpollDeadlineCommand struct {
	*strawpollDeadlineCommandFactory
	data *disgord.MessageCreate
	user *Users
}

func (c *strawpollDeadlineCommand) ExecuteMessageCreateCommand() {
	msg := c.data.Message

	split := strings.Split(msg.Content, " ")
	if len(split) < 3 {
		c.session.SendSimpleMessage(msg.ChannelID, "Incorrect number of arguments for command.")
		return
	}

	u, err := url.Parse(split[0])
	if err != nil {
		c.session.SendSimpleMessage(msg.ChannelID, "Error processing strawpoll url.")
		return
	}

	pollID := u.Path[1:]
	poll, _ := c.strawpollClient.GetPoll(pollID)

	now := time.Now()
	pollDeadline := time.Unix(poll.Poll.PollConfig.DeadlineAt, 0)
	if now.After(pollDeadline) {
		c.session.SendSimpleMessage(msg.ChannelID, "Could not set timer for poll. Deadline either missing or deadline has passed.")
		return
	}

	channelName := split[1]
	guild := c.session.Guild(msg.GuildID)
	channel := FindChannelByName(channelName, guild)

	roleName := strings.Join(split[2:]," ")
	roles, _ := c.session.Guild(msg.GuildID).GetRoles()
	role := FindRoleByName(roleName, roles)

	deadlineDuration := pollDeadline.Sub(now)
	timeToWait := time.NewTimer(deadlineDuration)
	strawpollDeadline := &StrawpollDeadline{
		User:        c.user.UsersID,
		Guild:       msg.GuildID,
		Channel:     channel.ID,
		Role:        role.ID,
		StrawpollID: pollID,
	}
	c.repo.SaveStrawpollDeadline(strawpollDeadline)
	go func() {
		<-timeToWait.C
		poll, err := c.strawpollClient.GetPoll(pollID)
		if err != nil {
			log.WithField("pollid",pollID).Error("Error fetching strawpoll ", err)
			return
		}
		pollAnswers := poll.Poll.PollOptions
		topAnswer := pollAnswers[0]
		for _, answer := range pollAnswers {
			if answer.VoteCount > topAnswer.VoteCount {
				topAnswer = answer
			}
		}
		result := fmt.Sprintf("%s Strawpoll has closed. The top vote for %s is %s with %d votes.", role.Mention(), poll.Poll.Title, topAnswer.Value, topAnswer.VoteCount)
		c.session.SendSimpleMessage(channel.ID, result)
		err = c.repo.DeleteStrawpollDeadlineByID(strawpollDeadline.StrawpollDeadlineID)
		if err != nil {
			log.WithField("strawpoll", strawpollDeadline).Error(err)
		}
	}()

	c.session.ReactToMessage(msg.ID, msg.ChannelID, "👍")
}

func RestartStrawpollDeadlines(client disgord.Session, dbClient StrawpollDeadlineRepository, strawpollClient *strawpoll.Client) {
	strawpolls, err := dbClient.GetAllStrawpollDeadlines()
	if err != nil {
		log.Error(err)
		return
	}
	for _, strawpoll := range strawpolls {

		poll, err := strawpollClient.GetPoll(strawpoll.StrawpollID)
		if err != nil {
			dbClient.DeleteStrawpollDeadlineByID(strawpoll.StrawpollDeadlineID)
			continue
		}

		now := time.Now()
		pollDeadline := time.Unix(poll.Poll.PollConfig.DeadlineAt, 0)
		if now.After(pollDeadline) {
			dbClient.DeleteStrawpollDeadlineByID(strawpoll.StrawpollDeadlineID)
			continue
		}

		timeToWait := time.NewTimer(pollDeadline.Sub(now))
		go func(strawpoll StrawpollDeadline) {
			<-timeToWait.C
			poll, _ := strawpollClient.GetPoll(strawpoll.StrawpollID)
			pollAnswers := poll.Poll.PollOptions
			topAnswer := pollAnswers[0]
			for _, answer := range pollAnswers {
				if answer.VoteCount > topAnswer.VoteCount {
					topAnswer = answer
				}
			}
			result := fmt.Sprintf("<@&%s> Strawpoll has closed. The top vote for %s is %s with %d votes.", strawpoll.Role, poll.Poll.Title, topAnswer.Value, topAnswer.VoteCount)
			client.Channel(strawpoll.Channel).CreateMessage(&disgord.CreateMessageParams{Content: result})
			dbClient.DeleteStrawpollDeadlineByID(strawpoll.StrawpollDeadlineID)
		}(strawpoll)
	}
}
