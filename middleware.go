package main

import (
	"context"
	"discordbot/botcommands/discord"
	"discordbot/repositories"
	"discordbot/repositories/rolecommand"
	"discordbot/repositories/users_repository"
	"encoding/json"
	"errors"
	"strings"

	"github.com/andersfylling/disgord"
)

type middlewareHolder struct {
	session               disgord.Session
	myself                *disgord.User
	roleCommandRepository *rolecommand.RoleCommandRepository
	usersRepository       *users_repository.UsersRepository
}

func newMiddlewareHolder(ctx context.Context, s disgord.Session, roleCommandRepository *rolecommand.RoleCommandRepository, usersRepository *users_repository.UsersRepository) (m *middlewareHolder, err error) {
	m = &middlewareHolder{session: s, roleCommandRepository: roleCommandRepository, usersRepository: usersRepository}
	if m.myself, err = s.CurrentUser().WithContext(ctx).Get(); err != nil {
		return nil, errors.New("unable to fetch info about the bot instance")
	}
	return m, nil
}

func (m *middlewareHolder) createMessageContentForNonCommand(evt interface{}) interface{} {
	e, ok := evt.(*disgord.MessageCreate)
	if !ok {
		return nil
	}

	user := repositories.Users{DiscordUsersID: e.Message.Author.ID}
	if !m.usersRepository.DoesUserExist(e.Message.Author.ID) {
		err := m.usersRepository.SaveUser(&user)
		if err != nil {
			log.Println(err)
			return nil
		}
	} else {
		user = m.usersRepository.GetUserByDiscordId(e.Message.Author.ID)
	}

	middleWareContent := discord.MiddleWareContent{MessageContent: e.Message.Content, UsersID: user.UsersID}
	jsonContent, err := json.Marshal(middleWareContent)
	if err != nil {
		log.Println(err)
		return nil
	}
	e.Message.Content = string(jsonContent)
	return evt
}

func (m *middlewareHolder) checkAndSaveUser(evt interface{}) interface{} {
	e, ok := evt.(*disgord.MessageCreate)
	if !ok {
		return nil
	}

	user := repositories.Users{DiscordUsersID: e.Message.Author.ID, UserName: e.Message.Author.Username}
	if !m.usersRepository.DoesUserExist(e.Message.Author.ID) {
		err := m.usersRepository.SaveUser(&user)
		if err != nil {
			log.Println(err)
			return nil
		}
	} else {
		user = m.usersRepository.GetUserByDiscordId(e.Message.Author.ID)
	}

	split := strings.Split(e.Message.Content, " ")
	var messageContent string
	if len(split) > 1 {
		messageContent = e.Message.Content[len(split[0])+1:]
	}
	middleWareContent := discord.MiddleWareContent{Command: split[0], MessageContent: messageContent, UsersID: user.UsersID}
	jsonContent, err := json.Marshal(middleWareContent)
	if err != nil {
		log.Println(err)
		return nil
	}
	e.Message.Content = string(jsonContent)
	return evt
}

func (m *middlewareHolder) isFromAdmin(evt interface{}) interface{} {
	if e, ok := evt.(*disgord.MessageCreate); ok {
		if e.Message.Author.ID != 124343682382954498 {
			return nil
		}
	}
	return evt
}

func (m *middlewareHolder) commandInUse(evt interface{}) interface{} {
	if msg, ok := evt.(*disgord.MessageCreate); ok {
		if inUse, err := m.roleCommandRepository.IsUserUsingCommand(msg.Message.Author.ID, msg.Message.ChannelID); err != nil || !inUse {
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

func (m *middlewareHolder) reactionMessage(evt interface{}) interface{} {
	if e, ok := evt.(*disgord.MessageReactionAdd); ok {
		if isCommand, err := m.roleCommandRepository.IsRoleCommandMessage(e.MessageID, e.PartialEmoji.ID); err != nil || !isCommand {
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
