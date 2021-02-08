package help

import (
	"context"
	"discordbot/botcommands/discord"

	"github.com/andersfylling/disgord"
)


const HelpString = "help"

type HelpCommand struct {
	commands []PrintHelp
}

type PrintHelp interface {
	PrintHelp() string
}

func New(commands []PrintHelp) *HelpCommand {
	return &HelpCommand{
		commands: commands,
	}
}

func (c *HelpCommand) PrintHelp() string {
	return "help - nani"
}

func (c *HelpCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, middleWareContent discord.MiddleWareContent) {
	helpList := "```Available Commands:\n"
	for _ , command := range c.commands {
		helpList += command.PrintHelp() + "\n"
	}
	helpList += "```"
	data.Message.Reply(context.Background(), s, helpList)
}