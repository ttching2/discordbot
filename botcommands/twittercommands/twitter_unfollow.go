package twittercommands

import (
	"context"
	"discordbot/botcommands"
	"discordbot/botcommands/discord"
	"discordbot/repositories"
	botTwitter "discordbot/twitter"

	"github.com/andersfylling/disgord"
)

const TwitterUnfollowString = "twitter-unfollow"

type TwitterUnfollowCommand struct {
	twitterClient *botTwitter.TwitterClient
	repo          repositories.TwitterFollowRepository
}

func (c *TwitterUnfollowCommand) PrintHelp() string {
	return botcommands.CommandPrefix + TwitterUnfollowString + " {screen_name} - unfollows Twitter user given the twitter users username."
}

func NewTwitterUnfollowCommand(twitterClient *botTwitter.TwitterClient, repo repositories.TwitterFollowRepository) *TwitterUnfollowCommand {
	return &TwitterUnfollowCommand{
		twitterClient: twitterClient,
		repo: repo,
	}
}

func (c *TwitterUnfollowCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, middleWareContent discord.MiddleWareContent) {
	msg := data.Message
	
	followedUsers := c.repo.GetAllUniqueFollowedUsers()
	foundUser := false
	for _, user := range followedUsers {
		if user.ScreenName == middleWareContent.MessageContent {
			foundUser = true
			break
		}
	}
	
	if !foundUser {
		msg.Reply(context.Background(), s, "Screen name not being followed.")
		msg.React(context.Background(), s, "👎")
		return
	}

	c.repo.DeleteFollowedUser(middleWareContent.MessageContent, msg.GuildID)
	userID := c.twitterClient.SearchForUser(middleWareContent.MessageContent)
	c.twitterClient.RemoveUserFromFollowList(userID)
	msg.React(context.Background(), s, "👍")
}