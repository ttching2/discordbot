package rolecommand

import (
	"database/sql"
	"discordbot/commands"
	"errors"
)

type roleCommandRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *roleCommandRepository {
	return &roleCommandRepository{
		db: db,
	}
}

func (r *roleCommandRepository) SaveCommandInProgress(command *commands.CommandInProgress) error {
	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	const insertStmt = `REPLACE INTO in_progress_role_command(guild, origin_channel, target_channel, user, role, emoji, stage) 
	VALUES (?, ?, ?, ?, ?, ?, ?);`
	stmt, err := tx.Prepare(insertStmt)

	if err != nil {
		return err
	}

	defer stmt.Close()

	res, err := stmt.Exec(
		command.Guild,
		command.OriginChannel,
		command.TargetChannel,
		command.User,
		command.Role,
		command.Emoji,
		command.Stage)

	if err != nil {
		return err
	}

	ID, _ := res.LastInsertId()
	if err != nil {
		return err
	}
	command.CommandInProgressID = ID

	tx.Commit()

	return nil
}

func (r *roleCommandRepository) SaveRoleCommand(roleCommand *commands.RoleCommand) error {
	const query = `INSERT INTO role_message_command(author, guild, msg, role, emoji) VALUES (?, ?, ?, ?, ?);`

	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	result, err := stmt.Exec(
		roleCommand.User,
		roleCommand.Guild,
		roleCommand.Message,
		roleCommand.Role,
		roleCommand.Emoji)

	if err != nil {
		return err
	}

	ID, _ := result.LastInsertId()
	if err != nil {
		return err
	}
	tx.Commit()

	roleCommand.RoleCommandID = ID

	return nil
}

func (r *roleCommandRepository) IsUserUsingCommand(user commands.Snowflake, channel commands.Snowflake) (bool, error) {
	const query = `SELECT * FROM in_progress_role_command WHERE user = ? AND origin_channel = ?;`

	rows, err := r.db.Query(query, user, channel)
	if err != nil {
		return false, err
	}

	defer rows.Close()

	return rows.Next(), nil
}

func (r *roleCommandRepository) GetCommandInProgress(user commands.Snowflake, channel commands.Snowflake) (commands.CommandInProgress, error) {
	const query = `SELECT * FROM in_progress_role_command WHERE user = ? AND origin_channel = ?;`

	row := r.db.QueryRow(query, user, channel)
	if row.Err() != nil {
		return commands.CommandInProgress{}, row.Err()
	}

	commandInProgress := commands.CommandInProgress{}

	row.Scan(
		&commandInProgress.CommandInProgressID,
		&commandInProgress.Guild,
		&commandInProgress.OriginChannel,
		&commandInProgress.TargetChannel,
		&commandInProgress.User,
		&commandInProgress.Role,
		&commandInProgress.Emoji,
		&commandInProgress.Stage)

	return commandInProgress, nil
}

func (r *roleCommandRepository) RemoveCommandProgress(user commands.Snowflake, channel commands.Snowflake) error {
	const query = `DELETE FROM in_progress_role_command WHERE user = ? AND origin_channel = ?;`

	result, err := r.db.Exec(query, user, channel)

	if err != nil {
		return err
	}

	if num, _ := result.RowsAffected(); num < 1 {
		return errors.New("no rows deleted")
	}
	return nil
}

func (r *roleCommandRepository) IsRoleCommandMessage(msg commands.Snowflake, emoji commands.Snowflake) (bool, error) {
	const query = `SELECT * FROM role_message_command WHERE msg = ? AND emoji = ?;`

	rows, err := r.db.Query(query, msg, emoji)
	if err != nil {
		return false, err
	}

	defer rows.Close()

	return rows.Next(), nil
}

func (r *roleCommandRepository) GetRoleCommand(msg commands.Snowflake) (commands.RoleCommand, error) {
	const query = `SELECT * FROM role_message_command WHERE msg = ?;`

	row := r.db.QueryRow(query, msg)
	if row.Err() != nil {
		return commands.RoleCommand{}, row.Err()
	}

	roleCommand := commands.RoleCommand{}

	err := row.Scan(
		&roleCommand.RoleCommandID,
		&roleCommand.User,
		&roleCommand.Guild,
		&roleCommand.Message,
		&roleCommand.Role,
		&roleCommand.Emoji)

	if err != nil {
		return commands.RoleCommand{}, err
	}

	return roleCommand, nil
}

func (r *roleCommandRepository) RemoveRoleReactCommand(msg commands.Snowflake) error {
	const query = `DELETE FROM role_message_command WHERE msg = ?;`

	result, err := r.db.Exec(query, msg)

	if err != nil {
		return err
	}

	if num, _ := result.RowsAffected(); num < 1 {
		return errors.New("no rows deleted")
	}
	return nil
}
