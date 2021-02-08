package rolemessage_test

import "github.com/andersfylling/disgord"

var data = disgord.MessageCreate{
	Message: &disgord.Message{
		GuildID: 456,
		ChannelID: 123,
		Content: "Message",
		Author: &disgord.User{
			ID: 123,
		},
	},
	
}