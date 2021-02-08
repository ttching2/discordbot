package util

import (
	"testing"

	"github.com/andersfylling/disgord"
)


var channelList = []*disgord.Channel{
	{Name: "Test", ID: 12345},
	{Name: "general", ID: 45432},
	{Name: "general-chat", ID: 1253690},
	{Name: "Gaming", ID: 18092348},
	{Name: "TwitterFeed", ID: 1},
	{Name: "👍", ID: 7983564723},
}

var roleList = []*disgord.Role{
	{Name: "Lord God"},
	{Name: "Test User"},
	{Name: "role"},
	{Name: "normal-users"},
	{Name: "👍"},
}

var emojiList = []*disgord.Emoji{
	{Name: "gamer_ready", ID: 124534},
	{Name: "Shuba", ID: 592304},
	{Name: "long_long_name_long", ID: 34209},
	{Name: "angry_face", ID: 283994089},
	{Name: "happy", ID: 29034809},
}

type stringTestPair struct {
	given string
	expected *disgord.Channel
}

type intTestPair struct {
	given disgord.Snowflake
	expected *disgord.Channel
}

func TestFindChannelByName(t *testing.T) {
	namesTofind := []stringTestPair{
		{"test", &disgord.Channel{Name: "Test", ID: 12345}},
		{"suh", nil},
		{"General", &disgord.Channel{Name: "general", ID: 45432}},
		{"👍", &disgord.Channel{Name: "👍", ID: 7983564723}},
	}
	for _, pair := range namesTofind {
		r := FindChannelByName(pair.given, channelList)
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

func TestFindChannelByID(t *testing.T) {
	idsToFind := []intTestPair{
		{12345, &disgord.Channel{Name: "Test", ID: 12345}},
		{5, nil},
		{45432, &disgord.Channel{Name: "general", ID: 45432}},
		{7983564723, &disgord.Channel{Name: "👍", ID: 7983564723}},
	}
	for _, pair := range idsToFind {
		r := FindChannelByID(pair.given, channelList)
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
		r := FindRoleByName(pair.given, roleList)
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
		r := FindEmojiByName(pair.given, emojiList)
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


func TestFindEmojiByID(t *testing.T) {
	idsToFind := []intEmojiTestPair{
		{124534, &disgord.Emoji{Name: "gamer_ready", ID: 124534}},
		{5, nil},
		{592304, &disgord.Emoji{Name: "Shuba", ID: 592304}},
		{34209, &disgord.Emoji{Name: "long_long_name_long", ID: 34209}},
	}
	for _, pair := range idsToFind {
		r := FindEmojiByID(pair.given, emojiList)
		if pair.expected == nil {
			if r != nil {
				t.Error(
					"For string", pair.given,
					"Expected", pair.expected,
					"Got", r,
				)
			}
		} else {
			if r.ID != pair.expected.ID {
				t.Error(
					"For string", pair.given,
					"Expected", pair.expected,
					"Got", r,
				)
			}
		}
	}
}
