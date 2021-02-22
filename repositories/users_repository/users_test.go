// build +integration

package users_repository_test

import (
	"database/sql"
	"discordbot/repositories/model"
	"discordbot/repositories/users_repository"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)


func initDB() *sql.DB {
	client, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")

	if err != nil {
		log.Fatal(err)
	}

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

	user := model.Users{
		UsersID: 1,
		DiscordUsersID: 1234,
		UserName: "test",
		IsAdmin: false,
	}

	db.Exec(`INSERT INTO users(users_id, discord_users_id, user_name) VALUES (?, ?, ?);`, user.UsersID, user.DiscordUsersID, user.UserName)

	result, _ := repo.GetUserByDiscordId(user.DiscordUsersID)

	if !reflect.DeepEqual(user, result) {
		t.Error("Error retrieving users. Mismatched results.")
	}
}

func TestSaveUser(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := users_repository.New(db)

	user := model.Users{
		UsersID: 1,
		DiscordUsersID: 1234,
		UserName: "test",
		IsAdmin: false,
	}

	err := repo.SaveUser(&user)

	if err != nil {
		t.Error(err)
		return
	}

	result, _ := repo.GetUserByDiscordId(user.DiscordUsersID)

	if !reflect.DeepEqual(user, result) {
		t.Error("Error retrieving users. Mismatched results.")
	}
}

func TestIsUserUsingCommand(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := users_repository.New(db)

	user := model.Users{
		UsersID: 1,
		DiscordUsersID: 1234,
		UserName: "test",
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