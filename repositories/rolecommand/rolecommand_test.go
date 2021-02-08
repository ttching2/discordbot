//build +integration

package rolecommand_test

import (
	"database/sql"
	"discordbot/repositories"
	"discordbot/repositories/rolecommand"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)


func initDB() *sql.DB {
	client, _ := sql.Open("sqlite3", ":memory:?_foreign_keys=on")

	query, _ := ioutil.ReadFile("..\\..\\dbscript.sql")

	if _, err := client.Exec(string(query)); err != nil {
		log.Fatal(err)
	}

	client.Exec(`INSERT INTO users(users_id, discord_users_id) VALUES (1234, 5678);`)

	return client
}

func TestSaveCommandInProgress(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := rolecommand.New(db)

	commandInProgress := repositories.CommandInProgress{
		User: 13456,
		Guild: 567,
		OriginChannel: 1234,
		Role: 1234364,
		Emoji: 23145,
		Stage: 1,
	}
	err := repo.SaveCommandInProgress(&commandInProgress)
	
	if err != nil {
		t.Error("Error saving in progress role command: ", err)
		return
	}

	result := repositories.CommandInProgress{}
	row := db.QueryRow(`SELECT * FROM in_progress_role_command WHERE in_progress_role_command_pk = 1;`)
	err = row.Scan(
		&result.CommandInProgressID,
		&result.Guild,
		&result.OriginChannel,
		&result.TargetChannel,
		&result.User,
		&result.Role,
		&result.Emoji,
		&result.Stage,
	)
	
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(commandInProgress, result) {
		t.Error("Mismatched structs found on save.")
	}
}
func TestSaveRoleCommand(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := rolecommand.New(db)

	roleCommand := repositories.RoleCommand{
		User: 1234,
		Guild: 567,
		Role: 1234364,
		Emoji: 23145,
		Message: 253435,
	}
	err := repo.SaveRoleCommand(&roleCommand)
	
	if err != nil {
		t.Error("Error saving completed role command: ", err)
		return
	}

	result := repositories.RoleCommand{}
	row := db.QueryRow(`SELECT * FROM role_message_command WHERE role_message_command_pk = 1;`)
	err = row.Scan(
		&result.RoleCommandID,
		&result.User,
		&result.Guild,
		&result.Message,
		&result.Role,
		&result.Emoji,
	)
	
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(roleCommand, result) {
		t.Error("Mismatched structs found on save.")
	}
}
func TestIsUserUsingCommand(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := rolecommand.New(db)

	commandInProgress := repositories.CommandInProgress{
		User: 13456,
		Guild: 567,
		OriginChannel: 1234,
		Role: 1234364,
		Emoji: 23145,
		Stage: 1,
	}
	repo.SaveCommandInProgress(&commandInProgress)
	using := repo.IsUserUsingCommand(commandInProgress.User, commandInProgress.OriginChannel)

	if !using {
		t.Error("User not found with in progress command.")
	}
}
func TestGetCommandInProgress(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := rolecommand.New(db)

	commandInProgress := repositories.CommandInProgress{
		User: 13456,
		Guild: 567,
		OriginChannel: 1234,
		Role: 1234364,
		Emoji: 23145,
		Stage: 1,
	}
	repo.SaveCommandInProgress(&commandInProgress)

	result := repo.GetCommandInProgress(commandInProgress.User, commandInProgress.OriginChannel)

	if !reflect.DeepEqual(result, commandInProgress) {
		t.Error("Mismatched structs found while getting command in progress.")
		return 
	}
}
func TestRemoveCommandProgress(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := rolecommand.New(db)

	commandInProgress := repositories.CommandInProgress{
		User: 13456,
		Guild: 567,
		OriginChannel: 1234,
		Role: 1234364,
		Emoji: 23145,
		Stage: 1,
	}
	repo.SaveCommandInProgress(&commandInProgress)
	err := repo.RemoveCommandProgress(commandInProgress.User, commandInProgress.OriginChannel)

	if err != nil {
		t.Error("Error deleting in progress role command. ", err)
	}
}
func TestIsRoleCommandMessage(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := rolecommand.New(db)

	roleCommand := repositories.RoleCommand{
		User: 1234,
		Guild: 567,
		Role: 1234364,
		Emoji: 23145,
		Message: 253435,
	}
	err := repo.SaveRoleCommand(&roleCommand)
	if err != nil {
		t.Error(err)
		return
	}
	isRoleCommand := repo.IsRoleCommandMessage(roleCommand.Message, roleCommand.Emoji)

	if !isRoleCommand {
		t.Error("Role command message not found when it should exist.")
	}
}
func TestGetRoleCommand(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := rolecommand.New(db)

	roleCommand := repositories.RoleCommand{
		User: 1234,
		Guild: 567,
		Role: 1234364,
		Emoji: 23145,
		Message: 253435,
	}
	err := repo.SaveRoleCommand(&roleCommand)
	if err != nil {
		t.Error(err)
		return
	}
	result := repo.GetRoleCommand(roleCommand.Message)

	if !reflect.DeepEqual(roleCommand, result) {
		t.Error("Error retrieving role command. Mismatched commands found.")
	}
}
func TestRemoveRoleReactCommand(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := rolecommand.New(db)

	roleCommand := repositories.RoleCommand{
		User: 1234,
		Guild: 567,
		Role: 1234364,
		Emoji: 23145,
		Message: 253435,
	}
	repo.SaveRoleCommand(&roleCommand)
	err := repo.RemoveRoleReactCommand(roleCommand.Message)

	if err != nil {
		t.Error(err)
		return
	}

	result := repo.IsRoleCommandMessage(roleCommand.Message, roleCommand.Emoji)

	if result {
		t.Error("Role command not deleted.")
	}
}