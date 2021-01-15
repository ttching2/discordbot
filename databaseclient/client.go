package databaseclient

import (
	"database/sql"
	"log"

	"discordbot/botcommands"

	"github.com/andersfylling/disgord"
	_ "github.com/mattn/go-sqlite3"
)

type Client struct {
	client *sql.DB
}

func NewClient() *Client {
	client, err := sql.Open("sqlite3", ".\\botdb")

	if err != nil {
		log.Fatal(err)
	}

	return &Client{
		client: client,
	}
}

func (c *Client) CloseClient() {
	c.client.Close()
}

func (c *Client) SaveCommandInProgress(user *disgord.User, command botcommands.CommandInProgress) {
	tx, err := c.client.Begin()

	if err != nil {
		log.Fatal(err)
		return
	}

	const insertStmt = `INSERT INTO in_progress_role_command(guild, channel, user, role, emoji, stage) 
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT (user) DO UPDATE SET 
	channel = ?,
	role = ?,
	emoji = ?,
	stage = ?
	WHERE user = ?;`
	stmt, err := tx.Prepare(insertStmt)

	if err != nil {
		log.Fatal(err)
		return
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		command.Guild,
		command.Channel,
		command.User,
		command.Role,
		command.Emoji,
		command.Stage,
		command.Channel,
		command.Role,
		command.Emoji,
		command.Stage,
		command.User)

	if err != nil {
		log.Fatal(err)
		return
	}

	tx.Commit()

}

func (c *Client) SaveRoleCommand(msgID disgord.Snowflake, roleCommand botcommands.CompletedRoleCommand) {
	tx, err := c.client.Begin()

	if err != nil {
		log.Fatal(err)
		return
	}

	const insertStmt = `INSERT INTO role_message_command(guild, msg, role, emoji) 
	VALUES (?, ?, ?, ?);`
	stmt, err := tx.Prepare(insertStmt)

	if err != nil {
		log.Fatal(err)
		return
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		roleCommand.Guild, 
		msgID, 
		roleCommand.Role, 
		roleCommand.Emoji)

	if err != nil {
		log.Fatal(err)
		return
	}

	tx.Commit()
}

func (c *Client) IsUserUsingCommand(user *disgord.User) bool {
	const query = `SELECT * FROM in_progress_role_command WHERE user = ?;`

	rows, err := c.client.Query(query, user.ID)
	if err != nil {
		log.Fatal(err)
		return false
	}

	defer rows.Close()

	return rows.Next()
}

func (c *Client) GetCommandInProgress(user *disgord.User) *botcommands.CommandInProgress {
	const query = `SELECT * FROM in_progress_role_command WHERE user = ?`

	row := c.client.QueryRow(query, user.ID)
	if row.Err() != nil {
		log.Fatal(row.Err())
		return nil
	}

	commandInProgress := botcommands.CommandInProgress{}

	row.Scan(
		&commandInProgress.CommandInProgressID,
		&commandInProgress.Guild, 
		&commandInProgress.Channel, 
		&commandInProgress.User, 
		&commandInProgress.Role,
		&commandInProgress.Emoji,
		&commandInProgress.Stage)

	return &commandInProgress
}

func (c *Client) RemoveCommandProgress(user disgord.Snowflake) {
	const query = `DELETE FROM in_progress_role_command WHERE user = ?`

	result, err := c.client.Exec(query, user)

	if err != nil {
		log.Fatal(err)
	}

	if num, _ := result.RowsAffected(); num < 1 {
		log.Fatal("Error in DELETE QUERY. No rows removed.")
	}
}

func (c *Client) IsRoleCommandMessage(msg disgord.Snowflake, emoji disgord.Snowflake) bool {
	const query = `SELECT * FROM role_message_command WHERE msg = ? AND emoji = ?`

	rows, err := c.client.Query(query, msg, emoji)
	if err != nil {
		log.Fatal(err)
		return false
	}

	defer rows.Close()

	return rows.Next()
}

func (c *Client) GetRoleCommand(msg disgord.Snowflake) *botcommands.CompletedRoleCommand {
	const query = `SELECT role_message_command_pk, guild, role, emoji FROM role_message_command WHERE msg = ?`

	row := c.client.QueryRow(query, msg)
	if row.Err() != nil {
		log.Fatal(row.Err())
		return nil
	}

	completedRoleCommand := botcommands.CompletedRoleCommand{}

	row.Scan(
		&completedRoleCommand.CompletedRoleCommandID,
		&completedRoleCommand.Guild,
		&completedRoleCommand.Role,
		&completedRoleCommand.Emoji)

	return &completedRoleCommand
}

func (c *Client) GetFollowedUser(screenName string) []botcommands.TwitterFollowCommand {
	const query = `SELECT * FROM twitter_follow_command WHERE screen_name = ?;`

	rows, _ := c.client.Query(query, screenName)
	if rows.Err() != nil {
		log.Fatal(rows.Err())
		return []botcommands.TwitterFollowCommand{}
	}

	completedCommand := []botcommands.TwitterFollowCommand{}

	for rows.Next() {
		row := botcommands.TwitterFollowCommand{}
		rows.Scan(
			&row.TwitterFollowCommandID,
			&row.ScreenName,
			&row.Channel,
			&row.Guild)
		completedCommand = append(completedCommand, row)
	}

	return completedCommand
}

func (c *Client) SaveUserToFollow(twitterFollow botcommands.TwitterFollowCommand) {
	const query = `INSERT INTO twitter_follow_command(screen_name, channel, guild) VALUES (?, ?, ?);`

	tx, err := c.client.Begin()

	if err != nil {
		log.Fatal(err)
		return
	}

	stmt, err := tx.Prepare(query)

	if err != nil {
		log.Fatal(err)
		return
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		twitterFollow.ScreenName, 
		twitterFollow.Channel,
		twitterFollow.Guild)

	if err != nil {
		log.Fatal(err)
		return
	}
}

func (c *Client) DeleteFollowedUser(screenName string, guild disgord.Snowflake) {
	const query = `DELETE FROM twitter_follow_command WHERE screen_name = ? AND guild = ?;`

	result, err := c.client.Exec(query, screenName, guild)

	if err != nil {
		log.Fatal(err)
	}

	if num, _ := result.RowsAffected(); num < 1 {
		log.Fatal("Error in DELETE QUERY. No rows removed.")
	}
}

func (c *Client) GetAllFollowedUsersInServer(guild disgord.Snowflake) []botcommands.TwitterFollowCommand {
	const query = `SELECT * FROM twitter_follow_command WHERE guild = ?;`

	rows, _ := c.client.Query(query,  guild)
	if rows.Err() != nil {
		log.Fatal(rows.Err())
		return []botcommands.TwitterFollowCommand{}
	}

	completedCommand := []botcommands.TwitterFollowCommand{}

	for rows.Next() {
		row := botcommands.TwitterFollowCommand{}
		rows.Scan(
			&row.TwitterFollowCommandID,
			&row.ScreenName,
			&row.Channel,
			&row.Guild)
		completedCommand = append(completedCommand, row)
	}

	return completedCommand
}