package botcommands

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"discordbot/strawpoll"
	myTwitter "discordbot/twitter"
	"discordbot/util"

	"github.com/andersfylling/disgord"
)

const CommandPrefix = "?"

const TwitterFollowString = "twitter-follow"
const TwitterUnfollowString = "twitter-unfollow"
const TwitterFollowListString = "twitter-follow-list"
const RoleReactString = "react"
const StrawPollDeadlineString = "strawpoll-deadline"
const HelpString = "help"

type baseCommand struct {
	Name string
	Description string
}

func (c *baseCommand) GetDescription() string {
	return c.Description
}

type Command interface {
	GetDescription() string
	ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, saveableCommand SaveableCommand)
}

type roleReactCommand struct {
	baseCommand
}

func NewRoleReactCommand() *roleReactCommand {
	r := &roleReactCommand {
		baseCommand: baseCommand{
			Name: RoleReactString,
			Description: CommandPrefix + "react - Creates a message with a reaction added by the bot. Anyone reacting to the message will receive a role attached to the message.",
		},
	}
	return r
}

func (c *roleReactCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, saveableCommand SaveableCommand) {
	msg := data.Message

	msg.Reply(context.Background(), s, "Which channel should this message be sent in.")
	command := CommandInProgress{
		Guild: msg.GuildID,
		User: msg.Author.ID,
		Stage: 1 }
	saveableCommand.SaveCommandInProgress(msg.Author, command)
}

type twitterFollowCommand struct {
	twitterClient *myTwitter.TwitterClient
	baseCommand
}

func NewTwitterFollowCommand(twitterClient *myTwitter.TwitterClient) Command {
	return &twitterFollowCommand{
		baseCommand: baseCommand{
			Name: TwitterFollowString,
			Description: CommandPrefix + "twitter-follow - Tells the bot to follow a twitter screen name and posts new tweets into a specified channel.",
		},
		twitterClient: twitterClient,
	}
}

func (r *twitterFollowCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, saveableCommand SaveableCommand) {
	msg := data.Message
	
	
	command := strings.Split(msg.Content, " ")
	if len(command) != 3 {
		msg.Reply(context.Background(), s, "Missing screen name of person to follow. Command use !twitter-follow screenName channel")
		return
	}
	screenName := command[1]
	channelName := command[2]

	channels, _ := s.Guild(msg.GuildID).GetChannels()
	for _, channel := range channels {
		if channel.Name == channelName {
			userID := r.twitterClient.SearchForUser(screenName)
			if userID == "" {
				msg.React(context.Background(), s, "👎")
				msg.Reply(context.Background(), s, "Twitter screen name not found.")
			}

			twitterFollowCommand := TwitterFollowCommand{
				ScreenName: screenName,
				Channel: channel.ID,
				Guild: msg.GuildID,
				ScreenNameID: userID,
			}
			r.twitterClient.AddUserToTrack(userID)
			saveableCommand.SaveUserToFollow(twitterFollowCommand)
			msg.React(context.Background(), s, "👍")
			return
		}
	}
	msg.React(context.Background(), s, "👎")
	msg.Reply(context.Background(), s, "Channel not found.")
}

type twitterUnfollowCommand struct {
	twitterClient *myTwitter.TwitterClient
	baseCommand
}

func NewTwitterUnfollowCommand(twitterClient *myTwitter.TwitterClient) Command {
	return &twitterUnfollowCommand{
		baseCommand: baseCommand{
			Name: TwitterUnfollowString,
			Description: CommandPrefix + "twitter-unfollow - Unfollow a twitter user and stop receiving posts.",
		},
		twitterClient: twitterClient,
	}
}

func (r *twitterUnfollowCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, saveableCommand SaveableCommand) {
	msg := data.Message
	command := strings.Split(msg.Content, " ")

	if len(command) < 2 {
		msg.Reply(context.Background(), s, "No screen name given.")
	}

	saveableCommand.DeleteFollowedUser(command[1], msg.GuildID)
	
	followedUsers := saveableCommand.GetAllUniqueFollowedUsers()
	for _, user := range followedUsers {
		if user.ScreenName == command[1] {
			msg.React(context.Background(), s, "👍")
			return
		}
	}
	
	userID := r.twitterClient.SearchForUser(command[1])
	r.twitterClient.RemoveUserFromFollowList(userID)
	msg.React(context.Background(), s, "👍")
}

type twitterFollowListCommand struct {
	baseCommand
}

func NewTwitterFollowListCommand() Command {
	return &twitterFollowListCommand{
		baseCommand: baseCommand{
			Name: TwitterFollowListString,
			Description: CommandPrefix + "twitter-follow-list - List all users currently being followed.",
		},
	}
}

func (r *twitterFollowListCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, saveableCommand SaveableCommand) {
	followList := ""
	followsInGuild := saveableCommand.GetAllFollowedUsersInServer(data.Message.GuildID)
	for _, follows := range followsInGuild{
		followList +=  follows.ScreenName + "\n"
	}

	data.Message.Reply(context.Background(), s, "Following:\n" + followList)
}

type strawPollDeadlineCommand struct {
	strawPollClient *strawpoll.Client
	baseCommand
}

func NewStrawPollCommand(strawpollClient *strawpoll.Client) Command {
	return &strawPollDeadlineCommand{
		strawPollClient: strawpollClient,
		baseCommand: baseCommand {
			Name: "strawpoll-deadline",
			Description: "!strawpoll-deadline {strawpoll-link} {channel} {role-mention}. Command to ping users the completiong and result of strawpoll.",
		},
	}
}

func (r *strawPollDeadlineCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, saveableCommand SaveableCommand) {
	msg := data.Message

	split := strings.Split(msg.Content, " ")
	if len(split) != 4 {
		msg.Reply(context.Background(), s, "Incorrect number of arguments for command.")
		return
	}

	u, err := url.Parse(split[1])
	if err != nil {
		msg.Reply(context.Background(), s, "Error processing strawpoll url.")
		return
	}

	pollID := u.Path[1:]
	poll, _ := r.strawPollClient.GetPoll(pollID)

	now := time.Now()
	if now.After(poll.Content.Deadline) {
		msg.Reply(context.Background(), s, "Could not set timer for poll. Deadline either missing or deadline has passed.")
		return
	}

	channelName := split[2]
	channels, _:= s.Guild(msg.GuildID).GetChannels()
	channel := util.FindChannelByName(channelName, channels)

	roleName := split[3]
	roles, _ := s.Guild(msg.GuildID).GetRoles()
	role := util.FindRoleByName(roleName, roles)

	deadlineDuration := poll.Content.Deadline.Sub(now)
	timeToWait := time.NewTimer(deadlineDuration)

	strawpollDeadline := saveableCommand.SaveStrawpollDeadline(&StrawpollDeadline{
		Guild: msg.GuildID,
		Channel: channel.ID,
		Role: role.ID,
		StrawpollID: pollID,
	})
	go func() {
		<-timeToWait.C
		poll, _ := r.strawPollClient.GetPoll(pollID)
		pollAnswers := poll.Content.Poll.PollAnswers
		topAnswer := pollAnswers[0]
		for _, answer := range pollAnswers {
			if answer.Votes > topAnswer.Votes {
				topAnswer = answer
			}
		}
		result := fmt.Sprintf("%s Strawpoll has closed. The top vote for %s is %s with %d votes.", role.Mention(), poll.Content.Title, topAnswer.Answer, topAnswer.Votes)
		s.Channel(channel.ID).CreateMessage(&disgord.CreateMessageParams{Content: result})
		saveableCommand.DeleteStrawpollDeadlineByID(strawpollDeadline.StrawpollDeadlineID)
	}()

	msg.React(context.Background(), s, "👍")
}


type helpCommand struct {
	Commands []Command
	baseCommand
}

func NewHelpCommand(commands map[string]Command) Command {
	var commandArray []Command
	for _, c := range commands {
		commandArray = append(commandArray, c)
	}
	return &helpCommand{
		baseCommand: baseCommand{
			Name: HelpString,
			Description: "",
		},
		Commands: commandArray,
	}
}

func (r *helpCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, saveableCommand SaveableCommand) {
	helpList := "```Available Commands:\n"
	for _ , c := range r.Commands {
		helpList += c.GetDescription() + "\n"
	}
	helpList += "```"
	data.Message.Reply(context.Background(), s, helpList)
}