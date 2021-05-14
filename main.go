package main

import (
	"container/list"
	"context"
	"database/sql"
	"os"

	"discordbot/challonge"
	"discordbot/commands"
	"discordbot/repositories/rolecommand"
	strawpollrepo "discordbot/repositories/strawpolldeadline"
	"discordbot/repositories/tourneyrepo"
	"discordbot/repositories/twitterfollow"
	"discordbot/repositories/users_repository"
	"discordbot/strawpoll"
	myTwitter "discordbot/twitter"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/disgord/std"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var log = &logrus.Logger{
	Out:          os.Stderr,
	Formatter:    new(logrus.TextFormatter),
	Hooks:        make(logrus.LevelHooks),
	Level:        logrus.InfoLevel,
	ReportCaller: true,
}

type discordConfig struct {
	botToken string
}

type botConfig struct {
	DiscordConfig   discordConfig
	TwitterConfig   myTwitter.TwitterClientConfig
	StrawPollConfig strawpoll.StrawPollConfig
	ChallongeConfig challonge.Config
}

type discordBot struct {
	*jobQueue
}

type jobQueue struct {
	onMessageCreate  *list.List
	onReactionAdd    *list.List
	onReactionRemove *list.List
	onMessageDelete  *list.List
}

type repositoryContainer struct {
	roleCommandRepo   commands.RoleReactRepository
	twitterFollowRepo commands.TwitterFollowRepository
	strawpollRepo     commands.StrawpollDeadlineRepository
	usersRepo         commands.UsersRepository
	tournamentRepo    commands.TournamentRepository
}

func main() {
	botConfig := botConfig{
		DiscordConfig: discordConfig{
			botToken: os.Getenv("DISCORD_TOKEN"),
		},
		TwitterConfig: myTwitter.TwitterClientConfig{
			ConsumerKey:    os.Getenv("TWITTER_API_KEY"),
			ConsumerSecret: os.Getenv("TWITTER_SECRET_KEY"),
			AccessToken:    os.Getenv("TWITTER_ACCESS_TOKEN"),
			AccessSecret:   os.Getenv("TWITTER_TOKEN_SECRET"),
		},
		StrawPollConfig: strawpoll.StrawPollConfig{
			ApiKey: os.Getenv("STRAWPOLL_TOKEN"),
		},
		ChallongeConfig: challonge.Config{
			Username: os.Getenv("CHALLONGE_USERNAME"),
			Apikey:   os.Getenv("CHALLONGE_API_KEY"),
		},
	}
	client := disgord.New(disgord.Config{
		BotToken: botConfig.DiscordConfig.botToken,
		Logger:   log, // optional logging
		Cache:    &disgord.CacheNop{},
	})

	bot, customMiddleWare := initializeBot(client, botConfig)
	run(client, bot, customMiddleWare)
}

func newSQLDB() *sql.DB {
	client, err := sql.Open("sqlite3", "botdb?_foreign_keys=on")

	if err != nil {
		log.Fatal(err)
	}

	return client
}

func initializeBot(s disgord.Session, config botConfig) (*discordBot, *middlewareHolder) {
	repos := newRepositoryContainer()
	jobQueue := newJobQueue()
	twitterClient := myTwitter.NewClient(config.TwitterConfig)
	strawpollClient := strawpoll.New(config.StrawPollConfig)
	challongeClient := challonge.New(config.ChallongeConfig)

	commands.RestartTwitterFollows(s, repos.twitterFollowRepo, twitterClient)
	commands.RestartStrawpollDeadlines(s, repos.strawpollRepo, strawpollClient)

	discordSession := commands.NewSimpleDiscordSession(s)
	customMiddleWare, err := newMiddlewareHolder(discordSession, jobQueue, repos, twitterClient, strawpollClient, challongeClient)
	discordBot := &discordBot{jobQueue: jobQueue}

	if err != nil {
		log.Fatal(err)
	}

	return discordBot, customMiddleWare
}

func newRepositoryContainer() *repositoryContainer {
	sqlDb := newSQLDB()
	return &repositoryContainer{
		roleCommandRepo:   rolecommand.New(sqlDb),
		twitterFollowRepo: twitterfollow.New(sqlDb),
		strawpollRepo:     strawpollrepo.New(sqlDb),
		usersRepo:         users_repository.New(sqlDb),
		tournamentRepo:    tourneyrepo.NewRepository(sqlDb),
	}
}

func run(client *disgord.Client, bot *discordBot, customMiddleWare *middlewareHolder) {

	content, _ := std.NewMsgFilter(context.Background(), client)
	content.SetPrefix(commands.CommandPrefix)

	// listen for messages
	client.Gateway().
		WithMiddleware(customMiddleWare.filterBotMsg, customMiddleWare.commandInUse, customMiddleWare.createMessageContentForNonCommand).
		MessageCreate(bot.handleMessageCreate)
	client.Gateway().
		WithMiddleware(customMiddleWare.filterBotMsg, content.StripPrefix, customMiddleWare.isFromAdmin, customMiddleWare.handleDiscordEvent).
		MessageCreate(bot.handleMessageCreate)
	client.Gateway().
		WithMiddleware(customMiddleWare.handleDiscordEvent).
		MessageDelete(bot.messageDelete)
	client.Gateway().
		WithMiddleware(customMiddleWare.filterOutBots, customMiddleWare.handleDiscordEvent).
		MessageReactionAdd(bot.reactionAdd)
	client.Gateway().
		WithMiddleware(customMiddleWare.filterOutBots, customMiddleWare.handleDiscordEvent).
		MessageReactionRemove(bot.reactonRemove)

	// connect now, and disconnect on system interrupt
	client.Gateway().StayConnectedUntilInterrupted()
}

type onMessageCreateCommand interface {
	ExecuteMessageCreateCommand()
}

type onReactionRemove interface {
	OnReactionRemove()
}

type onReactionAdd interface {
	OnReactionAdd()
}

type onMessageDelete interface {
	OnMessageDelete()
}

func (c *discordBot) handleMessageCreate(s disgord.Session, data *disgord.MessageCreate) {
	ele := c.onMessageCreate.Front()
	c.onMessageCreate.Remove(ele)
	ele.Value.(onMessageCreateCommand).ExecuteMessageCreateCommand()
}

func (c *discordBot) reactionAdd(disgord.Session, *disgord.MessageReactionAdd) {
	ele := c.onReactionAdd.Front()
	c.onReactionAdd.Remove(ele)
	ele.Value.(onReactionAdd).OnReactionAdd()
}

func (c *discordBot) reactonRemove(disgord.Session, *disgord.MessageReactionRemove) {
	ele := c.onReactionRemove.Front()
	c.onReactionRemove.Remove(ele)
	ele.Value.(onReactionRemove).OnReactionRemove()
}

func (c *discordBot) messageDelete(disgord.Session, *disgord.MessageDelete) {
	ele := c.onMessageDelete.Front()
	c.onMessageDelete.Remove(ele)
	ele.Value.(onMessageDelete).OnMessageDelete()
}

func newJobQueue() *jobQueue {
	return &jobQueue{
		onMessageCreate:  list.New(),
		onReactionAdd:    list.New(),
		onReactionRemove: list.New(),
		onMessageDelete:  list.New()}
}
