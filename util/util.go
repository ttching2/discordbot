package util

import (
	"strings"

	"github.com/andersfylling/disgord"
)

//FindChannelByName helper function to find a channel by name
func FindChannelByName(name string, channels []*disgord.Channel) *disgord.Channel {
	for _, channel := range channels {
		if strings.ToLower(channel.Name) == strings.ToLower(name) {
			return channel
		}
	}
	return nil
}

func FindChannelByID(id disgord.Snowflake, channels []*disgord.Channel) *disgord.Channel {
	for _, channel := range channels {
		if channel.ID == id {
			return channel
		}
	}
	return nil
}

//FindRoleByName helper function
func FindRoleByName(name string , roles []*disgord.Role) *disgord.Role {
	for _, role := range roles {
		if strings.ToLower(role.Name) == strings.ToLower(name) {
			return role
		}
	}
	return nil
}

//FindEmojiByName helper function
func FindEmojiByName(name string, emojis []*disgord.Emoji) *disgord.Emoji {
	for _, emoji := range emojis {
		if strings.ToLower(emoji.Name) == strings.ToLower(name) {
			return emoji
		}
	}
	return nil
}

func FindEmojiByID(id disgord.Snowflake, emojis []*disgord.Emoji) *disgord.Emoji {
	for _, emoji := range emojis {
		if emoji.ID == id {
			return emoji
		}
	}
	return nil
}