// build +integration

package twitterfollow_test

import (
	"database/sql"
	"discordbot/databaseclient"
	"discordbot/databaseclient/twitterfollow"
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

	client.Exec(`INSERT INTO users(users_id, discord_users_id) VALUES (1, 1234);`)

	return client
}

func TestGetFollowedUser(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := twitterfollow.New(db)

	twitterFollow := databaseclient.TwitterFollowCommand{
		TwitterFollowCommandID: 1,
		User: 1,
		ScreenName: "watson",
		Channel: 1234,
		Guild: 567,
		ScreenNameID: "abs123",
	}
	db.Exec(`INSERT INTO twitter_follow_command(author, screen_name, channel, guild, screen_name_id) VALUES (?, ?, ?, ?, ?);`, 
	&twitterFollow.User, &twitterFollow.ScreenName, &twitterFollow.Channel, &twitterFollow.Guild, &twitterFollow.ScreenNameID)

	result := repo.GetFollowedUser(twitterFollow.ScreenName)

	if len(result) != 1 {
		t.Error("Wrong number of rows returned. Expected: 1, Got: ", len(result))
	}

	if !reflect.DeepEqual(twitterFollow, result[0]) {
		t.Error("Error retrieving twitter follow command. Mismatched results.")
	}
}

func TestSaveUserToFollow(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := twitterfollow.New(db)

	twitterFollow := databaseclient.TwitterFollowCommand{
		User: 1,
		ScreenName: "watson",
		Channel: 1234,
		Guild: 567,
		ScreenNameID: "abs123",
	}
	err := repo.SaveUserToFollow(&twitterFollow)
	
	if err != nil {
		t.Error("Error saving twitter follow: ", err)
		return
	}

	result := databaseclient.TwitterFollowCommand{}
	row := db.QueryRow(`SELECT * FROM twitter_follow_command WHERE twitter_follow_command_id = 1;`)
	err = row.Scan(
		&result.TwitterFollowCommandID,
		&result.User,
		&result.ScreenName,
		&result.Channel,
		&result.Guild,
		&result.ScreenNameID,
	)
	
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(twitterFollow, result) {
		t.Error("Mismatched structs found on save.")
	}
}

func TestDeleteFollowedUser(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := twitterfollow.New(db)

	twitterFollow := databaseclient.TwitterFollowCommand{
		User: 1,
		ScreenName: "watson",
		Channel: 1234,
		Guild: 567,
		ScreenNameID: "abs123",
	}
	db.Exec(`INSERT INTO twitter_follow_command(author, screen_name, channel, guild, screen_name_id) VALUES (?, ?, ?, ?, ?);`, 
	&twitterFollow.User, &twitterFollow.ScreenName, &twitterFollow.Channel, &twitterFollow.Guild, &twitterFollow.ScreenNameID)

	err := repo.DeleteFollowedUser(twitterFollow.ScreenName, twitterFollow.Guild)

	if err != nil {
		t.Error(err)
		return
	}

	result := repo.GetFollowedUser(twitterFollow.ScreenName)

	if len(result) !=0  {
		t.Error("Wrong number of rows returned. Expected: 0, Got: ", len(result))
	}

}

func TestGetAllFollowedUsersInServer(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := twitterfollow.New(db)

	twitterFollow1 := databaseclient.TwitterFollowCommand{
		TwitterFollowCommandID: 1,
		User: 1,
		ScreenName: "watson",
		Channel: 1234,
		Guild: 567,
		ScreenNameID: "abs123",
	}
	twitterFollow2 := databaseclient.TwitterFollowCommand{
		TwitterFollowCommandID: 2,
		User: 1,
		ScreenName: "gura",
		Channel: 12343,
		Guild: 567,
		ScreenNameID: "sdeg2312",
	}
	twitterFollow3 := databaseclient.TwitterFollowCommand{
		TwitterFollowCommandID: 3,
		User: 1,
		ScreenName: "me",
		Channel: 5423,
		Guild: 654,
		ScreenNameID: "oijr90234",
	}

	db.Exec(`INSERT INTO twitter_follow_command(author, screen_name, channel, guild, screen_name_id) VALUES (?, ?, ?, ?, ?);`, 
	&twitterFollow1.User, &twitterFollow1.ScreenName, &twitterFollow1.Channel, &twitterFollow1.Guild, &twitterFollow1.ScreenNameID)
	db.Exec(`INSERT INTO twitter_follow_command(author, screen_name, channel, guild, screen_name_id) VALUES (?, ?, ?, ?, ?);`, 
	&twitterFollow2.User, &twitterFollow2.ScreenName, &twitterFollow2.Channel, &twitterFollow2.Guild, &twitterFollow2.ScreenNameID)
	db.Exec(`INSERT INTO twitter_follow_command(author, screen_name, channel, guild, screen_name_id) VALUES (?, ?, ?, ?, ?);`, 
	&twitterFollow3.User, &twitterFollow3.ScreenName, &twitterFollow3.Channel, &twitterFollow3.Guild, &twitterFollow3.ScreenNameID)

	results := repo.GetAllFollowedUsersInServer(567)

	if len(results) != 2 {
		t.Error("Wrong number of rows returned. Expected: 2, Got: ", len(results))
		return
	}

	if !reflect.DeepEqual(twitterFollow1, results[0]) {
		t.Error("Mismatched structs found while getting all twitter follows in a server.")
		return 
	}
	if !reflect.DeepEqual(twitterFollow2, results[1]) {
		t.Error("Mismatched structs found while getting all twitter follows in a server.")
	}
}

func TestGetAllUniqueFollowedUsers(t *testing.T) {
	db := initDB()
	defer db.Close()

	repo := twitterfollow.New(db)

	twitterFollow1 := databaseclient.TwitterFollowCommand{
		TwitterFollowCommandID: 1,
		User: 1,
		ScreenName: "watson",
		Channel: 1234,
		Guild: 567,
		ScreenNameID: "abs123",
	}
	twitterFollow2 := databaseclient.TwitterFollowCommand{
		TwitterFollowCommandID: 2,
		User: 1,
		ScreenName: "gura",
		Channel: 12343,
		Guild: 567,
		ScreenNameID: "sdeg2312",
	}
	twitterFollow3 := databaseclient.TwitterFollowCommand{
		TwitterFollowCommandID: 3,
		User: 1,
		ScreenName: "gura",
		Channel: 5423,
		Guild: 654,
		ScreenNameID: "sdeg2312",
	}

	db.Exec(`INSERT INTO twitter_follow_command(author, screen_name, channel, guild, screen_name_id) VALUES (?, ?, ?, ?, ?);`, 
	&twitterFollow1.User, &twitterFollow1.ScreenName, &twitterFollow1.Channel, &twitterFollow1.Guild, &twitterFollow1.ScreenNameID)
	db.Exec(`INSERT INTO twitter_follow_command(author, screen_name, channel, guild, screen_name_id) VALUES (?, ?, ?, ?, ?);`, 
	&twitterFollow2.User, &twitterFollow2.ScreenName, &twitterFollow2.Channel, &twitterFollow2.Guild, &twitterFollow2.ScreenNameID)
	db.Exec(`INSERT INTO twitter_follow_command(author, screen_name, channel, guild, screen_name_id) VALUES (?, ?, ?, ?, ?);`, 
	&twitterFollow3.User, &twitterFollow3.ScreenName, &twitterFollow3.Channel, &twitterFollow3.Guild, &twitterFollow3.ScreenNameID)

	results := repo.GetAllUniqueFollowedUsers()

	if len(results) != 2 {
		t.Error("Wrong number of rows returned. Expected: 2, Got: ", len(results))
		return
	}

	if twitterFollow1.ScreenName != results[0].ScreenName {
		t.Error("Mismatched screen name found while retrieving all unique twitter follows.")
		return 
	}
	if twitterFollow2.ScreenName != results[1].ScreenName {
		t.Error("Mismatched screen name found while retrieving all unique twitter follows.")
	}
}