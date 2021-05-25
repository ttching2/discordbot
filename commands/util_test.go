package commands_test

import (
	"discordbot/commands"
	"testing"

	"github.com/andersfylling/disgord"
)

var roleList = []*disgord.Role{
	{Name: "Lord God"},
	{Name: "Test User"},
	{Name: "role"},
	{Name: "normal-users"},
	{Name: "👍"},
}


type stringTestPair struct {
	given string
	expected *disgord.Channel
}

type intTestPair struct {
	given disgord.Snowflake
	expected *disgord.Channel
}

type stringRoleTestPair struct {
	given string
	expected *disgord.Role
}

func TestFindRoleByName(t *testing.T) {
	namesTofind := []stringRoleTestPair{
		{"lord god", &disgord.Role{Name: "Lord God"}},
		{"Test user", &disgord.Role{Name: "Test User"}},
		{"rhe", nil},
		{"👍", &disgord.Role{Name: "👍"}},
	}
	for _, pair := range namesTofind {
		r := commands.FindRoleByName(pair.given, roleList)
		if pair.expected == nil {
			if r != nil {
				t.Error(
					"For string", pair.given,
					"Expected", pair.expected,
					"Got", r,
				)
			}
		} else {
			if r.Name != pair.expected.Name {
				t.Error(
					"For string", pair.given,
					"Expected", pair.expected,
					"Got", r,
				)
			}
		}
	}
}

type stringEmojiTestPair struct {
	given string
	expected *disgord.Emoji
}

type intEmojiTestPair struct {
	given disgord.Snowflake
	expected *disgord.Emoji
}

func TestFindEmojiByName(t *testing.T) {
	idsToFind := []stringEmojiTestPair{
		{"gamer_ready", &disgord.Emoji{Name: "gamer_ready", ID: 124534}},
		{"not exist", nil},
		{"shuba", &disgord.Emoji{Name: "Shuba", ID: 592304}},
		{"long_long_name_long", &disgord.Emoji{Name: "long_long_name_long", ID: 34209}},
	}
	for _, pair := range idsToFind {
		r := commands.FindEmojiByName(pair.given, emojiList)
		if pair.expected == nil {
			if r != nil {
				t.Error(
					"For string", pair.given,
					"Expected", pair.expected,
					"Got", r,
				)
			}
		} else {
			if r.Name != pair.expected.Name {
				t.Error(
					"For string", pair.given,
					"Expected", pair.expected,
					"Got", r,
				)
			}
		}
	}
}
