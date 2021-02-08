package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"discordbot/botcommands"
	"discordbot/botcommands/discord"
	"discordbot/botcommands/help"
	"discordbot/botcommands/rolemessage"
	"discordbot/botcommands/strawpolldeadline"
	"discordbot/botcommands/twittercommands"
	"discordbot/repositories/rolecommand"
	strawpollrepo "discordbot/repositories/strawpolldeadline"
	"discordbot/repositories/twitterfollow"
	"discordbot/repositories/users_repository"
	"discordbot/strawpoll"
	myTwitter "discordbot/twitter"

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
	twitterClient *myTwitter.TwitterClient
	strawpollClient *strawpoll.Client
	commands map[string]interface{}
	customMiddleWare *middlewareHolder
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

func newSqlDb() *sql.DB {
	client, err := sql.Open("sqlite3", "botdb?_foreign_keys=on")

	if err != nil {
		log.Fatal(err)
	}

	return client
}

func initializeBot(client *disgord.Client, config BotConfig) discordBot {
	twitterClient := myTwitter.NewClient(config.TwitterConfig)
	strawpollClient := strawpoll.New(config.StrawPollConfig)
	sqlDb := newSqlDb()

	roleCommandRepository := rolecommand.New(sqlDb)
	twitterCommandRepository := twitterfollow.New(sqlDb)
	strawpollCommandRepository := strawpollrepo.New(sqlDb)
	usersRepository := users_repository.New(sqlDb)

	twittercommands.RestartTwitterFollows(client, twitterCommandRepository, twitterClient)
	
	strawpolldeadline.RestartStrawpollDeadlines(client, strawpollCommandRepository, strawpollClient)

	commands := make(map[string]interface{})
	commands[rolemessage.RoleReactString] = rolemessage.New(roleCommandRepository)
	commands[twittercommands.TwitterFollowString] = twittercommands.NewTwitterFollowCommand(twitterClient, twitterCommandRepository)
	commands[twittercommands.TwitterFollowListString] = twittercommands.NewTwitterFollowListCommand(twitterCommandRepository)
	commands[twittercommands.TwitterUnfollowString] = twittercommands.NewTwitterUnfollowCommand(twitterClient, twitterCommandRepository)
	commands[strawpolldeadline.StrawPollDeadlineString] = strawpolldeadline.New(strawpollClient, strawpollCommandRepository)

	var commandList []help.PrintHelp
	for _, c := range commands {
		commandList = append(commandList, c.(help.PrintHelp))
	}
	commands[help.HelpString] = help.New(commandList[:])

	customMiddleWare, _ := newMiddlewareHolder(context.Background(), client, roleCommandRepository, usersRepository)

	return discordBot{
		twitterClient: twitterClient,
		commands: commands,
		strawpollClient: strawpollClient,
		customMiddleWare: customMiddleWare,
	}
}

func run(client *disgord.Client, bot discordBot) {

	content, _ := std.NewMsgFilter(context.Background(), client)
	content.SetPrefix(botcommands.CommandPrefix)

	//TODO Find a better way for doing this
	reactMessage := bot.commands[rolemessage.RoleReactString].(*rolemessage.RoleMessageCommand)
	// listen for messages
	client.Gateway().
		WithMiddleware(bot.customMiddleWare.filterBotMsg, bot.customMiddleWare.commandInUse, bot.customMiddleWare.createMessageContentForNonCommand).
		MessageCreate(bot.SpecialCase)
	client.Gateway().
		WithMiddleware(bot.customMiddleWare.filterBotMsg, content.StripPrefix, bot.customMiddleWare.isFromAdmin, bot.customMiddleWare.checkAndSaveUser).
		MessageCreate(bot.ExecuteCommand)
	client.Gateway().
		WithMiddleware(bot.customMiddleWare.reactionMessage).
		MessageDelete(reactMessage.RemoveReactRoleMessage)
	client.Gateway().
		WithMiddleware(bot.customMiddleWare.filterOutBots, bot.customMiddleWare.reactionMessage).
		MessageReactionAdd(reactMessage.AddRole)
	client.Gateway().
		WithMiddleware(bot.customMiddleWare.filterOutBots, bot.customMiddleWare.reactionMessage).
		MessageReactionRemove(reactMessage.RemoveRole)

	// connect now, and disconnect on system interrupt
	client.Gateway().StayConnectedUntilInterrupted()
}

func (bot *discordBot) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate) {
	msg := data.Message
	middleWareContent := discord.MiddleWareContent{}
	json.Unmarshal([]byte(msg.Content), &middleWareContent)
	//TODO could be done better :/
	if _, ok := bot.commands[middleWareContent.Command]; !ok {
		msg.Reply(context.Background(), s, fmt.Sprintf("Command %s not found", middleWareContent.Command))
		return
	}
	bot.commands[middleWareContent.Command].(discord.MessageCreateHandler).ExecuteCommand(s, data, middleWareContent)
}

//TODO REDO THIS SOMEHOW :^(
func (bot *discordBot) SpecialCase(s disgord.Session, data *disgord.MessageCreate) {
	msg := data.Message
	middleWareContent := discord.MiddleWareContent{}
	json.Unmarshal([]byte(msg.Content), &middleWareContent)
	bot.commands[rolemessage.RoleReactString].(*rolemessage.RoleMessageCommand).ReactRoleMessage(s, data, middleWareContent)
}

func (bot *discordBot) reactionAdd(disgord.Session, *disgord.MessageReactionAdd) {

}

func (bot *discordBot) reactonRemove(disgord.Session, *disgord.MessageReactionRemove) {

}

func (bot *discordBot) messageDelete(disgord.Session, *disgord.MessageDelete) {

}