package commands

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/andersfylling/disgord"
)

//FindChannelByName helper function to find a channel by name
func findChannelByName(name string, channels []*disgord.Channel) *disgord.Channel {
	for _, channel := range channels {
		if strings.EqualFold(channel.Name, name) {
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
		if strings.EqualFold(role.Name, name) {
			return role
		}
	}
	return nil
}

//FindEmojiByName helper function
func FindEmojiByName(name string, emojis []*disgord.Emoji) *disgord.Emoji {
	for _, emoji := range emojis {
		if strings.EqualFold(emoji.Name, name) {
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

func FindChannelByName(channel string, g Guild) *disgord.Channel{
	channels, _ := g.GetChannels()
	return findChannelByName(channel, channels)
}

func FindTargetChannel(channel disgord.Snowflake, g Guild) *disgord.Channel {
	channels, _ := g.GetChannels()
	return findChannelByID(channel, channels)
}

func FindTargetEmoji(emoji disgord.Snowflake, g Guild) *disgord.Emoji {
	emojis, _ := g.GetEmojis()
	return findEmojiByID(emoji, emojis)
}

func createMention(s Snowflake) string {
	return "<@&" + s.String() + ">"
}

func doHttpGetRequest(link string) io.Reader {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		log.Error(err)
		return nil
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		return nil
	}
	defer r.Body.Close()

	result, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return nil
	}
	return bytes.NewReader(result)
}