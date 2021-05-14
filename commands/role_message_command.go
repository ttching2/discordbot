package commands

import (
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/sirupsen/logrus"
)

const RoleReactString = "react"

type roleCommandRequestFactory struct {
	repo    RoleReactRepository
	session DiscordSession
}

func NewRoleCommandRequestFactory(s DiscordSession, repo RoleReactRepository) *roleCommandRequestFactory {
	return &roleCommandRequestFactory{
		session: s,
		repo:    repo,
	}
}

func (c *roleCommandRequestFactory) PrintHelp() string {
	return CommandPrefix + RoleReactString + " - command to create a reaction role message for assigning roles to people. Usage follow commands given by bot."
}

func (c *roleCommandRequestFactory) CreateRequest(data *disgord.MessageCreate, user *Users) interface{} {
	return &roleCommandRequest{
		roleCommandRequestFactory: c,
		data:                      data,
		user:                      user,
	}
}

type roleCommandRequest struct {
	*roleCommandRequestFactory
	data *disgord.MessageCreate
	user *Users
}

func (c roleCommandRequest) ExecuteMessageCreateCommand() {
	msg := c.data.Message

	c.session.SendSimpleMessage(msg.ChannelID, "Which channel should this message be sent in.")
	command := CommandInProgress{
		Guild:         msg.GuildID,
		OriginChannel: msg.ChannelID,
		User:          msg.Author.ID,
		Stage:         1}
	c.repo.SaveCommandInProgress(&command)
}

type inProgressRoleCommand struct {
	repo    RoleReactRepository
	session DiscordSession
	data    *disgord.MessageCreate
	user    *Users
}

func NewInProgressRoleCommand(s DiscordSession, r RoleReactRepository, d *disgord.MessageCreate, u *Users) *inProgressRoleCommand {
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
		c.session.ReactToMessage(msg.ID, msg.ChannelID, "👎")
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
		targetChannel := FindTargetChannel(commandInProgress.TargetChannel, guild)
		reactEmoji := FindTargetEmoji(commandInProgress.Emoji, guild)

		msg, err := c.session.SendSimpleMessage(targetChannel.ID, msg.Content)
		if err != nil {
			log.Error(err)
			return
		}
		c.session.ReactToMessage(msg.ID, targetChannel.ID, reactEmoji)
		err = c.saveRoleCommand(c.user.UsersID, commandInProgress, msg.ID)
		if err != nil {
			log.Error(err)
			return
		}
		err = c.repo.RemoveCommandProgress(commandInProgress.User, commandInProgress.OriginChannel)
		if err != nil {
			log.Error(err)
		}
	default:
	}
}

func (c *inProgressRoleCommand) getChannelFromUser(g Guild, commandInProgress CommandInProgress, s DiscordSession) {
	msg := c.data.Message
	channel := FindChannelByName(msg.Content, g)
	if channel == nil {
		c.session.SendSimpleMessage(msg.ChannelID, "Channel not found. Aborting command.")
		c.repo.RemoveCommandProgress(msg.Author.ID, msg.ChannelID)
		return
	}
	commandInProgress.TargetChannel = channel.ID
	c.session.SendSimpleMessage(msg.ChannelID, "Enter role to be assigned")
	commandInProgress.Stage = 2
	c.repo.SaveCommandInProgress(&commandInProgress)
}

func (c *inProgressRoleCommand) getRoleFromUser(g Guild, commandInProgress CommandInProgress, s DiscordSession) {
	msg := c.data.Message
	roles, _ := g.GetRoles()
	role := FindRoleByName(msg.Content, roles)
	if role == nil {
		c.session.SendSimpleMessage(msg.ChannelID, "Role not found. Aborting command.")
		c.repo.RemoveCommandProgress(msg.Author.ID, msg.ChannelID)
		return
	}
	commandInProgress.Role = role.ID
	c.session.SendSimpleMessage(msg.ChannelID, "Enter reaction to use.")
	commandInProgress.Stage = 3
	c.repo.SaveCommandInProgress(&commandInProgress)
}

func (c *inProgressRoleCommand) getEmojiFromUser(g Guild, commandInProgress CommandInProgress, s DiscordSession) {
	msg := c.data.Message
	emojis, _ := g.GetEmojis()
	//TODO if it uses a unicode emoji we get a panik
	emojiName := strings.Split(msg.Content, ":")
	emoji := FindEmojiByName(emojiName[1], emojis)
	if emoji == nil {
		c.session.SendSimpleMessage(msg.ChannelID, "Reaction not found. Aborting command.")
		c.repo.RemoveCommandProgress(msg.Author.ID, msg.ChannelID)
		return
	}
	commandInProgress.Emoji = emoji.ID
	c.session.SendSimpleMessage(msg.ChannelID, "Enter message to use")
	commandInProgress.Stage = 4
	c.repo.SaveCommandInProgress(&commandInProgress)
}

func (c *inProgressRoleCommand) saveRoleCommand(authorID int64, commandInProgress CommandInProgress, botMsgID Snowflake) error {
	roleCommand := RoleCommand{
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
	repo RoleReactRepository
	data *disgord.MessageDelete
}

func NewRemoveRoleMessage(r RoleReactRepository, d *disgord.MessageDelete) *removeRoleMessage {
	return &removeRoleMessage{
		repo: r,
		data: d,
	}
}

func (c removeRoleMessage) OnMessageDelete() {
	c.repo.RemoveRoleReactCommand(c.data.MessageID)
}

type addRoleReact struct {
	repo    RoleReactRepository
	session DiscordSession
	data    *disgord.MessageReactionAdd
}

func NewAddRoleReact(r RoleReactRepository, s DiscordSession, d *disgord.MessageReactionAdd) *addRoleReact {
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
	repo    RoleReactRepository
	session DiscordSession
	data    *disgord.MessageReactionRemove
}

func NewRemoveRoleReact(r RoleReactRepository, s DiscordSession, d *disgord.MessageReactionRemove) *removeRoleReact {
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
