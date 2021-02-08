package discord

import (
	"context"
	"discordbot/repositories"

	"github.com/andersfylling/disgord"
)

/*
TODO Probably rename this package and some other stuff
*/

type Guild interface {
	GetChannels(...disgord.Flag) ([]*disgord.Channel, error)
	GetRoles(flags ...disgord.Flag) ([]*disgord.Role, error)
	GetEmojis(flags ...disgord.Flag) ([]*disgord.Emoji, error)
}

type MessageCreateHandler interface {
	ExecuteCommand(disgord.Session, *disgord.MessageCreate, MiddleWareContent)
	PrintHelp() string
}

type DiscordMessageInfo struct {
	Content   string
	UserID    repositories.Snowflake
	AuthorID  int64
	ChannelID repositories.Snowflake
	Reply     func(ctx context.Context, s disgord.Session, data ...interface{}) (*disgord.Message, error)
}

type MiddleWareContent struct {
	Command        string
	MessageContent string
	UsersID        int64
}
