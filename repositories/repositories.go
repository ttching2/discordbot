package repositories

type UsersRepository interface {
	GetUserByDiscordId(user Snowflake) (Users, error)
	DoesUserExist(user Snowflake) (bool, error)
	SaveUser(user *Users) error
}

/*
RoleReactRepository interface for role commands to db
*/
type RoleReactRepository interface {
	SaveCommandInProgress(command *CommandInProgress) error
	SaveRoleCommand(roleCommand *RoleCommand) error
	IsUserUsingCommand(user Snowflake, channel Snowflake) (bool, error)
	GetCommandInProgress(user Snowflake, channel Snowflake) (CommandInProgress, error)
	RemoveCommandProgress(user Snowflake, channel Snowflake) error
	IsRoleCommandMessage(msg Snowflake, emoji Snowflake) (bool, error)
	GetRoleCommand(msg Snowflake) (RoleCommand, error)
	RemoveRoleReactCommand(msg Snowflake) error
}

/*
TwitterFollowRepository interface for twitter follows
*/
type TwitterFollowRepository interface {
	GetFollowedUser(screenName string) ([]TwitterFollowCommand, error)
	SaveUserToFollow(twitterFollow *TwitterFollowCommand) error
	DeleteFollowedUser(screenName string, guild Snowflake) error
	GetAllFollowedUsersInServer(guild Snowflake) ([]TwitterFollowCommand, error)
	GetAllUniqueFollowedUsers() ([]TwitterFollowCommand, error)
}

/*
StrawpollDeadlineRepository interface for strawpoll commands
*/
type StrawpollDeadlineRepository interface {
	SaveStrawpollDeadline(strawpollDeadline *StrawpollDeadline) error
	GetAllStrawpollDeadlines() ([]StrawpollDeadline, error)
	DeleteStrawpollDeadlineByID(ID int64) error
}