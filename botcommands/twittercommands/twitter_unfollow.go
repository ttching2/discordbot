package twittercommands

import (
	"context"
	"discordbot/botcommands"
	"discordbot/repositories"
	"discordbot/repositories/model"
	botTwitter "discordbot/twitter"

	"github.com/andersfylling/disgord"
	log "github.com/sirupsen/logrus"
)

const TwitterUnfollowString = "twitter-unfollow"

type twitterUnfollowCommandFactory struct {
	twitterClient *botTwitter.TwitterClient
	repo          repositories.TwitterFollowRepository
	session       disgord.Session
}

func (c *twitterUnfollowCommandFactory) PrintHelp() string {
	return botcommands.CommandPrefix + TwitterUnfollowString + " {screen_name} - unfollows Twitter user given the twitter users username."
}

func NewTwitterUnfollowCommandFactory(session disgord.Session, twitterClient *botTwitter.TwitterClient, repo repositories.TwitterFollowRepository) *twitterUnfollowCommandFactory {
	return &twitterUnfollowCommandFactory{
		twitterClient: twitterClient,
		repo:          repo,
		session:       session,
	}
}

func (c *twitterUnfollowCommandFactory) CreateRequest(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &twitterUnfollowCommand{
		twitterUnfollowCommandFactory: c,
		data:                          data,
		user:                          user,
	}
}

type twitterUnfollowCommand struct {
	*twitterUnfollowCommandFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c *twitterUnfollowCommand) ExecuteMessageCreateCommand() {
	msg := c.data.Message

	followedUsers, err := c.repo.GetAllUniqueFollowedUsers()
	if err != nil {
		msg.React(context.Background(), c.session, "👎")
		log.Error(err)
		return
	}

	foundUser := false
	for _, user := range followedUsers {
		if user.ScreenName == msg.Content {
			foundUser = true
			break
		}
	}

	if !foundUser {
		msg.Reply(context.Background(), c.session, "Screen name not being followed.")
		msg.React(context.Background(), c.session, "👎")
		return
	}

	c.repo.DeleteFollowedUser(msg.Content, msg.GuildID)
	userID := c.twitterClient.SearchForUser(msg.Content)
	c.twitterClient.RemoveUserFromFollowList(userID)
	msg.React(context.Background(), c.session, "👍")
}
