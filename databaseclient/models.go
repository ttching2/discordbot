package databaseclient

type Snowflake uint64

/*
T
*/
type TwitterFollowCommand struct {
	TwitterFollowCommandID int64
	User                   int
	ScreenName             string
	ScreenNameID           string
	Channel                Snowflake
	Guild                  Snowflake
}

/*
CommandInProgress - track commands in progress of being made for role message
*/
type CommandInProgress struct {
	CommandInProgressID int64
	Guild               Snowflake
	Channel             Snowflake
	User                Snowflake
	Role                Snowflake
	Emoji               Snowflake
	Stage               int
}

/*
RoleCommand - role messages to keep track of
*/
type RoleCommand struct {
	RoleCommandID int64
	User          int
	Guild         Snowflake
	Role          Snowflake
	Emoji         Snowflake
	Message       Snowflake
}

/*
StrawpollDeadline db model
*/
type StrawpollDeadline struct {
	StrawpollDeadlineID int64
	User                int
	StrawpollID         string
	Guild               Snowflake
	Channel             Snowflake
	Role                Snowflake
}