package main

import (
	"discordbot/challonge"
	"discordbot/commands"
	"discordbot/strawpoll"
	"discordbot/twitter"
	"errors"
	"strconv"
	"strings"

	"github.com/andersfylling/disgord"
)

type middlewareHolder struct {
	session        commands.DiscordSession
	commandFactory map[string]func(data *disgord.MessageCreate, user *commands.Users)interface{}
	myself         *disgord.User
	*jobQueue
	*repositoryContainer
}

type messageCreateRequestFactory interface {
	CreateRequest(*disgord.MessageCreate, *commands.Users) interface{}
}

type middlewareChallongeClient struct {
	client *challonge.Client
}

func (c *middlewareChallongeClient) GetParticipants(tourneyID string) []challonge.Participant {
	pc := c.client.Participant.Index(tourneyID)
	var r []challonge.Participant
	for _, p := range pc {
		r = append(r, p.Participant)
	}
	return r
}
func (c *middlewareChallongeClient) GetMatches(tourneyID string) []challonge.Match {
	mc := c.client.Match.Index(tourneyID)
	var r []challonge.Match
	for _, m := range mc {
		r = append(r, m.Match)
	}
	return r
}
func (c *middlewareChallongeClient) GetMatch(tourneyID string, matchID int) challonge.Match {
	mc := c.client.Match.Show(tourneyID, strconv.Itoa(matchID))
	return mc.Match
}
func (c *middlewareChallongeClient) UpdateMatch(tourneyID string, matchID int, params challonge.MatchQueryParams) {
	c.client.Match.Update(tourneyID, strconv.Itoa(matchID), params)
}

func newMiddlewareHolder(discordSession commands.DiscordSession,
	jobQueue *jobQueue,
	repos *repositoryContainer,
	twitterClient *twitter.TwitterClient,
	strawpollClient *strawpoll.Client,
	challongeeClient *challonge.Client) (m *middlewareHolder, err error) {

	cclient := &middlewareChallongeClient{challongeeClient}

	roleCommandFactory := commands.NewRoleCommandRequestFactory(discordSession, repos.roleCommandRepo)
	twitterCommandFactory := commands.NewTwitterFollowCommandFactory(discordSession, twitterClient, repos.twitterFollowRepo)
	strawpollFactory := commands.NewCommandFactory(discordSession, strawpollClient, repos.strawpollRepo)
	tourneyFactory := commands.NewTourneyCommandRequestFactory(discordSession, repos.tournamentRepo, cclient)

	commandMap := make(map[string]func(data *disgord.MessageCreate, user *commands.Users)interface{})
	
	commandMap[commands.RoleReactString] = roleCommandFactory.CreateRequest
	commandMap[commands.TwitterFollowString] = twitterCommandFactory.CreateFollowCommand
	commandMap[commands.TwitterFollowListString] = twitterCommandFactory.CreateFollowListRequest
	commandMap[commands.TwitterUnfollowString] = twitterCommandFactory.CreateUnfollowRequest
	commandMap[commands.StrawPollDeadlineString] = strawpollFactory.CreateRequest
	commandMap[commands.TournamentCommandString] = tourneyFactory.CreateRequest
	commandMap[commands.TournamentAddOrganizerString] = tourneyFactory.CreateAddOrganizerCommand
	commandMap[commands.TournamentNextLosersMatchString] = tourneyFactory.CreateNextLosersCommnad
	commandMap[commands.TournamentMatchWinString] = tourneyFactory.CreateWinnerCommand
	commandMap[commands.TournamentFinishString] = tourneyFactory.CreateTourneyCloseCommand

	// var commandList []help.PrintHelp
	// for _, c := range commands {
	// 	commandList = append(commandList, c.(help.PrintHelp))
	// }
	// commands[help.HelpString] = help.NewCommandFactory(client, commandList[:])

	m = &middlewareHolder{
		session:             discordSession,
		jobQueue:            jobQueue,
		commandFactory:      commandMap,
		repositoryContainer: repos}

	if m.myself, err = discordSession.CurrentUser(); err != nil {
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

	user := commands.Users{DiscordUsersID: e.Message.Author.ID, UserName: e.Message.Author.Username}
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

	createCommand, ok := m.commandFactory[split[0]]
	if !ok {
		return nil
	}

	c := createCommand(e, &user)
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
	return commands.NewRemoveRoleMessage(m.roleCommandRepo, e)
}

func (m *middlewareHolder) createMessageContentForNonCommand(evt interface{}) interface{} {
	e, ok := evt.(*disgord.MessageCreate)
	if !ok {
		return nil
	}

	user := commands.Users{DiscordUsersID: e.Message.Author.ID}
	if !m.usersRepo.DoesUserExist(e.Message.Author.ID) {
		err := m.usersRepo.SaveUser(&user)
		if err != nil {
			log.Println(err)
			return nil
		}
	} else {
		user, _ = m.usersRepo.GetUserByDiscordId(e.Message.Author.ID)
	}

	m.jobQueue.onMessageCreate.PushBack(commands.NewInProgressRoleCommand(m.session, m.roleCommandRepo, e, &user))
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
	return commands.NewAddRoleReact(m.roleCommandRepo, m.session, e)
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
	return commands.NewRemoveRoleReact(m.roleCommandRepo, m.session, e)
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
