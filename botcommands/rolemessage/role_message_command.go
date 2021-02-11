package rolemessage

import (
	"context"
	"discordbot/botcommands"
	"discordbot/botcommands/discord"
	"discordbot/repositories"
	"discordbot/util"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const RoleReactString = "react"

type RoleMessageCommand struct {
	repo repositories.RoleReactRepository
}

func New(repo repositories.RoleReactRepository) *RoleMessageCommand {
	return &RoleMessageCommand{
		repo: repo,
	}
}

func (c *RoleMessageCommand) PrintHelp() string { 
	return botcommands.CommandPrefix + RoleReactString + " - command to create a reaction role message for assigning roles to people. Usage follow commands given by bot."
}

func (c *RoleMessageCommand) ExecuteCommand(s disgord.Session, data *disgord.MessageCreate, middleWareContent discord.MiddleWareContent) {
	msg := data.Message

	msg.Reply(context.Background(), s, "Which channel should this message be sent in.")
	command := repositories.CommandInProgress{
		Guild:         msg.GuildID,
		OriginChannel: msg.ChannelID,
		User:          msg.Author.ID,
		Stage:         1}
	c.repo.SaveCommandInProgress(&command)
}

func (c *RoleMessageCommand) ReactRoleMessage(s disgord.Session, data *disgord.MessageCreate, middleWareContent discord.MiddleWareContent) {
	msg := data.Message

	commandInProgress, err := c.repo.GetCommandInProgress(msg.Author.ID, msg.ChannelID)
	if err != nil {
		msg.React(context.Background(), s, "👎")
		log.Error(err)
		return
	}
	guild := s.Guild(msg.GuildID)
	messageContents := discord.DiscordMessageInfo{
		Content:   middleWareContent.MessageContent,
		UserID:  msg.Author.ID,
		AuthorID: middleWareContent.UsersID,
		ChannelID: msg.ChannelID,
		Reply:     msg.Reply,
	}
	switch commandInProgress.Stage {
	case 1:
		c.getChannelFromUser(guild, messageContents, commandInProgress, s)
	case 2:
		c.getRoleFromUser(guild, messageContents, commandInProgress, s)
	case 3:
		c.getEmojiFromUser(guild, messageContents, commandInProgress, s)
	case 4:
		c.createRoleMessageCommand(guild, messageContents, commandInProgress, s)
	default:
	}
}

func (c *RoleMessageCommand) getChannelFromUser(g discord.Guild, msg discord.DiscordMessageInfo, commandInProgress repositories.CommandInProgress, s disgord.Session) {
	channels, _ := g.GetChannels()
	channel := util.FindChannelByName(msg.Content, channels)
	if channel == nil {
		msg.Reply(context.Background(), s, "Channel not found. Aborting command.")
		c.repo.RemoveCommandProgress(msg.UserID, msg.ChannelID)
		return
	}
	commandInProgress.TargetChannel = channel.ID
	msg.Reply(context.Background(), s, "Enter role to be assigned")
	commandInProgress.Stage = 2
	c.repo.SaveCommandInProgress(&commandInProgress)
}

func (c *RoleMessageCommand) getRoleFromUser(g discord.Guild, msg discord.DiscordMessageInfo, commandInProgress repositories.CommandInProgress, s disgord.Session) {
	roles, _ := g.GetRoles()
	role := util.FindRoleByName(msg.Content, roles)
	if role == nil {
		msg.Reply(context.Background(), s, "Role not found. Aborting command.")
		c.repo.RemoveCommandProgress(msg.UserID, msg.ChannelID)
		return
	}
	commandInProgress.Role = role.ID
	msg.Reply(context.Background(), s, "Enter reaction to use.")
	commandInProgress.Stage = 3
	c.repo.SaveCommandInProgress(&commandInProgress)
}

func (c *RoleMessageCommand) getEmojiFromUser(g discord.Guild, msg discord.DiscordMessageInfo, commandInProgress repositories.CommandInProgress, s disgord.Session) {
	emojis, _ := g.GetEmojis()
	//TODO if it uses a unicode emoji we get a panik
	emojiName := strings.Split(msg.Content, ":")
	emoji := util.FindEmojiByName(emojiName[1], emojis)
	if emoji == nil {
		msg.Reply(context.Background(), s, "Reaction not found. Aborting command.")
		c.repo.RemoveCommandProgress(msg.UserID, msg.ChannelID)
		return
	}
	commandInProgress.Emoji = emoji.ID
	msg.Reply(context.Background(), s, "Enter message to use")
	commandInProgress.Stage = 4
	c.repo.SaveCommandInProgress(&commandInProgress)
}

func (c *RoleMessageCommand) createRoleMessageCommand(g discord.Guild, msg discord.DiscordMessageInfo, commandInProgress repositories.CommandInProgress, s disgord.Session) {
	channels, _ := g.GetChannels()
	commandChannel := util.FindChannelByID(commandInProgress.OriginChannel, channels)

	botMsg , _ := commandChannel.SendMsg(context.Background(), s, &disgord.Message{Content: msg.Content})

	emojis, _ := g.GetEmojis()
	emoji := util.FindEmojiByID(commandInProgress.Emoji, emojis)
	botMsg.React(context.Background(), s, emoji)

	roleCommand := repositories.RoleCommand {
		User:    msg.AuthorID,
		Guild:   commandInProgress.Guild,
		Role:    commandInProgress.Role,
		Emoji:   commandInProgress.Emoji,
		Message: botMsg.ID,
	}
	err := c.repo.SaveRoleCommand(&roleCommand)

	if err != nil {
		log.WithFields(logrus.Fields{
			"roleCommand": roleCommand,
		}).Error(err)
	}

	c.repo.RemoveCommandProgress(msg.UserID, msg.ChannelID)
}

func (c *RoleMessageCommand) RemoveReactRoleMessage(s disgord.Session, data *disgord.MessageDelete) {
	c.repo.RemoveRoleReactCommand(data.MessageID)
}

//Bot role needs to be above role to give the role.
func (c *RoleMessageCommand) AddRole(s disgord.Session, data *disgord.MessageReactionAdd) {
	userID := data.UserID
	command, err := c.repo.GetRoleCommand(data.MessageID)
	if err != nil {
		log.Error(err)
	}
	s.Guild(command.Guild).Member(userID).AddRole(command.Role)
}

func (c *RoleMessageCommand) RemoveRole(s disgord.Session, data *disgord.MessageReactionRemove) {
	userID := data.UserID
	command, err := c.repo.GetRoleCommand(data.MessageID)
	if err != nil {
		log.Error(err)
	}
	s.Guild(command.Guild).Member(userID).RemoveRole(command.Role)
}
