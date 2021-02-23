package discord

import (
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