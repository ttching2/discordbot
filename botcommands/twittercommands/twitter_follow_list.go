package twittercommands

import (
	"context"
	"discordbot/botcommands"
	"discordbot/botcommands/discord"
	"discordbot/repositories"

	"github.com/andersfylling/disgord"
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
	followsInGuild := c.repo.GetAllFollowedUsersInServer(data.Message.GuildID)
	for _, follows := range followsInGuild{
		followList +=  follows.ScreenName + "\n"
	}

	data.Message.Reply(context.Background(), s, "Following:\n" + followList)
}