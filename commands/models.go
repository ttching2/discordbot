package commands

import "github.com/andersfylling/disgord"

type Snowflake = disgord.Snowflake

type Users struct {
	UsersID        int64
	DiscordUsersID Snowflake
	UserName       string
	IsAdmin        bool
}

/*
T
*/
type TwitterFollowCommand struct {
	TwitterFollowCommandID int64
	User                   int64
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
	OriginChannel       Snowflake
	TargetChannel       Snowflake
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
	User          int64
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
	User                int64
	StrawpollID         string
	Guild               Snowflake
	Channel             Snowflake
	Role                Snowflake
}

type Tournament struct {
	TournamentID    int64
	User            int64
	ChallongeID     string
	DiscordServerID Snowflake
	Organizers      []Users
	Participants    []TournamentParticipant
	CurrentMatch    int
}

type TournamentParticipant struct {
	TournamentParticipantID int64
	Name                    string
	ChallongeID             int
}

type MangaNotification struct {
	MangaNotificationID int64
	User                int64
	Guild               Snowflake
	Channel             Snowflake
	Role                Snowflake
}

type MangaLink struct {
	MangaLinkID        int64
	MangaLink          string
	MangaNotifications []MangaNotification
}
