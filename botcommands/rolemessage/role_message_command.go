package rolemessage

import (
	"context"
	"discordbot/botcommands"
	"discordbot/botcommands/discord"
	"discordbot/repositories"
	"discordbot/repositories/model"
	"discordbot/util"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const RoleReactString = "react"

type roleCommandRequestFactory struct {
	repo    repositories.RoleReactRepository
	session disgord.Session
}

func NewRoleCommandRequestFactory(s disgord.Session, repo repositories.RoleReactRepository) *roleCommandRequestFactory {
	return &roleCommandRequestFactory{
		session: s,
		repo:    repo,
	}
}

func (c *roleCommandRequestFactory) PrintHelp() string { 
	return botcommands.CommandPrefix + RoleReactString + " - command to create a reaction role message for assigning roles to people. Usage follow commands given by bot."
}

func (c *roleCommandRequestFactory) CreateRequest(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &roleCommandRequest{
		roleCommandRequestFactory: c,
		data:                      data,
		user:                      user,
	}
}

type roleCommandRequest struct {
	*roleCommandRequestFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c roleCommandRequest) ExecuteMessageCreateCommand() {
	msg := c.data.Message

	msg.Reply(context.Background(), c.session, "Which channel should this message be sent in.")
	command := model.CommandInProgress{
		Guild:         msg.GuildID,
		OriginChannel: msg.ChannelID,
		User:          msg.Author.ID,
		Stage:         1}
	c.repo.SaveCommandInProgress(&command)
}

type InProgressRoleCommand struct {
	Repo   repositories.RoleReactRepository
	S      disgord.Session
	Data   *disgord.MessageCreate
	UserID *model.Users
}

func (c InProgressRoleCommand) ExecuteMessageCreateCommand() {
	msg := c.Data.Message

	commandInProgress, err := c.Repo.GetCommandInProgress(msg.Author.ID, msg.ChannelID)
	if err != nil {
		msg.React(context.Background(), c.S, "👎")
		log.Error(err)
		return
	}
	guild := c.S.Guild(msg.GuildID)
	messageContents := discord.DiscordMessageInfo{
		Content:   msg.Content,
		UserID:    msg.Author.ID,
		AuthorID:  c.UserID.UsersID,
		ChannelID: msg.ChannelID,
		Reply:     msg.Reply,
	}
	switch commandInProgress.Stage {
	case 1:
		c.getChannelFromUser(guild, messageContents, commandInProgress, c.S)
	case 2:
		c.getRoleFromUser(guild, messageContents, commandInProgress, c.S)
	case 3:
		c.getEmojiFromUser(guild, messageContents, commandInProgress, c.S)
	case 4:
		targetChannel := util.FindTargetChannel(commandInProgress.TargetChannel, guild)
		reactEmoji := util.FindTargetEmoji(commandInProgress.Emoji, guild)

		msg, err := targetChannel.SendMsg(context.Background(), c.S, msg)
		if err != nil {
			log.Error(err)
			return
		}
		msg.React(context.Background(), c.S, reactEmoji)
		err = c.saveRoleCommand(messageContents.AuthorID, commandInProgress, msg.ID)
		if err != nil {
			return
		}
		c.Repo.RemoveCommandProgress(messageContents.UserID, messageContents.ChannelID)
	default:
	}
}

func (c *InProgressRoleCommand) getChannelFromUser(g discord.Guild, msg discord.DiscordMessageInfo, commandInProgress model.CommandInProgress, s disgord.Session) {
	channel := util.FindChannelByName(msg.Content, g)
	if channel == nil {
		msg.Reply(context.Background(), s, "Channel not found. Aborting command.")
		c.Repo.RemoveCommandProgress(msg.UserID, msg.ChannelID)
		return
	}
	commandInProgress.TargetChannel = channel.ID
	msg.Reply(context.Background(), s, "Enter role to be assigned")
	commandInProgress.Stage = 2
	c.Repo.SaveCommandInProgress(&commandInProgress)
}

func (c *InProgressRoleCommand) getRoleFromUser(g discord.Guild, msg discord.DiscordMessageInfo, commandInProgress model.CommandInProgress, s disgord.Session) {
	roles, _ := g.GetRoles()
	role := util.FindRoleByName(msg.Content, roles)
	if role == nil {
		msg.Reply(context.Background(), s, "Role not found. Aborting command.")
		c.Repo.RemoveCommandProgress(msg.UserID, msg.ChannelID)
		return
	}
	commandInProgress.Role = role.ID
	msg.Reply(context.Background(), s, "Enter reaction to use.")
	commandInProgress.Stage = 3
	c.Repo.SaveCommandInProgress(&commandInProgress)
}

func (c *InProgressRoleCommand) getEmojiFromUser(g discord.Guild, msg discord.DiscordMessageInfo, commandInProgress model.CommandInProgress, s disgord.Session) {
	emojis, _ := g.GetEmojis()
	//TODO if it uses a unicode emoji we get a panik
	emojiName := strings.Split(msg.Content, ":")
	emoji := util.FindEmojiByName(emojiName[1], emojis)
	if emoji == nil {
		msg.Reply(context.Background(), s, "Reaction not found. Aborting command.")
		c.Repo.RemoveCommandProgress(msg.UserID, msg.ChannelID)
		return
	}
	commandInProgress.Emoji = emoji.ID
	msg.Reply(context.Background(), s, "Enter message to use")
	commandInProgress.Stage = 4
	c.Repo.SaveCommandInProgress(&commandInProgress)
}

func (c *InProgressRoleCommand) saveRoleCommand(authorID int64, commandInProgress model.CommandInProgress, botMsgID model.Snowflake) error {
	roleCommand := model.RoleCommand{
		User:    authorID,
		Guild:   commandInProgress.Guild,
		Role:    commandInProgress.Role,
		Emoji:   commandInProgress.Emoji,
		Message: botMsgID,
	}
	err := c.Repo.SaveRoleCommand(&roleCommand)

	if err != nil {
		log.WithFields(logrus.Fields{
			"roleCommand": roleCommand,
		}).Error(err)
		return err
	}
	return nil
}

type RemoveRoleMessage struct {
	Repo repositories.RoleReactRepository
	Data *disgord.MessageDelete
}

func (c RemoveRoleMessage) OnMessageDelete() {
	c.Repo.RemoveRoleReactCommand(c.Data.MessageID)
}

type AddRoleReact struct {
	Repo    repositories.RoleReactRepository
	Session disgord.Session
	Data    *disgord.MessageReactionAdd
}

//Bot role needs to be above role to give the role.
func (c *AddRoleReact) OnReactionAdd() {
	userID := c.Data.UserID
	command, err := c.Repo.GetRoleCommand(c.Data.MessageID)
	if err != nil {
		log.Error(err)
	}
	c.Session.Guild(command.Guild).Member(userID).AddRole(command.Role)
}

type RemoveRoleReact struct {
	Repo    repositories.RoleReactRepository
	Session disgord.Session
	Data    *disgord.MessageReactionRemove
}

func (c *RemoveRoleReact) OnReactionRemove() {
	userID := c.Data.UserID
	command, err := c.Repo.GetRoleCommand(c.Data.MessageID)
	if err != nil {
		log.Error(err)
	}
	c.Session.Guild(command.Guild).Member(userID).RemoveRole(command.Role)
}
