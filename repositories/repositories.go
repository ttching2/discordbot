package repositories

import "discordbot/repositories/model"

type UsersRepository interface {
	GetUserByDiscordId(user model.Snowflake) (model.Users, error)
	DoesUserExist(user model.Snowflake) bool
	SaveUser(*model.Users) error
}

/*
RoleReactRepository interface for role commands to db
*/
type RoleReactRepository interface {
	SaveCommandInProgress(command *model.CommandInProgress) error
	SaveRoleCommand(roleCommand *model.RoleCommand) error
	IsUserUsingCommand(user model.Snowflake, channel model.Snowflake) (bool, error)
	GetCommandInProgress(user model.Snowflake, channel model.Snowflake) (model.CommandInProgress, error)
	RemoveCommandProgress(user model.Snowflake, channel model.Snowflake) error
	IsRoleCommandMessage(msg model.Snowflake, emoji model.Snowflake) (bool, error)
	GetRoleCommand(msg model.Snowflake) (model.RoleCommand, error)
	RemoveRoleReactCommand(msg model.Snowflake) error
}

/*
TwitterFollowRepository interface for twitter follows
*/
type TwitterFollowRepository interface {
	GetFollowedUser(screenName string) ([]model.TwitterFollowCommand, error)
	SaveUserToFollow(twitterFollow *model.TwitterFollowCommand) error
	DeleteFollowedUser(screenName string, guild model.Snowflake) error
	GetAllFollowedUsersInServer(guild model.Snowflake) ([]model.TwitterFollowCommand, error)
	GetAllUniqueFollowedUsers() ([]model.TwitterFollowCommand, error)
}

/*
StrawpollDeadlineRepository interface for strawpoll commands
*/
type StrawpollDeadlineRepository interface {
	SaveStrawpollDeadline(*model.StrawpollDeadline) error
	GetAllStrawpollDeadlines() ([]model.StrawpollDeadline, error)
	DeleteStrawpollDeadlineByID(ID int64) error
}