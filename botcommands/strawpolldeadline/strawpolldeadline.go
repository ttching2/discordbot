package strawpolldeadline

import (
	"context"
	"discordbot/botcommands"
	"discordbot/repositories"
	"discordbot/repositories/model"
	"discordbot/strawpoll"
	"discordbot/util"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/andersfylling/disgord"
	log "github.com/sirupsen/logrus"
)

const StrawPollDeadlineString = "strawpoll-deadline"

type strawpollDeadlineCommandFactory struct {
	strawpollClient *strawpoll.Client
	repo            repositories.StrawpollDeadlineRepository
	session         disgord.Session
}

func (c *strawpollDeadlineCommandFactory) PrintHelp() string {
	return botcommands.CommandPrefix + StrawPollDeadlineString + "{strawpoll_url} {channel_name} {role_name} - Ping role in given channel when deadline is met and announce results."
}

func NewCommandFactory(session disgord.Session, strawpollClient *strawpoll.Client, repo repositories.StrawpollDeadlineRepository) *strawpollDeadlineCommandFactory {
	return &strawpollDeadlineCommandFactory{
		strawpollClient: strawpollClient,
		repo:            repo,
		session:         session,
	}
}

func (c *strawpollDeadlineCommandFactory) CreateRequest(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &strawpollDeadlineCommand {
		strawpollDeadlineCommandFactory: c,
		data: data,
		user: user,
	}
}

type strawpollDeadlineCommand struct {
	*strawpollDeadlineCommandFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c *strawpollDeadlineCommand) ExecuteMessageCreateCommand() {
	msg := c.data.Message

	split := strings.Split(msg.Content, " ")
	if len(split) != 3 {
		msg.Reply(context.Background(), c.session, "Incorrect number of arguments for command.")
		return
	}

	u, err := url.Parse(split[0])
	if err != nil {
		msg.Reply(context.Background(), c.session, "Error processing strawpoll url.")
		return
	}

	pollID := u.Path[1:]
	poll, _ := c.strawpollClient.GetPoll(pollID)

	now := time.Now()
	if now.After(poll.Content.Deadline) {
		msg.Reply(context.Background(), c.session, "Could not set timer for poll. Deadline either missing or deadline has passed.")
		return
	}

	channelName := split[1]
	guild := c.session.Guild(msg.GuildID)
	channel := util.FindChannelByName(channelName, guild)

	roleName := split[2]
	roles, _ := c.session.Guild(msg.GuildID).GetRoles()
	role := util.FindRoleByName(roleName, roles)

	deadlineDuration := poll.Content.Deadline.Sub(now)
	timeToWait := time.NewTimer(deadlineDuration)
	strawpollDeadline := &model.StrawpollDeadline{
		User:        c.user.UsersID,
		Guild:       msg.GuildID,
		Channel:     channel.ID,
		Role:        role.ID,
		StrawpollID: pollID,
	}
	c.repo.SaveStrawpollDeadline(strawpollDeadline)
	go func() {
		<-timeToWait.C
		poll, _ := c.strawpollClient.GetPoll(pollID)
		pollAnswers := poll.Content.Poll.PollAnswers
		topAnswer := pollAnswers[0]
		for _, answer := range pollAnswers {
			if answer.Votes > topAnswer.Votes {
				topAnswer = answer
			}
		}
		result := fmt.Sprintf("%s Strawpoll has closed. The top vote for %s is %s with %d votes.", role.Mention(), poll.Content.Title, topAnswer.Answer, topAnswer.Votes)
		c.session.Channel(channel.ID).CreateMessage(&disgord.CreateMessageParams{Content: result})
		err := c.repo.DeleteStrawpollDeadlineByID(strawpollDeadline.StrawpollDeadlineID)
		if err != nil {
			log.WithField("strawpoll", strawpollDeadline).Error(err)
		}
	}()

	msg.React(context.Background(), c.session, "👍")
}

func RestartStrawpollDeadlines(client disgord.Session, dbClient repositories.StrawpollDeadlineRepository, strawpollClient *strawpoll.Client) {
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
		if now.After(poll.Content.Deadline) {
			dbClient.DeleteStrawpollDeadlineByID(strawpoll.StrawpollDeadlineID)
			continue
		}

		timeToWait := time.NewTimer(poll.Content.Deadline.Sub(now))
		go func(strawpoll model.StrawpollDeadline) {
			<-timeToWait.C
			poll, _ := strawpollClient.GetPoll(strawpoll.StrawpollID)
			pollAnswers := poll.Content.Poll.PollAnswers
			topAnswer := pollAnswers[0]
			for _, answer := range pollAnswers {
				if answer.Votes > topAnswer.Votes {
					topAnswer = answer
				}
			}
			result := fmt.Sprintf("<@&%s> Strawpoll has closed. The top vote for %s is %s with %d votes.", strawpoll.Role, poll.Content.Title, topAnswer.Answer, topAnswer.Votes)
			client.Channel(strawpoll.Channel).CreateMessage(&disgord.CreateMessageParams{Content: result})
			dbClient.DeleteStrawpollDeadlineByID(strawpoll.StrawpollDeadlineID)
		}(strawpoll)
	}
}
