package rolecommand

import (
	"database/sql"
	"discordbot/repositories"
	"errors"
	"log"
)


type RoleCommandRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *RoleCommandRepository {
	return &RoleCommandRepository{
		db: db,
	}
}

func (r *RoleCommandRepository) SaveCommandInProgress(command *repositories.CommandInProgress) error {
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

func (r *RoleCommandRepository) SaveRoleCommand(roleCommand *repositories.RoleCommand) error {
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

	result , err := stmt.Exec(
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

func (r *RoleCommandRepository) IsUserUsingCommand(user repositories.Snowflake, channel repositories.Snowflake) bool {
	const query = `SELECT * FROM in_progress_role_command WHERE user = ? AND origin_channel = ?;`

	rows, err := r.db.Query(query, user, channel)
	if err != nil {
		log.Println(err)
		return false
	}

	defer rows.Close()

	return rows.Next()
}

func (r *RoleCommandRepository) GetCommandInProgress(user repositories.Snowflake, channel repositories.Snowflake) repositories.CommandInProgress {
	const query = `SELECT * FROM in_progress_role_command WHERE user = ? AND origin_channel = ?;`

	row := r.db.QueryRow(query, user, channel)
	if row.Err() != nil {
		log.Println(row.Err())
		return repositories.CommandInProgress{}
	}

	commandInProgress := repositories.CommandInProgress{}

	row.Scan(
		&commandInProgress.CommandInProgressID,
		&commandInProgress.Guild, 
		&commandInProgress.OriginChannel,
		&commandInProgress.TargetChannel,
		&commandInProgress.User, 
		&commandInProgress.Role,
		&commandInProgress.Emoji,
		&commandInProgress.Stage)

	return commandInProgress
}

func (r *RoleCommandRepository) RemoveCommandProgress(user repositories.Snowflake, channel repositories.Snowflake) error {
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

func (r *RoleCommandRepository) IsRoleCommandMessage(msg repositories.Snowflake, emoji repositories.Snowflake) bool {
	const query = `SELECT * FROM role_message_command WHERE msg = ? AND emoji = ?;`

	rows, err := r.db.Query(query, msg, emoji)
	if err != nil {
		log.Println(err)
		return false
	}

	defer rows.Close()

	return rows.Next()
}

func (r *RoleCommandRepository) GetRoleCommand(msg repositories.Snowflake) repositories.RoleCommand {
	const query = `SELECT * FROM role_message_command WHERE msg = ?;`

	row := r.db.QueryRow(query, msg)
	if row.Err() != nil {
		log.Println(row.Err())
		return repositories.RoleCommand{}
	}

	roleCommand := repositories.RoleCommand{}

	err := row.Scan(
		&roleCommand.RoleCommandID,
		&roleCommand.User,
		&roleCommand.Guild,
		&roleCommand.Message,
		&roleCommand.Role,
		&roleCommand.Emoji)

	if err != nil {
		log.Println(err)
		return repositories.RoleCommand{}
	}

	return roleCommand
}

func (r *RoleCommandRepository) RemoveRoleReactCommand(msg repositories.Snowflake) error {
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