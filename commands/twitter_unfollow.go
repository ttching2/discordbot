package commands

import (
	botTwitter "discordbot/twitter"

	"github.com/andersfylling/disgord"
)

const TwitterUnfollowString = "twitter-unfollow"

type twitterUnfollowCommandFactory struct {
	twitterClient *botTwitter.TwitterClient
	repo          TwitterFollowRepository
	session       disgord.Session
}

func (c *twitterUnfollowCommandFactory) PrintHelp() string {
	return CommandPrefix + TwitterUnfollowString + " {screen_name} - unfollows Twitter user given the twitter users username."
}

func (c *twitterFollowCommandFactory) CreateUnfollowRequest(data *disgord.MessageCreate, user *Users) interface{} {
	return &twitterUnfollowCommand{
		twitterFollowCommandFactory: c,
		data:                        data,
		user:                        user,
	}
}

type twitterUnfollowCommand struct {
	*twitterFollowCommandFactory
	data *disgord.MessageCreate
	user *Users
}

func (c *twitterUnfollowCommand) ExecuteMessageCreateCommand() {
	msg := c.data.Message

	followedUsers, err := c.repo.GetAllUniqueFollowedUsers()
	if err != nil {
		c.session.ReactToMessage(msg.ID, msg.ChannelID, "👎")
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
		c.session.ReactToMessage(msg.ID, msg.ChannelID, "👎")
		c.session.SendSimpleMessage(msg.ChannelID, "Screen name not being followed.")
		return
	}

	c.repo.DeleteFollowedUser(msg.Content, msg.GuildID)
	userID := c.twitterClient.SearchForUser(msg.Content)
	c.twitterClient.RemoveUserFromFollowList(userID)
	c.session.ReactToMessage(msg.ID, msg.ChannelID, "👍")
}
