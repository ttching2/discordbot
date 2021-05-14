package commands

import (
	"context"

	"github.com/andersfylling/disgord"
)

/*
TODO Probably rename this package and some other stuff
*/

type DiscordSession interface {
	SendSimpleMessage(Snowflake, string) (*disgord.Message, error)
	ReactToMessage(msg Snowflake, channel Snowflake, emoji interface{})
	CurrentUser() (*disgord.User, error)
	Guild(Snowflake) Guild
}

func NewSimpleDiscordSession(s disgord.Session) *simpleDiscordSession{
	return &simpleDiscordSession{disgordSession: s}
}

type simpleDiscordSession struct {
	disgordSession disgord.Session
}

func (s *simpleDiscordSession) SendSimpleMessage(channel Snowflake, msg string) (*disgord.Message, error) {
	return s.disgordSession.WithContext(context.Background()).SendMsg(channel, createSimpleDisgordMessage(msg))
}

func (s *simpleDiscordSession) SendMessage(channel Snowflake, params *disgord.CreateMessageParams) {
	s.disgordSession.WithContext(context.Background()).SendMsg(channel, params.Content)
}

func (s *simpleDiscordSession) ReactToMessage(msg Snowflake, channel Snowflake, emoji interface{}) {
	s.disgordSession.Channel(channel).Message(msg).Reaction(emoji).WithContext(context.Background()).Create()
}

func (s *simpleDiscordSession) CurrentUser() (*disgord.User, error) {
	return s.disgordSession.CurrentUser().WithContext(context.Background()).Get()
}

func (s *simpleDiscordSession) Guild(guild Snowflake) Guild {
	return s.disgordSession.Guild(guild)
}

func createSimpleDisgordMessage(m string) *disgord.CreateMessageParams {
	return &disgord.CreateMessageParams{
		Content: m,
	}
}

type Guild interface {
	GetChannels(...disgord.Flag) ([]*disgord.Channel, error)
	GetRoles(flags ...disgord.Flag) ([]*disgord.Role, error)
	GetEmojis(flags ...disgord.Flag) ([]*disgord.Emoji, error)
	Member(userID Snowflake) disgord.GuildMemberQueryBuilder
}