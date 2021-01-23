package main

import "github.com/andersfylling/disgord"

type middlewareHolder struct {
	session disgord.Session
	myself  *disgord.User
}

func (bot *discordBot) commandInUse(evt interface{}) interface{} {
	if msg, ok := evt.(*disgord.MessageCreate); ok {
		if !bot.saveableCommand.IsUserUsingCommand(msg.Message.Author) {
			return nil
		}
	}

	return evt
}

func (m *middlewareHolder) filterBotMsg(evt interface{}) interface{} {
	if e, ok := evt.(*disgord.MessageCreate); ok {
		if e.Message.Author.ID == m.myself.ID {
			return nil
		}
	}

	return evt
}

func (bot *discordBot) reactionMessage(evt interface{}) interface{} {
	if e, ok := evt.(*disgord.MessageReactionAdd); ok {
		if !bot.saveableCommand.IsRoleCommandMessage(e.MessageID, e.PartialEmoji.ID) {
			return nil
		}
	}

	return evt
}

func (m *middlewareHolder) filterOutBots(evt interface{}) interface{} {
	if e, ok := evt.(*disgord.MessageReactionAdd); ok {
		if e.UserID == m.myself.ID {
			return nil
		}
	}

	return evt
}

func StripCommand(evt interface{}, command string) {
	msg := getMsg(evt)
	msg.Content = msg.Content[len(command):]
}

func getMsg(evt interface{}) (msg *disgord.Message) {
	switch t := evt.(type) {
	case *disgord.MessageCreate:
		msg = t.Message
	case *disgord.MessageUpdate:
		msg = t.Message
	default:
		msg = nil
	}

	return msg
}