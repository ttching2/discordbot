package commands_test

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

var emojiList = []*disgord.Emoji{
	{Name: "gamer_ready", ID: 124534},
	{Name: "Shuba", ID: 592304},
	{Name: "long_long_name_long", ID: 34209},
	{Name: "angry_face", ID: 283994089},
	{Name: "happy", ID: 29034809},
}



func testCreateRoleCommand(t *testing.T) {
	
}