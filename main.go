package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"discordbot/botcommands"
	db "discordbot/databaseclient"
	"discordbot/strawpoll"
	myTwitter "discordbot/twitter"
	"discordbot/util"

	"github.com/dghubble/go-twitter/twitter"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/disgord/std"
	"github.com/sirupsen/logrus"
)

var log = &logrus.Logger{
	Out:       os.Stderr,
	Formatter: new(logrus.TextFormatter),
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.DebugLevel,
}

type discordConfig struct {
	botToken string
}

type BotConfig struct {
	DiscordConfig discordConfig
	TwitterConfig myTwitter.TwitterClientConfig
	StrawPollConfig strawpoll.StrawPollConfig
}

type discordBot struct {
	saveableCommand botcommands.SaveableCommand
	twitterClient *myTwitter.TwitterClient
	strawpollClient *strawpoll.Client
	commands map[string]botcommands.Command
}

func main() {
	botConfig := BotConfig{
		DiscordConfig: discordConfig{
			botToken: os.Getenv("DISCORD_TOKEN"),
		},
		TwitterConfig: myTwitter.TwitterClientConfig{
			ConsumerKey: os.Getenv("TWITTER_API_KEY"),
			ConsumerSecret: os.Getenv("TWITTER_SECRET_KEY"),
			AccessToken: os.Getenv("TWITTER_ACCESS_TOKEN"),
			AccessSecret: os.Getenv("TWITTER_TOKEN_SECRET"),
		},
		StrawPollConfig: strawpoll.StrawPollConfig{
			ApiKey: os.Getenv("STRAWPOLL_TOKEN"),
		},
	}
	client := disgord.New(disgord.Config{
		BotToken: botConfig.DiscordConfig.botToken,
		Logger:   log, // optional logging
		Cache:    &disgord.CacheNop{},
	})
	
	bot := initializeBot(client, botConfig)
	run(client, bot)
}

func initializeBot(client *disgord.Client, config BotConfig) discordBot {
	twitterClient := myTwitter.NewClient(config.TwitterConfig)
	dbClient := db.NewClient()

	setupTwitterClient(client, dbClient, twitterClient)
	strawpollClient := strawpoll.New(config.StrawPollConfig)

	restartStrawpollDeadlines(client, dbClient, strawpollClient)

	commands := make(map[string]botcommands.Command)
	commands[botcommands.RoleReactString] = botcommands.NewRoleReactCommand()
	commands[botcommands.TwitterFollowString] = botcommands.NewTwitterFollowCommand(twitterClient)
	commands[botcommands.TwitterFollowListString] = botcommands.NewTwitterFollowListCommand()
	commands[botcommands.TwitterUnfollowString] = botcommands.NewTwitterUnfollowCommand(twitterClient)
	commands[botcommands.StrawPollDeadlineString] = botcommands.NewStrawPollCommand(strawpollClient)

	commands[botcommands.HelpString] = botcommands.NewHelpCommand(commands)

	return discordBot{
		saveableCommand: dbClient,
		twitterClient: twitterClient,
		commands: commands,
		strawpollClient: strawpollClient,
	}
}

func run(client *disgord.Client, bot discordBot) {

	content, _ := std.NewMsgFilter(context.Background(), client)
	customMiddleWare, _ := newMiddlewareHolder(context.Background(), client)
	content.SetPrefix(botcommands.CommandPrefix)

	// listen for messages
	client.Gateway().
		WithMiddleware(customMiddleWare.filterBotMsg, bot.commandInUse).
		MessageCreate(bot.reactRoleMessage)
	client.Gateway().
		WithMiddleware(customMiddleWare.filterBotMsg, content.StripPrefix, bot.isFromAdmin, bot.isBotCommand).
		MessageCreate(bot.ExecuteCommand)
	client.Gateway().
		WithMiddleware(bot.reactionMessage).
		MessageDelete(bot.removeReactRoleMessage)
	client.Gateway().
		WithMiddleware(customMiddleWare.filterOutBots, bot.reactionMessage).
		MessageReactionAdd(bot.addRole)
	client.Gateway().
		WithMiddleware(customMiddleWare.filterOutBots, bot.reactionMessage).
		MessageReactionRemove(bot.removeRole)

	// connect now, and disconnect on system interrupt
	client.Gateway().StayConnectedUntilInterrupted()
}

func restartStrawpollDeadlines(client *disgord.Client, dbClient botcommands.StrawpollDeadlineRepository, strawpollClient *strawpoll.Client) {
	for _, strawpoll := range dbClient.GetAllStrawpollDeadlines() {

		poll, err := strawpollClient.GetPoll(strawpoll.StrawpollID)
		if err != nil {
			dbClient.DeleteStrawpollDeadlineByID(strawpoll.StrawpollDeadlineID)
			continue
		}

		now := time.Now()
		if now.After(poll.Content.Deadline) {
			dbClient.DeleteStrawpollDeadlineByID(strawpoll.StrawpollDeadlineID)
			continue
		}

		timeToWait := time.NewTimer(poll.Content.Deadline.Sub(now))
		go func(strawpoll botcommands.StrawpollDeadline) {
			<-timeToWait.C
			poll, _ := strawpollClient.GetPoll(strawpoll.StrawpollID)
			pollAnswers := poll.Content.Poll.PollAnswers
			topAnswer := pollAnswers[0]
			for _, answer := range pollAnswers {
				if answer.Votes > topAnswer.Votes {
					topAnswer = answer
				}
			}
			result := fmt.Sprintf("<@&%s> Strawpoll has closed. The top vote for %s is %s with %d votes.", strawpoll.Role, poll.Content.Title, topAnswer.Answer, topAnswer.Votes)
			client.Channel(strawpoll.Channel).CreateMessage(&disgord.CreateMessageParams{Content: result})
			dbClient.DeleteStrawpollDeadlineByID(strawpoll.StrawpollDeadlineID)
		}(*strawpoll)
	}
}

func setupTwitterClient(client *disgord.Client, dbClient botcommands.SavedTwitterFollowCommand, twitterClient *myTwitter.TwitterClient) {
	tweetHandler := func(tweet *twitter.Tweet) {
		discordMessage := fmt.Sprintf("New Tweet by %s https://www.twitter.com/%s/status/%s", tweet.User.Name, tweet.User.ScreenName, tweet.IDStr)

		newMessageParams := &disgord.CreateMessageParams {
			Content: discordMessage,
		}

		twitterFollowCommands := dbClient.GetFollowedUser(tweet.User.ScreenName)

		if twitterFollowCommands == nil {
			return
		}

		for i := range twitterFollowCommands {
			client.Channel(twitterFollowCommands[i].Channel).CreateMessage(newMessageParams)
		}
		
	}
	twitterClient.SetTweetDemux(tweetHandler)

	var followedUsers []string
	for _, followed := range dbClient.GetAllUniqueFollowedUsers() {
		followedUsers = append(followedUsers, followed.ScreenNameID)
	}
	twitterClient.AddUsersToTrack(followedUsers)
}

func (bot *discordBot) isFromAdmin(evt interface{}) interface{} {
	if e, ok := evt.(*disgord.MessageCreate); ok {
		if e.Message.Author.ID != 124343682382954498 {
			return nil
		}
	}
	return evt
}

func (bot *discordBot) isBotCommand(evt interface{}) interface{} {
	if e, ok := evt.(*disgord.MessageCreate); ok {
		splitContent := strings.Split(e.Message.Content, " ")
		if _, ok := bot.commands[splitContent[0]]; !ok {
			return nil
		}
	}
	return evt
}

func (bot *discordBot) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate) {
	msg := data.Message
	command := strings.Split(msg.Content, " ")
	//TODO could be done better :/
	bot.commands[command[0]].ExecuteCommand(s, data, bot.saveableCommand)
}

func (bot *discordBot) removeReactRoleMessage(s disgord.Session, data *disgord.MessageDelete) {
	bot.saveableCommand.RemoveRoleReactCommand(data.MessageID)
}

func (bot *discordBot) reactRoleMessage(s disgord.Session, data *disgord.MessageCreate) {
	msg := data.Message

	commandInProgress := bot.saveableCommand.GetCommandInProgress(msg.Author)
	switch commandInProgress.Stage {
	case 1:
		//stage 1 ask for channel
		channels, _ := s.Guild(msg.GuildID).GetChannels()
		channel := util.FindChannelByName(msg.Content, channels)
		if channel == nil {
			msg.Reply(context.Background(), s, "Channel not found. Aborting command.")
			bot.saveableCommand.RemoveCommandProgress(msg.Author.ID)
			return 
		}
		commandInProgress.Channel = channel.ID
		msg.Reply(context.Background(), s, "Enter role to be assigned")
		commandInProgress.Stage = 2
		bot.saveableCommand.SaveCommandInProgress(msg.Author, *commandInProgress)
	case 2:
		//stage 2 ask for role
		roles, _ := s.Guild(msg.GuildID).GetRoles()
		role := util.FindRoleByName(msg.Content, roles)
		if role == nil {
			msg.Reply(context.Background(), s, "Role not found. Aborting command.")
			bot.saveableCommand.RemoveCommandProgress(msg.Author.ID)
			return 
		}
		commandInProgress.Role = role.ID
		msg.Reply(context.Background(), s, "Enter reaction to use.")
		commandInProgress.Stage = 3
		bot.saveableCommand.SaveCommandInProgress(msg.Author, *commandInProgress)
		
	case 3:
		//stage 3 ask for reaction
		emojis, _ := s.Guild(msg.GuildID).GetEmojis()
		//TODO if it uses a unicode emoji we get a panik
		emojiName := strings.Split(msg.Content, ":")
		emoji := util.FindEmojiByName(emojiName[1], emojis)
		if emoji == nil {
			msg.Reply(context.Background(), s, "Reaction not found. Aborting command.")
			bot.saveableCommand.RemoveCommandProgress(msg.Author.ID)
			return 
		}
		commandInProgress.Emoji = emoji.ID
		msg.Reply(context.Background(), s, "Enter message to use")
		commandInProgress.Stage = 4
		bot.saveableCommand.SaveCommandInProgress(msg.Author, *commandInProgress)
		
	case 4:
		//stage 4 ask for message
		channels, _ := s.Guild(msg.GuildID).GetChannels()
		commandChannel := util.FindChannelByID(commandInProgress.Channel, channels)

		botMsg, _ := commandChannel.SendMsg(context.Background(), s, msg)

		emojis, _ := s.Guild(msg.GuildID).GetEmojis()
		emoji := util.FindEmojiByID(commandInProgress.Emoji, emojis)
		botMsg.React(context.Background(), s, emoji)

		roleCommand := botcommands.CompletedRoleCommand{
			Guild: commandInProgress.Guild,
			Role: commandInProgress.Role,
			Emoji: commandInProgress.Emoji}
		bot.saveableCommand.SaveRoleCommand(botMsg.ID, roleCommand)
		bot.saveableCommand.RemoveCommandProgress(msg.Author.ID)
	default:
		//error
	}
}

//Bot role needs to be above role to give the role.
func (bot *discordBot) addRole(s disgord.Session, data *disgord.MessageReactionAdd) {
	userID := data.UserID
	command := bot.saveableCommand.GetRoleCommand(data.MessageID)
	s.Guild(command.Guild).Member(userID).AddRole(command.Role)
}

func (bot *discordBot) removeRole(s disgord.Session, data *disgord.MessageReactionRemove) {
	userID := data.UserID
	command := bot.saveableCommand.GetRoleCommand(data.MessageID)
	s.Guild(command.Guild).Member(userID).RemoveRole(command.Role)
}

func newMiddlewareHolder(ctx context.Context, s disgord.Session) (m *middlewareHolder, err error) {
	m = &middlewareHolder{session: s}
	if m.myself, err = s.CurrentUser().WithContext(ctx).Get(); err != nil {
		return nil, errors.New("unable to fetch info about the bot instance")
	}
	return m, nil
}