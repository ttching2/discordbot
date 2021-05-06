package twittercommands

import (
	"context"
	"discordbot/botcommands"
	"discordbot/repositories"
	"discordbot/repositories/model"
	botTwitter "discordbot/twitter"
	"discordbot/util"
	"fmt"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/dghubble/go-twitter/twitter"
	log "github.com/sirupsen/logrus"
)

const TwitterFollowString = "twitter-follow"

type twitterFollowCommandFactory struct {
	twitterClient *botTwitter.TwitterClient
	repo          repositories.TwitterFollowRepository
	session       disgord.Session
}

func NewTwitterFollowCommandFactory(session disgord.Session, twitterClient *botTwitter.TwitterClient, repo repositories.TwitterFollowRepository) *twitterFollowCommandFactory {
	return &twitterFollowCommandFactory{
		twitterClient: twitterClient,
		repo:          repo,
		session:       session,
	}
}

func (c *twitterFollowCommandFactory) CreateFollowCommand(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &twitterFollowCommand{
		twitterFollowCommandFactory: c,
		data:                        data,
		user:                        user,
	}
}

type twitterFollowCommand struct {
	*twitterFollowCommandFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c *twitterFollowCommandFactory) PrintHelp() string {
	return botcommands.CommandPrefix + TwitterFollowString + " {screen_name} {channel_name} - have the bot follow a given user on Twitter and post new Tweets to a given channel."
}

func (c *twitterFollowCommand) ExecuteMessageCreateCommand() {
	msg := c.data.Message

	command := strings.Split(c.data.Message.Content, " ")
	if len(command) != 2 {
		msg.Reply(context.Background(), c.session, "Missing screen name of person to follow. Command use !twitter-follow screenName channel")
		return
	}
	screenName := command[0]
	channelName := command[1]

	userID := c.twitterClient.SearchForUser(screenName)
	if userID == "" {
		msg.React(context.Background(), c.session, "👎")
		msg.Reply(context.Background(), c.session, "Twitter screen name not found.")
		return
	}

	guild := c.session.Guild(msg.GuildID)
	channel := util.FindChannelByName(channelName, guild)
	if channel != nil {
		twitterFollowCommand := model.TwitterFollowCommand{
			User:         c.user.UsersID,
			ScreenName:   screenName,
			Channel:      channel.ID,
			Guild:        msg.GuildID,
			ScreenNameID: userID,
		}
		c.twitterClient.AddUserToTrack(userID)
		err := c.repo.SaveUserToFollow(&twitterFollowCommand)
		if err != nil {
			log.WithField("twitterFollowCommand", twitterFollowCommand).Error(err)
			msg.React(context.Background(), c.session, "👎")
			return
		}
		msg.React(context.Background(), c.session, "👍")
	} else {
		msg.React(context.Background(), c.session, "👎")
		msg.Reply(context.Background(), c.session, "Channel not found.")
	}
}

func RestartTwitterFollows(client disgord.Session, dbClient repositories.TwitterFollowRepository, twitterClient *botTwitter.TwitterClient) {
	tweetHandler := func(tweet *twitter.Tweet) {
		if tweet.InReplyToScreenName != "" {
			return
		}

		discordMessage := fmt.Sprintf("New Tweet by **%s** \nhttps://twitter.com/%s/status/%s", tweet.User.Name, tweet.User.ScreenName, tweet.IDStr)

		newMessageParams := &disgord.CreateMessageParams{
			Content: discordMessage,
		}

		twitterFollowCommands, err := dbClient.GetFollowedUser(tweet.User.ScreenName)

		if err != nil {
			log.WithField("twitterScreenName", tweet.User.ScreenName).Error(err)
			return
		}

		for i := range twitterFollowCommands {
			client.Channel(twitterFollowCommands[i].Channel).CreateMessage(newMessageParams)
		}

	}
	twitterClient.SetTweetDemux(tweetHandler)

	uniqueFollowedUsers, err := dbClient.GetAllUniqueFollowedUsers()
	if err != nil {
		log.Error(err)
		return
	}

	var followedUsers []string
	for _, followed := range uniqueFollowedUsers {
		followedUsers = append(followedUsers, followed.ScreenNameID)
	}
	twitterClient.AddUsersToTrack(followedUsers)
}
