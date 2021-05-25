package repositories_test

import (
	"database/sql"
	"discordbot/commands"
	"discordbot/repositories"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() *sql.DB {
	client, _ := sql.Open("sqlite3", ":memory:?_foreign_keys=on")

	query, _ := ioutil.ReadFile("../dbscript.sql")

	if _, err := client.Exec(string(query)); err != nil {
		log.Fatal(err)
	}

	client.Exec(`INSERT INTO users(users_id, discord_users_id) VALUES (1234, 5678);`)

	return client
}

func TestSaveMangaNotification(t *testing.T) {
	db := initDB()
	defer db.Close()

	d := repositories.NewMangaNotificationRepository(db)

	manganotification := commands.MangaNotification{
		User:     1234,
		MangaURL: "website.com/manga",
		Guild:    1234,
		Channel:  5678,
		Role:     1357,
	}
	err := d.SaveMangaNotification(&manganotification)

	if err != nil {
		t.Error(err)
		return
	}

	row := db.QueryRow(`SELECT * FROM manga_notification WHERE manga_notification_id = 1`)
	result := commands.MangaNotification{}
	err = row.Scan(
		&result.MangaNotificationID,
		&result.User,
		&result.MangaURL,
		&result.Guild,
		&result.Channel,
		&result.Role,
	)

	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(manganotification, result) {
		t.Error("Mismatched structs found on save.")
	}
}

func TestGetAllStrawpolls(t *testing.T) {
	db := initDB()
	defer db.Close()

	mndb := repositories.NewMangaNotificationRepository(db)

	mn1 := commands.MangaNotification{
		MangaNotificationID: 1,
		User: 1234,
		MangaURL: "manga.com/manga",
		Guild: 1234,
		Channel: 5678,
		Role: 1357,
	}
	mn2 := commands.MangaNotification{
		MangaNotificationID: 2,
		User: 1234,
		MangaURL: "manga.com/webtoon",
		Guild: 1543,
		Channel: 345623,
		Role: 64352,
	}

	_, err := db.Exec(`INSERT INTO manga_notification(manga_url, author, guild, channel, role) VALUES  (?, ?, ?, ?, ?);`, &mn1.MangaURL, &mn1.User, &mn1.Guild, &mn1.Channel, &mn1.Role)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = db.Exec(`INSERT INTO manga_notification(manga_url, author, guild, channel, role) VALUES  (?, ?, ?, ?, ?);`, &mn2.MangaURL, &mn2.User, &mn2.Guild, &mn2.Channel, &mn2.Role)
	if err != nil {
		log.Println(err)
		return
	}

	rs, err := mndb.GetAllMangaNotifications()	

	if err != nil {
		t.Error(err)
		return
	}

	if len(rs) != 2 {
		t.Error("Wrong number of rows returned. Expected 2 received ", len(rs))
		return
	}

	if !reflect.DeepEqual(rs[0], mn1) {
		t.Error("Error with retrieving manga notification.")
	}

	if !reflect.DeepEqual(rs[1], mn2) {
		t.Error("Error with retrieving manga notification.")
	}
}