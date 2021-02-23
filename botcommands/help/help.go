package help

import (
	"context"
	"discordbot/repositories/model"

	"github.com/andersfylling/disgord"
)

const HelpString = "help"

type helpCommandFactory struct {
	commands []PrintHelp
	session  disgord.Session
}

type PrintHelp interface {
	PrintHelp() string
}

func NewCommandFactory(session disgord.Session, commands []PrintHelp) *helpCommandFactory {
	return &helpCommandFactory{
		commands: commands,
		session:  session,
	}
}

func (c *helpCommandFactory) CreateRequest(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &helpCommand {
		helpCommandFactory: c,
		data: data,
		user: user,
	}
}

type helpCommand struct {
	*helpCommandFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c *helpCommand) PrintHelp() string {
	return "help - nani"
}

func (c *helpCommand) ExecuteMessageCreateCommand() {
	helpList := "```Available Commands:\n"
	for _, command := range c.commands {
		helpList += command.PrintHelp() + "\n"
	}
	helpList += "```"
	c.data.Message.Reply(context.Background(), c.session, helpList)
}
