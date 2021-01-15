package botcommands

import (
	"github.com/andersfylling/disgord"
)

/*
 */
type InProgressCommands interface {
	SaveCommandInProgress(user *disgord.User, commmand CommandInProgress)
	SaveRoleCommand(msgID disgord.Snowflake, roleCommand CompletedRoleCommand)
	IsUserUsingCommand(user *disgord.User) bool
	GetCommandInProgress(user *disgord.User) *CommandInProgress
	RemoveCommandProgress(user disgord.Snowflake)
}

type TwitterFollowCommand struct {
	TwitterFollowCommandID int
	ScreenName string
	Channel disgord.Snowflake
	Guild disgord.Snowflake
}

type SavedTwitterFollowCommand interface {
	GetFollowedUser(screenName string) []TwitterFollowCommand
	SaveUserToFollow(twitterFollow TwitterFollowCommand)
	DeleteFollowedUser(screenName string, guild disgord.Snowflake)
	GetAllFollowedUsersInServer(guild disgord.Snowflake) []TwitterFollowCommand
}

/*

*/
type SavedRoleCommands interface {
	IsRoleCommandMessage(msg disgord.Snowflake, emoji disgord.Snowflake) bool
	GetRoleCommand(msg disgord.Snowflake) *CompletedRoleCommand
	// RemoveRoleReactCommand()
}

/*
Commands - interface for all commands? or maybe just role commands
*/
type Commands interface {
	InProgressCommands
	SavedRoleCommands
	SavedTwitterFollowCommand
}

/*
CommandInProgress - track commands in progress of being made for role message
*/
type CommandInProgress struct {
	CommandInProgressID int
	Guild disgord.Snowflake
	Channel disgord.Snowflake
	User disgord.Snowflake
	Role disgord.Snowflake
	Emoji disgord.Snowflake
	Stage int
}

/*
CompletedRoleCommand - role messages to keep track of
*/
type CompletedRoleCommand struct {
	CompletedRoleCommandID int
	Guild disgord.Snowflake
	Role disgord.Snowflake
	Emoji disgord.Snowflake
}

/*
InMemoryCommandProgress - something that keeps track of commands? Maybe just role commands?
*/
type InMemoryCommandProgress struct {
	CommandsInProgress map[disgord.Snowflake]*CommandInProgress
	CompletedRoleCommand map[disgord.Snowflake]*CompletedRoleCommand
}

func NewInMemoryCommandProgress() InMemoryCommandProgress {
	return InMemoryCommandProgress{
		make(map[disgord.Snowflake]*CommandInProgress),
		make(map[disgord.Snowflake]*CompletedRoleCommand)}
}

// func (c InMemoryCommandProgress) SaveCommandInProgress(user *disgord.User, guild disgord.Snowflake) {
// 	c.CommandsInProgress[user.ID] = &CommandInProgress{ 0, guild, nil, user, nil, nil, "", 1 }
// }

func (c InMemoryCommandProgress) IsUserUsingCommand(user *disgord.User) bool {
	_, ok := c.CommandsInProgress[user.ID]
	return ok
}

func (c InMemoryCommandProgress) GetCommandInProgress(user *disgord.User) *CommandInProgress {
	return c.CommandsInProgress[user.ID]
}

func (c InMemoryCommandProgress) RemoveCommandProgress(user disgord.Snowflake) {
	delete(c.CommandsInProgress, user)
}

func (c InMemoryCommandProgress) SaveRoleCommand(msgID disgord.Snowflake, roleCommand CompletedRoleCommand) {
	c.CompletedRoleCommand[msgID] = &roleCommand
}

// func (c InMemoryCommandProgress) IsRoleCommandMessage(msg disgord.Snowflake, emoji disgord.Snowflake) bool {
// 	command, ok := c.CompletedRoleCommand[msg]
// 	if command.Emoji.ID != emoji {
// 		return false
// 	}
// 	return ok
// }

func (c InMemoryCommandProgress) GetRoleCommand(msg disgord.Snowflake) *CompletedRoleCommand {
	return c.CompletedRoleCommand[msg]
}