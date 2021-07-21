package commands

import (
	"github.com/andersfylling/disgord"
)

type twitterFollowListCommandFactory struct {
	repo    TwitterFollowRepository
	session disgord.Session
}

func (c *twitterFollowListCommandFactory) PrintHelp() string {
	return CommandPrefix + TwitterFollowListString + " - lists all currently followed users for this discord."
}

func (c *twitterFollowCommandFactory) CreateFollowListRequest(data *disgord.MessageCreate, user *Users) interface{} {
	return &twitterFollowListCommand{
		twitterFollowCommandFactory: c,
		data:                        data,
		user:                        user,
	}
}

type twitterFollowListCommand struct {
	*twitterFollowCommandFactory
	data *disgord.MessageCreate
	user *Users
}

func (c *twitterFollowListCommand) ExecuteMessageCreateCommand() {
	followList := ""
	followsInGuild, err := c.repo.GetAllFollowedUsersInServer(c.data.Message.GuildID)
	if err != nil {
		c.session.ReactToMessage(c.data.Message.ID, c.data.Message.ChannelID, "👎")
		log.Error(err)
		return
	}
	for _, follows := range followsInGuild {
		followList += follows.ScreenName + "\n"
	}

	c.session.SendSimpleMessage(c.data.Message.ChannelID, "Following:\n"+followList)
}
