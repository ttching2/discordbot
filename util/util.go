package util

import (
	"discordbot/botcommands/discord"
	"strings"

	"github.com/andersfylling/disgord"
)

//FindChannelByName helper function to find a channel by name
func findChannelByName(name string, channels []*disgord.Channel) *disgord.Channel {
	for _, channel := range channels {
		if strings.ToLower(channel.Name) == strings.ToLower(name) {
			return channel
		}
	}
	return nil
}

func findChannelByID(id disgord.Snowflake, channels []*disgord.Channel) *disgord.Channel {
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

func findEmojiByID(id disgord.Snowflake, emojis []*disgord.Emoji) *disgord.Emoji {
	for _, emoji := range emojis {
		if emoji.ID == id {
			return emoji
		}
	}
	return nil
}

func FindChannelByName(channel string, g discord.Guild) *disgord.Channel{
	channels, _ := g.GetChannels()
	return findChannelByName(channel, channels)
}

func FindTargetChannel(channel disgord.Snowflake, g discord.Guild) *disgord.Channel {
	channels, _ := g.GetChannels()
	return findChannelByID(channel, channels)
}

func FindTargetEmoji(emoji disgord.Snowflake, g discord.Guild) *disgord.Emoji {
	emojis, _ := g.GetEmojis()
	return findEmojiByID(emoji, emojis)
}