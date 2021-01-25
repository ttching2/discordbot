package botcommands

import (
	"github.com/andersfylling/disgord"
)

type TwitterFollowCommand struct {
	TwitterFollowCommandID int
	ScreenName string
	ScreenNameID string
	Channel disgord.Snowflake
	Guild disgord.Snowflake
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
SaveableCommand - interface for all commands? or maybe just role commands
*/
type SaveableCommand interface {
	InProgressCommands
	SavedRoleCommands
	SavedTwitterFollowCommand
}

/*
 */
type InProgressCommands interface {
	SaveCommandInProgress(user *disgord.User, commmand CommandInProgress)
	SaveRoleCommand(msgID disgord.Snowflake, roleCommand CompletedRoleCommand)
	IsUserUsingCommand(user *disgord.User) bool
	GetCommandInProgress(user *disgord.User) *CommandInProgress
	RemoveCommandProgress(user disgord.Snowflake)
}

type SavedTwitterFollowCommand interface {
	GetFollowedUser(screenName string) []TwitterFollowCommand
	SaveUserToFollow(twitterFollow TwitterFollowCommand)
	DeleteFollowedUser(screenName string, guild disgord.Snowflake)
	GetAllFollowedUsersInServer(guild disgord.Snowflake) []TwitterFollowCommand
	GetAllUniqueFollowedUsers() []TwitterFollowCommand
}

/*

*/
type SavedRoleCommands interface {
	IsRoleCommandMessage(msg disgord.Snowflake, emoji disgord.Snowflake) bool
	GetRoleCommand(msg disgord.Snowflake) *CompletedRoleCommand
	RemoveRoleReactCommand(msg disgord.Snowflake)
}