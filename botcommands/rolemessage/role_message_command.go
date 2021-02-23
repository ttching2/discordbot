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

type inProgressRoleCommand struct {
	repo    repositories.RoleReactRepository
	session disgord.Session
	data    *disgord.MessageCreate
	user    *model.Users
}

func NewInProgressRoleCommand(s disgord.Session, r repositories.RoleReactRepository, d *disgord.MessageCreate, u *model.Users) *inProgressRoleCommand {
	return &inProgressRoleCommand{
		session: s,
		repo:    r,
		data:    d,
		user:    u,
	}
}

func (c *inProgressRoleCommand) ExecuteMessageCreateCommand() {
	msg := c.data.Message

	commandInProgress, err := c.repo.GetCommandInProgress(msg.Author.ID, msg.ChannelID)
	if err != nil {
		msg.React(context.Background(), c.session, "👎")
		log.Error(err)
		return
	}
	guild := c.session.Guild(msg.GuildID)

	switch commandInProgress.Stage {
	case 1:
		c.getChannelFromUser(guild, commandInProgress, c.session)
	case 2:
		c.getRoleFromUser(guild, commandInProgress, c.session)
	case 3:
		c.getEmojiFromUser(guild, commandInProgress, c.session)
	case 4:
		targetChannel := util.FindTargetChannel(commandInProgress.TargetChannel, guild)
		reactEmoji := util.FindTargetEmoji(commandInProgress.Emoji, guild)

		msg, err := targetChannel.SendMsg(context.Background(), c.session, msg)
		if err != nil {
			log.Error(err)
			return
		}
		msg.React(context.Background(), c.session, reactEmoji)
		err = c.saveRoleCommand(c.user.UsersID, commandInProgress, msg.ID)
		if err != nil {
			return
		}
		c.repo.RemoveCommandProgress(c.user.DiscordUsersID, msg.ChannelID)
	default:
	}
}

func (c *inProgressRoleCommand) getChannelFromUser(g discord.Guild, commandInProgress model.CommandInProgress, s disgord.Session) {
	msg := c.data.Message
	channel := util.FindChannelByName(msg.Content, g)
	if channel == nil {
		msg.Reply(context.Background(), s, "Channel not found. Aborting command.")
		c.repo.RemoveCommandProgress(msg.Author.ID, msg.ChannelID)
		return
	}
	commandInProgress.TargetChannel = channel.ID
	msg.Reply(context.Background(), s, "Enter role to be assigned")
	commandInProgress.Stage = 2
	c.repo.SaveCommandInProgress(&commandInProgress)
}

func (c *inProgressRoleCommand) getRoleFromUser(g discord.Guild, commandInProgress model.CommandInProgress, s disgord.Session) {
	msg := c.data.Message
	roles, _ := g.GetRoles()
	role := util.FindRoleByName(msg.Content, roles)
	if role == nil {
		msg.Reply(context.Background(), s, "Role not found. Aborting command.")
		c.repo.RemoveCommandProgress(msg.Author.ID, msg.ChannelID)
		return
	}
	commandInProgress.Role = role.ID
	msg.Reply(context.Background(), s, "Enter reaction to use.")
	commandInProgress.Stage = 3
	c.repo.SaveCommandInProgress(&commandInProgress)
}

func (c *inProgressRoleCommand) getEmojiFromUser(g discord.Guild, commandInProgress model.CommandInProgress, s disgord.Session) {
	msg := c.data.Message
	emojis, _ := g.GetEmojis()
	//TODO if it uses a unicode emoji we get a panik
	emojiName := strings.Split(msg.Content, ":")
	emoji := util.FindEmojiByName(emojiName[1], emojis)
	if emoji == nil {
		msg.Reply(context.Background(), s, "Reaction not found. Aborting command.")
		c.repo.RemoveCommandProgress(msg.Author.ID, msg.ChannelID)
		return
	}
	commandInProgress.Emoji = emoji.ID
	msg.Reply(context.Background(), s, "Enter message to use")
	commandInProgress.Stage = 4
	c.repo.SaveCommandInProgress(&commandInProgress)
}

func (c *inProgressRoleCommand) saveRoleCommand(authorID int64, commandInProgress model.CommandInProgress, botMsgID model.Snowflake) error {
	roleCommand := model.RoleCommand{
		User:    authorID,
		Guild:   commandInProgress.Guild,
		Role:    commandInProgress.Role,
		Emoji:   commandInProgress.Emoji,
		Message: botMsgID,
	}
	err := c.repo.SaveRoleCommand(&roleCommand)

	if err != nil {
		log.WithFields(logrus.Fields{
			"roleCommand": roleCommand,
		}).Error(err)
		return err
	}
	return nil
}

type removeRoleMessage struct {
	repo repositories.RoleReactRepository
	data *disgord.MessageDelete
}

func NewRemoveRoleMessage(r repositories.RoleReactRepository, d *disgord.MessageDelete) *removeRoleMessage {
	return &removeRoleMessage{
		repo: r,
		data: d,
	}
}

func (c removeRoleMessage) OnMessageDelete() {
	c.repo.RemoveRoleReactCommand(c.data.MessageID)
}

type addRoleReact struct {
	repo    repositories.RoleReactRepository
	session disgord.Session
	data    *disgord.MessageReactionAdd
}

func NewAddRoleReact(r repositories.RoleReactRepository, s disgord.Session, d *disgord.MessageReactionAdd) *addRoleReact {
	return &addRoleReact{
		repo:    r,
		session: s,
		data:    d,
	}
}

//Bot role needs to be above role to give the role.
func (c *addRoleReact) OnReactionAdd() {
	userID := c.data.UserID
	command, err := c.repo.GetRoleCommand(c.data.MessageID)
	if err != nil {
		log.Error(err)
	}
	c.session.Guild(command.Guild).Member(userID).AddRole(command.Role)
}

type removeRoleReact struct {
	repo    repositories.RoleReactRepository
	session disgord.Session
	data    *disgord.MessageReactionRemove
}

func NewRemoveRoleReact(r repositories.RoleReactRepository, s disgord.Session, d *disgord.MessageReactionRemove) *removeRoleReact {
	return &removeRoleReact{
		repo:    r,
		session: s,
		data:    d,
	}
}

func (c *removeRoleReact) OnReactionRemove() {
	userID := c.data.UserID
	command, err := c.repo.GetRoleCommand(c.data.MessageID)
	if err != nil {
		log.Error(err)
	}
	c.session.Guild(command.Guild).Member(userID).RemoveRole(command.Role)
}
