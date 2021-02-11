package twittercommands

import (
	"context"
	"discordbot/botcommands"
	"discordbot/botcommands/discord"
	"discordbot/repositories"

	"github.com/andersfylling/disgord"
	log "github.com/sirupsen/logrus"
)

const TwitterFollowListString = "twitter-follow-list"

type TwitterFollowListCommand struct {
	repo repositories.TwitterFollowRepository
}

func (c *TwitterFollowListCommand) PrintHelp() string {
	return botcommands.CommandPrefix + TwitterFollowListString + " - lists all currently followed users for this discord."
}

func NewTwitterFollowListCommand(repo repositories.TwitterFollowRepository) *TwitterFollowListCommand {
	return &TwitterFollowListCommand{
		repo: repo,
	}
}

func (c *TwitterFollowListCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, middleWareContent discord.MiddleWareContent) {
	followList := ""
	followsInGuild, err := c.repo.GetAllFollowedUsersInServer(data.Message.GuildID)
	if err != nil {
		data.Message.React(context.Background(), s, "👎")
		log.Error(err)
		return
	}
	for _, follows := range followsInGuild{
		followList +=  follows.ScreenName + "\n"
	}

	data.Message.Reply(context.Background(), s, "Following:\n" + followList)
}