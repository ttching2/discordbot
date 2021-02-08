// build +integration

package users_repository_test

import (
	"database/sql"
	"discordbot/repositories"
	"discordbot/repositories/users_repository"
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
	
	return client
}

func TestGetUserByDiscordId(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := users_repository.New(db)

	user := repositories.Users{
		UsersID: 1,
		DiscordUsersID: 1234,
		IsAdmin: false,
	}

	db.Exec(`INSERT INTO users(users_id, discord_users_id) VALUES (?, ?);`, user.UsersID, user.DiscordUsersID)

	result := repo.GetUserByDiscordId(user.DiscordUsersID)

	if !reflect.DeepEqual(user, result) {
		t.Error("Error retrieving users. Mismatched results.")
	}
}

func TestSaveUser(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := users_repository.New(db)

	user := repositories.Users{
		UsersID: 1,
		DiscordUsersID: 1234,
		IsAdmin: false,
	}

	err := repo.SaveUser(&user)

	if err != nil {
		t.Error(err)
		return
	}

	result := repo.GetUserByDiscordId(user.DiscordUsersID)

	if !reflect.DeepEqual(user, result) {
		t.Error("Error retrieving users. Mismatched results.")
	}
}

func TestIsUserUsingCommand(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := users_repository.New(db)

	user := repositories.Users{
		UsersID: 1,
		DiscordUsersID: 1234,
		IsAdmin: false,
	}

	err := repo.SaveUser(&user)

	if err != nil {
		t.Error(err)
		return
	}

	exists := repo.DoesUserExist(user.DiscordUsersID)

	if !exists {
		t.Error("User not found with in progress command.")
	}
}