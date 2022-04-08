package commands

import (
	"context"

	"github.com/andersfylling/disgord"
)

const discordEmojiCDN = "https://cdn.discordapp.com/emojis/"

type DiscordSession interface {
	SendMessage(Snowflake, *disgord.CreateMessageParams) (*disgord.Message, error)
	SendSimpleMessage(Snowflake, string) (*disgord.Message, error)
	ReactToMessage(msg Snowflake, channel Snowflake, emoji interface{})
	ReactWithThumbsDown(*disgord.Message)
	ReactWithThumbsUp(*disgord.Message)
	CurrentUser() (*disgord.User, error)
	Guild(Snowflake) Guild
	Channel(Snowflake) Channel
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

func (s *simpleDiscordSession) SendMessage(channel Snowflake, params *disgord.CreateMessageParams) (*disgord.Message, error) {
	return s.disgordSession.WithContext(context.Background()).SendMsg(channel, params)
}

func (s *simpleDiscordSession) ReactToMessage(msg Snowflake, channel Snowflake, emoji interface{}) {
	s.disgordSession.Channel(channel).Message(msg).Reaction(emoji).WithContext(context.Background()).Create()
}

func (s *simpleDiscordSession) ReactWithThumbsDown(msg *disgord.Message) {
	s.ReactToMessage(msg.ID, msg.ChannelID, "👎")
}

func (s *simpleDiscordSession)ReactWithThumbsUp(msg *disgord.Message) {
	s.ReactToMessage(msg.ID, msg.ChannelID, "👍")
}

func (s *simpleDiscordSession) CurrentUser() (*disgord.User, error) {
	return s.disgordSession.CurrentUser().WithContext(context.Background()).Get()
}

func (s *simpleDiscordSession) Guild(guild Snowflake) Guild {
	return s.disgordSession.Guild(guild)
}

func (s *simpleDiscordSession) Channel(channel Snowflake) Channel {
	return s.disgordSession.Channel(channel)
}

func createSimpleDisgordMessage(m string) *disgord.CreateMessageParams {
	return &disgord.CreateMessageParams{
		Content: m,
	}
}

type Guild interface {
	GetChannels() ([]*disgord.Channel, error)
	GetRoles() ([]*disgord.Role, error)
	GetEmojis() ([]*disgord.Emoji, error)
	Member(userID Snowflake) disgord.GuildMemberQueryBuilder
}

type Channel interface {
	GetMessages(params *disgord.GetMessagesParams) ([]*disgord.Message, error)
	DeleteMessages(params *disgord.DeleteMessagesParams) error
	Message(id Snowflake) disgord.MessageQueryBuilder
}