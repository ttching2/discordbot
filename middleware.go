package main

import "github.com/andersfylling/disgord"

type middlewareHolder struct {
	session disgord.Session
	myself  *disgord.User
}

func (bot *discordBot) commandInUse(evt interface{}) interface{} {
	if msg, ok := evt.(*disgord.MessageCreate); ok {
		if !bot.commands.IsUserUsingCommand(msg.Message.Author) {
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
		if !bot.commands.IsRoleCommandMessage(e.MessageID, e.PartialEmoji.ID) {
			return nil
		}
	}

	return evt
}

func (m *middlewareHolder) isReactMessageCommand(evt interface{}) interface{} {
	command := "!react"
	if e, ok := evt.(*disgord.MessageCreate); ok {
		if e.Message.Content[:len(command)] != command {
			return nil
		}
	}

	stripCommand(evt, command)
	return evt
}

func (m *middlewareHolder) isTwitterFollowCommand(evt interface{}) interface{} {
	command := "!twitter-follow"
	if e, ok := evt.(*disgord.MessageCreate); ok {
		if e.Message.Content[:len(command)] != command {
			return nil
		}
	}

	stripCommand(evt, command)
	return evt
}

func (m *middlewareHolder) isTwitterFollowRemoveCommand(evt interface{}) interface{} {
	command := "!twitter-follow-remove"
	if e, ok := evt.(*disgord.MessageCreate); ok {
		if e.Message.Content[:len(command)] !=  command {
			return nil
		}
	}

	stripCommand(evt, command)
	return evt
}

func (m *middlewareHolder) isTwitterFollowListCommand(evt interface{}) interface{} {
	command := "!twitter-follow-list"
	if e, ok := evt.(*disgord.MessageCreate); ok {
		if e.Message.Content[:len(command)] != command {
			return nil
		}
	}
	stripCommand(evt, command)
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

func stripCommand(evt interface{}, command string) {
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