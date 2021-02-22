package main

import (
	"context"
	"discordbot/botcommands/discord"
	"discordbot/botcommands/help"
	"discordbot/botcommands/rolemessage"
	"discordbot/botcommands/strawpolldeadline"
	"discordbot/botcommands/twittercommands"
	"discordbot/repositories/model"
	"discordbot/strawpoll"
	"discordbot/twitter"
	"encoding/json"
	"errors"
	"strings"

	"github.com/andersfylling/disgord"
)

type middlewareHolder struct {
	session        disgord.Session
	commandFactory map[string]messageCreateRequestFactory
	myself         *disgord.User
	*jobQueue
	*repositoryContainer
}

type messageCreateRequestFactory interface {
	CreateRequest(*disgord.MessageCreate, *model.Users) interface{}
}

func newMiddlewareHolder(client disgord.Session, jobQueue *jobQueue, repos *repositoryContainer, twitterClient *twitter.TwitterClient, strawpollClient *strawpoll.Client) (m *middlewareHolder, err error) {
	commands := make(map[string]messageCreateRequestFactory)
	commands[rolemessage.RoleReactString] = rolemessage.NewRoleCommandRequestFactory(client, repos.roleCommandRepo)
	commands[twittercommands.TwitterFollowString] = twittercommands.NewTwitterFollowCommandFactory(client, twitterClient, repos.twitterFollowRepo)
	commands[twittercommands.TwitterFollowListString] = twittercommands.NewTwitterFollowListCommandFactory(client, repos.twitterFollowRepo)
	commands[twittercommands.TwitterUnfollowString] = twittercommands.NewTwitterUnfollowCommandFactory(client, twitterClient, repos.twitterFollowRepo)
	commands[strawpolldeadline.StrawPollDeadlineString] = strawpolldeadline.NewCommandFactory(client, strawpollClient, repos.strawpollRepo)

	var commandList []help.PrintHelp
	for _, c := range commands {
		commandList = append(commandList, c.(help.PrintHelp))
	}
	commands[help.HelpString] = help.NewCommandFactory(client, commandList[:])

	m = &middlewareHolder{
		session:             client,
		jobQueue:            jobQueue,
		commandFactory:      commands,
		repositoryContainer: repos}

	if m.myself, err = client.CurrentUser().WithContext(context.Background()).Get(); err != nil {
		return nil, errors.New("unable to fetch info about the bot instance")
	}
	return m, nil
}

func (m *middlewareHolder) handleDiscordEvent(evt interface{}) interface{} {
	switch eventType := evt.(type) {
	case *disgord.MessageCreate:
		return m.createOnMessageCommand(eventType)
	case *disgord.MessageDelete:
		return m.onMessageDelete(eventType)
	case *disgord.MessageReactionAdd:
		return m.reactionAdd(eventType)
	case *disgord.MessageReactionRemove:
		return m.reactionRemove(eventType)
	default:
		return nil
	}
}

func (m *middlewareHolder) createOnMessageCommand(e *disgord.MessageCreate) interface{} {

	user := model.Users{DiscordUsersID: e.Message.Author.ID, UserName: e.Message.Author.Username}
	if !m.usersRepo.DoesUserExist(e.Message.Author.ID) {
		err := m.usersRepo.SaveUser(&user)
		if err != nil {
			log.Println(err)
			return nil
		}
	} else {
		user, _ = m.usersRepo.GetUserByDiscordId(e.Message.Author.ID)
	}
	split := strings.Split(e.Message.Content, " ")
	var messageContent string
	if len(split) > 1 {
		messageContent = e.Message.Content[len(split[0])+1:]
	}

	f, ok := m.commandFactory[split[0]]
	if !ok {
		return nil
	}

	c := f.CreateRequest(e, &user)
	m.jobQueue.onMessageCreate.PushBack(c)
	e.Message.Content = messageContent
	return e
}

func (m *middlewareHolder) onMessageDelete(e *disgord.MessageDelete) interface{} {
	c := m.createOnMessageDeleteAction(e)
	m.jobQueue.onMessageDelete.PushBack(c)
	return e
}

func (m *middlewareHolder) createOnMessageDeleteAction(e *disgord.MessageDelete) onMessageDelete {
	return &rolemessage.RemoveRoleMessage{Repo: m.roleCommandRepo, Data: e}
}

func (m *middlewareHolder) createMessageContentForNonCommand(evt interface{}) interface{} {
	e, ok := evt.(*disgord.MessageCreate)
	if !ok {
		return nil
	}

	user := model.Users{DiscordUsersID: e.Message.Author.ID}
	if !m.usersRepo.DoesUserExist(e.Message.Author.ID) {
		err := m.usersRepo.SaveUser(&user)
		if err != nil {
			log.Println(err)
			return nil
		}
	} else {
		user, _ = m.usersRepo.GetUserByDiscordId(e.Message.Author.ID)
	}

	m.jobQueue.onMessageCreate.PushBack(rolemessage.InProgressRoleCommand{
		S:      m.session,
		Repo:   m.roleCommandRepo,
		Data:   e,
		UserID: &user,
	})
	return evt
}

func (m *middlewareHolder) reactionAdd(e *disgord.MessageReactionAdd) interface{} {
	if isCommand, err := m.roleCommandRepo.IsRoleCommandMessage(e.MessageID, e.PartialEmoji.ID); err != nil || !isCommand {
		return nil
	}

	c := m.createReactionAddAction(e)
	m.jobQueue.onReactionAdd.PushBack(c)

	return e
}

func (m *middlewareHolder) createReactionAddAction(e *disgord.MessageReactionAdd) onReactionAdd {
	return &rolemessage.AddRoleReact{
		Repo:    m.roleCommandRepo,
		Session: m.session,
		Data:    e,
	}
}

func (m *middlewareHolder) reactionRemove(e *disgord.MessageReactionRemove) interface{} {
	if isCommand, err := m.roleCommandRepo.IsRoleCommandMessage(e.MessageID, e.PartialEmoji.ID); err != nil || !isCommand {
		return nil
	}

	c := m.createReactionRemoveAction(e)
	m.jobQueue.onReactionRemove.PushBack(c)

	return e
}

func (m *middlewareHolder) createReactionRemoveAction(e *disgord.MessageReactionRemove) onReactionRemove {
	return &rolemessage.RemoveRoleReact{
		Repo:    m.roleCommandRepo,
		Session: m.session,
		Data:    e,
	}
}

func (m *middlewareHolder) checkAndSaveUser(evt interface{}) interface{} {
	e, ok := evt.(*disgord.MessageCreate)
	if !ok {
		return nil
	}

	user := model.Users{DiscordUsersID: e.Message.Author.ID, UserName: e.Message.Author.Username}
	if !m.usersRepo.DoesUserExist(e.Message.Author.ID) {
		err := m.usersRepo.SaveUser(&user)
		if err != nil {
			log.Println(err)
			return nil
		}
	} else {
		user, _ = m.usersRepo.GetUserByDiscordId(e.Message.Author.ID)
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
		if inUse, err := m.roleCommandRepo.IsUserUsingCommand(msg.Message.Author.ID, msg.Message.ChannelID); err != nil || !inUse {
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
