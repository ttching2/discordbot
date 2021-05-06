package botcommands

import (
	"discordbot/repositories/model"

	"github.com/andersfylling/disgord"
)

const CommandPrefix = "$"

type DiscordSession interface {
	SendMessage(model.Snowflake, *disgord.CreateMessageParams)
	ReactToMessage(msg model.Snowflake, channel model.Snowflake, emoji interface{})
}

func createSimpleDisgordMessage(m string) *disgord.CreateMessageParams {
	return &disgord.CreateMessageParams{
		Content: m,
	}
}