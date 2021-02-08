package repositories

type UsersRepository interface {
	GetUserByDiscordId(user Snowflake) Users
	DoesUserExist(user Snowflake) bool
	SaveUser(user *Users) error
}

/*
RoleReactRepository interface for role commands to db
*/
type RoleReactRepository interface {
	SaveCommandInProgress(command *CommandInProgress) error
	SaveRoleCommand(roleCommand *RoleCommand) error
	IsUserUsingCommand(user Snowflake, channel Snowflake) bool
	GetCommandInProgress(user Snowflake, channel Snowflake) CommandInProgress
	RemoveCommandProgress(user Snowflake, channel Snowflake) error
	IsRoleCommandMessage(msg Snowflake, emoji Snowflake) bool
	GetRoleCommand(msg Snowflake) RoleCommand
	RemoveRoleReactCommand(msg Snowflake) error
}

/*
TwitterFollowRepository interface for twitter follows
*/
type TwitterFollowRepository interface {
	GetFollowedUser(screenName string) []TwitterFollowCommand
	SaveUserToFollow(twitterFollow *TwitterFollowCommand) error
	DeleteFollowedUser(screenName string, guild Snowflake) error
	GetAllFollowedUsersInServer(guild Snowflake) []TwitterFollowCommand
	GetAllUniqueFollowedUsers() []TwitterFollowCommand
}

/*
StrawpollDeadlineRepository interface for strawpoll commands
*/
type StrawpollDeadlineRepository interface {
	SaveStrawpollDeadline(strawpollDeadline *StrawpollDeadline) error
	GetAllStrawpollDeadlines() []StrawpollDeadline
	DeleteStrawpollDeadlineByID(ID int64) error
}