package strawpolldeadline

import (
	"context"
	"discordbot/botcommands"
	"discordbot/botcommands/discord"
	"discordbot/repositories"
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

type StrawpollDeadlineCommand struct {
	strawpollClient *strawpoll.Client
	repo            repositories.StrawpollDeadlineRepository
}

func (c *StrawpollDeadlineCommand) PrintHelp() string {
	return botcommands.CommandPrefix + StrawPollDeadlineString + "{strawpoll_url} {channel_name} {role_name} - Ping role in given channel when deadline is met and announce results."
}

func New(strawpollClient *strawpoll.Client, repo repositories.StrawpollDeadlineRepository) *StrawpollDeadlineCommand {
	return &StrawpollDeadlineCommand{
		strawpollClient: strawpollClient,
		repo:            repo,
	}
}

func (c *StrawpollDeadlineCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, middleWareContent discord.MiddleWareContent) {
	msg := data.Message

	split := strings.Split(middleWareContent.MessageContent, " ")
	if len(split) != 3 {
		msg.Reply(context.Background(), s, "Incorrect number of arguments for command.")
		return
	}

	u, err := url.Parse(split[0])
	if err != nil {
		msg.Reply(context.Background(), s, "Error processing strawpoll url.")
		return
	}

	pollID := u.Path[1:]
	poll, _ := c.strawpollClient.GetPoll(pollID)

	now := time.Now()
	if now.After(poll.Content.Deadline) {
		msg.Reply(context.Background(), s, "Could not set timer for poll. Deadline either missing or deadline has passed.")
		return
	}

	channelName := split[1]
	channels, _ := s.Guild(msg.GuildID).GetChannels()
	channel := util.FindChannelByName(channelName, channels)

	roleName := split[2]
	roles, _ := s.Guild(msg.GuildID).GetRoles()
	role := util.FindRoleByName(roleName, roles)

	deadlineDuration := poll.Content.Deadline.Sub(now)
	timeToWait := time.NewTimer(deadlineDuration)
	strawpollDeadline := &repositories.StrawpollDeadline{
		User:        middleWareContent.UsersID,
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
		s.Channel(channel.ID).CreateMessage(&disgord.CreateMessageParams{Content: result})
		err := c.repo.DeleteStrawpollDeadlineByID(strawpollDeadline.StrawpollDeadlineID)
		if err != nil {
			log.WithField("strawpoll", strawpollDeadline).Error(err)
		}
	}()

	msg.React(context.Background(), s, "👍")
}

func RestartStrawpollDeadlines(client *disgord.Client, dbClient repositories.StrawpollDeadlineRepository, strawpollClient *strawpoll.Client) {
	strawpolls, err := dbClient.GetAllStrawpollDeadlines() 
	if err != nil {
		log.Error(err)
		return
	}
	for _, strawpoll := range strawpolls{

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
		go func(strawpoll repositories.StrawpollDeadline) {
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
