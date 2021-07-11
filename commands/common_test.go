package commands_test

import (
	"discordbot/commands"

	"github.com/andersfylling/disgord"
)

type onMessageCreateCommand interface {
	ExecuteMessageCreateCommand()
}

type mockSession struct {
	message          string
	reactedMessageID commands.Snowflake
	guild            commands.Guild
}

func (s *mockSession) SendSimpleMessage(channel commands.Snowflake, m string) (*disgord.Message, error) {
	s.message = m
	return nil, nil
}

func (s *mockSession) ReactToMessage(msg commands.Snowflake, channel commands.Snowflake, emoji interface{}) {
	s.reactedMessageID = msg
}

func (s *mockSession) getReactedMessage() commands.Snowflake {
	return s.reactedMessageID
}

func (s *mockSession) CurrentUser() (*disgord.User, error)        { return nil, nil }
func (s *mockSession) Guild(id commands.Snowflake) commands.Guild { return s.guild }
func (s *mockSession) ReactWithThumbsDown(*disgord.Message) {}
func (s *mockSession) ReactWithThumbsUp(*disgord.Message) {}

type mockGuild struct {
	channels []*disgord.Channel
	roles    []*disgord.Role
	emojis   []*disgord.Emoji
}

var commonMockGuild = mockGuild{
	channels: []*disgord.Channel{
		{Name: "channel"},
		{Name: "mock-channel"},
		{Name: "big-fat-channel"},
		{Name: "hehe"},
		{Name: "👍"},
	},
	roles: []*disgord.Role{
		{Name: "Lord God"},
		{Name: "Test User"},
		{Name: "role"},
		{Name: "normal-users"},
		{Name: "👍"},
	},
	emojis: []*disgord.Emoji{
		{Name: "emoji"},
		{Name: "test_emoji"},
		{Name: "hypergatcha11"},
		{Name: "super_emoji_racer"},
		{Name: "👍"},
	},
}

func (g *mockGuild) GetChannels(...disgord.Flag) ([]*disgord.Channel, error) {
	return g.channels, nil
}

func (g *mockGuild) GetRoles(flags ...disgord.Flag) ([]*disgord.Role, error) {
	return g.roles, nil
}

func (g *mockGuild) GetEmojis(flags ...disgord.Flag) ([]*disgord.Emoji, error) {
	return g.emojis, nil
}

func (g *mockGuild) Member(userID commands.Snowflake) disgord.GuildMemberQueryBuilder {
	return nil
}
