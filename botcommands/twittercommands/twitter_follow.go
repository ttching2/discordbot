package twittercommands

import (
	"context"
	"discordbot/botcommands"
	"discordbot/botcommands/discord"
	"discordbot/repositories"
	botTwitter "discordbot/twitter"
	"discordbot/util"
	"fmt"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/dghubble/go-twitter/twitter"
	log "github.com/sirupsen/logrus"
)

const TwitterFollowString = "twitter-follow"

type TwitterFollowCommand struct {
	twitterClient *botTwitter.TwitterClient
	repo          repositories.TwitterFollowRepository
}

func NewTwitterFollowCommand(twitterClient *botTwitter.TwitterClient, repo repositories.TwitterFollowRepository) *TwitterFollowCommand {
	return &TwitterFollowCommand{
		twitterClient: twitterClient,
		repo:          repo,
	}
}

func (c *TwitterFollowCommand) PrintHelp() string {
	return botcommands.CommandPrefix + TwitterFollowListString + " {screen_name} {channel_name} - have the bot follow a given user on Twitter and post new Tweets to a given channel."
}

func (c *TwitterFollowCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, middleWareContent discord.MiddleWareContent) {
	msg := data.Message

	command := strings.Split(middleWareContent.MessageContent, " ")
	if len(command) != 2 {
		msg.Reply(context.Background(), s, "Missing screen name of person to follow. Command use !twitter-follow screenName channel")
		return
	}
	screenName := command[0]
	channelName := command[1]

	userID := c.twitterClient.SearchForUser(screenName)
	if userID == "" {
		msg.React(context.Background(), s, "👎")
		msg.Reply(context.Background(), s, "Twitter screen name not found.")
		return
	}

	channels, _ := s.Guild(msg.GuildID).GetChannels()
	channel := util.FindChannelByName(channelName, channels)
	if channel != nil {
		twitterFollowCommand := repositories.TwitterFollowCommand{
			User:         middleWareContent.UsersID,
			ScreenName:   screenName,
			Channel:      channel.ID,
			Guild:        msg.GuildID,
			ScreenNameID: userID,
		}
		c.twitterClient.AddUserToTrack(userID)
		err := c.repo.SaveUserToFollow(&twitterFollowCommand)
		if err != nil {
			log.WithField("twitterFollowCommand", twitterFollowCommand).Error(err)
			msg.React(context.Background(), s, "👎")
			return
		}
		msg.React(context.Background(), s, "👍")
	} else {
		msg.React(context.Background(), s, "👎")
		msg.Reply(context.Background(), s, "Channel not found.")
	}
}

func RestartTwitterFollows(client *disgord.Client, dbClient repositories.TwitterFollowRepository, twitterClient *botTwitter.TwitterClient) {
	tweetHandler := func(tweet *twitter.Tweet) {
		discordMessage := fmt.Sprintf("New Tweet by %s https://www.twitter.com/%s/status/%s", tweet.User.Name, tweet.User.ScreenName, tweet.IDStr)

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
	for _, followed := range  uniqueFollowedUsers{
		followedUsers = append(followedUsers, followed.ScreenNameID)
	}
	twitterClient.AddUsersToTrack(followedUsers)
}
