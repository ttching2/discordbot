package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"discordbot/botcommands"
	db "discordbot/databaseclient"
	myTwitter "discordbot/twitter"

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
}

type discordBot struct {
	commands botcommands.Commands
	twitterClient *myTwitter.TwitterClient
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
	}
	client := disgord.New(disgord.Config{
		BotToken: botConfig.DiscordConfig.botToken,
		Logger:   log, // optional logging
		Cache:    &disgord.CacheNop{},
	})

	twitterClient := myTwitter.NewClient(botConfig.TwitterConfig)
	
	bot := initializeBot(client, twitterClient)
	run(client, bot)
}

func initializeBot(client *disgord.Client, twitterClient *myTwitter.TwitterClient) discordBot {
	dbClient := db.NewClient()
	setupTwitterClient(client, dbClient, twitterClient)
	return discordBot{
		commands: dbClient,
		twitterClient: twitterClient,
	}
}

func run(client *disgord.Client, bot discordBot) {

	content, _ := std.NewMsgFilter(context.Background(), client)
	customMiddleWare, _ := newMiddlewareHolder(context.Background(), client)
	content.SetPrefix("!")

	// listen for messages
	client.Gateway().
		WithMiddleware(customMiddleWare.filterBotMsg, bot.commandInUse).
		MessageCreate(bot.reactRoleMessage)
	client.Gateway().
		WithMiddleware(content.HasPrefix, customMiddleWare.isReactMessageCommand).
		MessageCreate(bot.reactionRoleCommand)
	client.Gateway().
		WithMiddleware(content.HasPrefix, customMiddleWare.isTwitterFollowCommand).
		MessageCreate(bot.twitterFollowCommand)
	client.Gateway().
		WithMiddleware(content.HasPrefix, customMiddleWare.isTwitterFollowListCommand).
		MessageCreate(bot.twitterFollowList)
	client.Gateway().
		WithMiddleware(content.HasPrefix, customMiddleWare.isTwitterFollowRemoveCommand).
		MessageCreate(bot.twitterUnfollowCommand)
	client.Gateway().
		WithMiddleware(customMiddleWare.filterOutBots, bot.reactionMessage).
		MessageReactionAdd(bot.addRole)
	client.Gateway().
		WithMiddleware(customMiddleWare.filterOutBots, bot.reactionMessage).
		MessageReactionRemove(bot.removeRole)

	// connect now, and disconnect on system interrupt
	client.Gateway().StayConnectedUntilInterrupted()
}

func setupTwitterClient(client *disgord.Client, dbClient botcommands.SavedTwitterFollowCommand, twitterClient *myTwitter.TwitterClient) {
	tweetHandler := func(tweet *twitter.Tweet) {
		discordMessage := fmt.Sprintf("New Tweet by %s https://www.twitter.com/%s/status/%s", tweet.User.Name, tweet.User.ScreenName, tweet.IDStr)

		newMessageParams := &disgord.CreateMessageParams {
			Content: discordMessage,
		}
		twitterFollowCommands := dbClient.GetFollowedUser(tweet.User.ScreenName)
		for i := range twitterFollowCommands {
			client.Channel(twitterFollowCommands[i].Channel).CreateMessage(newMessageParams)
		}
		
	}
	twitterClient.SetTweetDemux(tweetHandler)
}

func (bot *discordBot) reactionRoleCommand(s disgord.Session, data *disgord.MessageCreate) {
	msg := data.Message
		
	msg.Reply(context.Background(), s, "Which channel should this message be sent in.")
	command := botcommands.CommandInProgress{
		Guild: msg.GuildID,
		User: msg.Author.ID,
		Stage: 1 }
	bot.commands.SaveCommandInProgress(msg.Author, command)
}

func (bot *discordBot) twitterFollowCommand(s disgord.Session, data *disgord.MessageCreate) {
	msg := data.Message
	command := strings.Split(msg.Content, " ")
	if len(command) != 3 {
		msg.Reply(context.Background(), s, "Missing screen name of person to follow. Command use !twitter-follow screenName channel")
		return
	}
	screenName := command[1]
	channelName := command[2]

	channels, _ := s.Guild(msg.GuildID).GetChannels()
	for i := range channels {
		if channels[i].Name == channelName {
			twitterFollowCommand := botcommands.TwitterFollowCommand{
				ScreenName: screenName,
				Channel: channels[i].ID,
				Guild: msg.GuildID,
			}
			bot.commands.SaveUserToFollow(twitterFollowCommand)
			bot.twitterClient.AddUserToTrack(screenName)
			return
		}
	}
	msg.Reply(context.Background(), s, "Channel not found")
}

func (bot *discordBot)  twitterFollowList(s disgord.Session, data *disgord.MessageCreate) {
	followList := ""
	followsInGuild := bot.commands.GetAllFollowedUsersInServer(data.Message.GuildID)
	for i := range followsInGuild{
		followList +=  followsInGuild[i].ScreenName + "\n"
	}

	data.Message.Reply(context.Background(), s, "Following:\n" + followList)
}

func (bot *discordBot) twitterUnfollowCommand(s disgord.Session, data *disgord.MessageCreate) {
	msg := data.Message
	userToUnfollow := msg.Content
	bot.commands.DeleteFollowedUser(userToUnfollow, msg.GuildID)
}

func (bot *discordBot) reactRoleMessage(s disgord.Session, data *disgord.MessageCreate) {
	msg := data.Message

	commandInProgress := bot.commands.GetCommandInProgress(msg.Author)
	switch commandInProgress.Stage {
	case 1:
		//stage 1 ask for channel
		channels, _ := s.Guild(msg.GuildID).GetChannels()
		for i := range channels {
			if channels[i].Name == msg.Content {
				commandInProgress.Channel = channels[i].ID
				msg.Reply(context.Background(), s, "Enter role to be assigned")
				commandInProgress.Stage = 2
				bot.commands.SaveCommandInProgress(msg.Author, *commandInProgress)
				return
			}
		}
		msg.Reply(context.Background(), s, "Channel not found. Aborting command.")
		bot.commands.RemoveCommandProgress(msg.Author.ID)
	case 2:
		//stage 2 ask for role
		roles, _ := s.Guild(msg.GuildID).GetRoles()
		for i := range roles {
			if roles[i].Name == msg.Content {
				commandInProgress.Role = roles[i].ID
				msg.Reply(context.Background(), s, "Enter reaction to use.")
				commandInProgress.Stage = 3
				bot.commands.SaveCommandInProgress(msg.Author, *commandInProgress)
				return
			}
		}
		msg.Reply(context.Background(), s, "Role not found. Aborting command.")
		bot.commands.RemoveCommandProgress(msg.Author.ID)
	case 3:
		//stage 3 ask for reaction
		emojis, _ := s.Guild(msg.GuildID).GetEmojis()
		for i := range emojis {
			emojiName := strings.Split(msg.Content, ":")
			if emojis[i].Name == emojiName[1] {
				commandInProgress.Emoji = emojis[i].ID
				msg.Reply(context.Background(), s, "Enter message to use")
				commandInProgress.Stage = 4
				bot.commands.SaveCommandInProgress(msg.Author, *commandInProgress)
				return
			}
		}
		msg.Reply(context.Background(), s, "Reaction not found. Aborting command.")
		bot.commands.RemoveCommandProgress(msg.Author.ID)
	case 4:
		//stage 4 ask for message
		channels, _ := s.Guild(msg.GuildID).GetChannels()
		var commandChannel *disgord.Channel
		for i := range channels {
			if channels[i].ID == commandInProgress.Channel {
				commandChannel = channels[i]
			}
		}

		botMsg, _ := commandChannel.SendMsg(context.Background(), s, msg)

		emojis, _ := s.Guild(msg.GuildID).GetEmojis()
		var emoji *disgord.Emoji
		for i := range emojis {
			if emojis[i].ID == commandInProgress.Emoji {
				emoji = emojis[i]
			}
		}
		botMsg.React(context.Background(), s, emoji)
		roleCommand := botcommands.CompletedRoleCommand{
			Guild: commandInProgress.Guild,
			Role: commandInProgress.Role,
			Emoji: commandInProgress.Emoji}
		bot.commands.SaveRoleCommand(botMsg.ID, roleCommand)
		bot.commands.RemoveCommandProgress(msg.Author.ID)
	default:
		//error
	}
}

func (bot *discordBot) addRole(s disgord.Session, data *disgord.MessageReactionAdd) {
	userID := data.UserID
	command := bot.commands.GetRoleCommand(data.MessageID)
	s.Guild(command.Guild).Member(userID).AddRole(command.Role)
}

func (bot *discordBot) removeRole(s disgord.Session, data *disgord.MessageReactionRemove) {
	userID := data.UserID
	command := bot.commands.GetRoleCommand(data.MessageID)
	s.Guild(command.Guild).Member(userID).RemoveRole(command.Role)
}

func newMiddlewareHolder(ctx context.Context, s disgord.Session) (m *middlewareHolder, err error) {
	m = &middlewareHolder{session: s}
	if m.myself, err = s.CurrentUser().WithContext(ctx).Get(); err != nil {
		return nil, errors.New("unable to fetch info about the bot instance")
	}
	return m, nil
}