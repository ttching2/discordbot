package twittercommands

import (
	"context"
	"discordbot/botcommands"
	"discordbot/repositories"
	"discordbot/repositories/model"

	"github.com/andersfylling/disgord"
	log "github.com/sirupsen/logrus"
)

const TwitterFollowListString = "twitter-follow-list"

type twitterFollowListCommandFactory struct {
	repo    repositories.TwitterFollowRepository
	session disgord.Session
}

func (c *twitterFollowListCommandFactory) PrintHelp() string {
	return botcommands.CommandPrefix + TwitterFollowListString + " - lists all currently followed users for this discord."
}

func NewTwitterFollowListCommandFactory(session disgord.Session, repo repositories.TwitterFollowRepository) *twitterFollowListCommandFactory {
	return &twitterFollowListCommandFactory{
		repo:    repo,
		session: session,
	}
}

func (c *twitterFollowListCommandFactory) CreateRequest(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &twitterFollowListCommand {
		twitterFollowListCommandFactory: c,
		data: data,
		user: user,
	}
}

type twitterFollowListCommand struct {
	*twitterFollowListCommandFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c *twitterFollowListCommand) ExecuteMessageCreateCommand() {
	followList := ""
	followsInGuild, err := c.repo.GetAllFollowedUsersInServer(c.data.Message.GuildID)
	if err != nil {
		c.data.Message.React(context.Background(), c.session, "👎")
		log.Error(err)
		return
	}
	for _, follows := range followsInGuild {
		followList += follows.ScreenName + "\n"
	}

	c.data.Message.Reply(context.Background(), c.session, "Following:\n"+followList)
}
