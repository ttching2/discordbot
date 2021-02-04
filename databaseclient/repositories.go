package databaseclient

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

/*
users {
	user_id int
	discord_user_id int
	admin bool default false
}

rolecommand{
	role_command_id
	author fk users(discord_user_id)
	guild int
	channel int
	role int
	emoji int
	message int
	complete bool default false
	stage int default 0
}

twitter_follow_command {
	twitter_follow_command_id int
	author fk users(discord_user_id)
	screen_name string
	screen_name_id string
	channel int
	guild int
}

strawpoll_deadline_command {
	strawpoll_deadline_id int
	author fk users(discord_user_id)
	strawpoll_id string
	guild int
	channel int
	role int
}
*/