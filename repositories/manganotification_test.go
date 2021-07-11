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

func TestGetAllMangaNotifications(t *testing.T) {
	db := initDB()
	defer db.Close()

	mndb := repositories.NewMangaNotificationRepository(db)

	mn1 := commands.MangaNotification{
		MangaNotificationID: 1,
		User: 1234,
		Guild: 1234,
		Channel: 5678,
		Role: 1357,
	}
	mn2 := commands.MangaNotification{
		MangaNotificationID: 2,
		User: 1234,
		Guild: 1543,
		Channel: 345623,
		Role: 64352,
	}

	_, err := db.Exec(`INSERT INTO manga_notification(author, guild, channel, role) VALUES  ( ?, ?, ?, ?);`, &mn1.User, &mn1.Guild, &mn1.Channel, &mn1.Role)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = db.Exec(`INSERT INTO manga_notification(author, guild, channel, role) VALUES  ( ?, ?, ?, ?);`, &mn2.User, &mn2.Guild, &mn2.Channel, &mn2.Role)
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

func TestAddMangaLink(t *testing.T) {
	db := initDB()
	defer db.Close()

	mldb := repositories.NewMangaLinkRepository(db)
	mndb := repositories.NewMangaNotificationRepository(db)

	mn1 := commands.MangaNotification{
		MangaNotificationID: 1,
		User: 1234,
		Guild: 1234,
		Channel: 5678,
		Role: 1357,
	}

	_, err := db.Exec(`INSERT INTO manga_notification(author, guild, channel, role) VALUES  ( ?, ?, ?, ?);`, &mn1.User, &mn1.Guild, &mn1.Channel, &mn1.Role)
	if err != nil {
		log.Println(err)
		return
	}

	ml1 := commands.MangaLink{
		MangaLinkID: 1,
		MangaLink: "manga.com/manga",
		MangaNotifications: []commands.MangaNotification{mn1},
	}

	_, err = db.Exec(`INSERT INTO manga_links(manga_link) VALUES  (?);`, &ml1.MangaLink)
	if err != nil {
		log.Println(err)
		return
	}

	err = mndb.AddMangaLink(1,1)
	if err != nil {
		log.Println(err)
		return
	}

	rs, err := mldb.GetMangaLinkByLink("manga.com/manga")	

	if err != nil {
		t.Error(err)
		return
	}

	if len(rs.MangaNotifications) != 1 {
		t.Errorf("Error with saving manga link to notification.")
	}

	if rs.MangaNotifications[0].MangaNotificationID != 1 {
		t.Error("Error with saving manga link to notification.")
	}
}

func TestSaveMangaLink(t *testing.T) {
	db := initDB()
	defer db.Close()

	d := repositories.NewMangaLinkRepository(db)

	mangalink := commands.MangaLink{
		MangaLinkID: 1,
		MangaLink: "manga.com/manga",
	}
	err := d.SaveMangaLink(&mangalink)

	if err != nil {
		t.Error(err)
		return
	}

	row := db.QueryRow(`SELECT * FROM manga_links WHERE manga_link_id = 1`)
	result := commands.MangaLink{}
	err = row.Scan(
		&result.MangaLinkID,
		&result.MangaLink,
	)

	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(mangalink, result) {
		t.Error("Mismatched structs found on save.")
	}
}

func TestGetMangaLinkByLink(t *testing.T) {
	db := initDB()
	defer db.Close()

	mndb := repositories.NewMangaLinkRepository(db)

	mnn := commands.MangaNotification{
		MangaNotificationID: 1,
		Guild: 1,
		Channel: 2,
		Role: 3,
		User: 1234,
	}
	mn1 := commands.MangaLink{
		MangaLinkID: 1,
		MangaLink: "manga.com/manga",
		MangaNotifications: []commands.MangaNotification{mnn},
	}

	_, err := db.Exec(`INSERT INTO manga_notification(guild, channel, role, author) VALUES  (?, ?, ?, ?);`, &mnn.Guild, &mnn.Channel, &mnn.Role, &mnn.User)
	_, err = db.Exec(`INSERT INTO manga_links(manga_link) VALUES  (?);`, &mn1.MangaLink)
	_, err = db.Exec(`INSERT INTO manga_notification_links(manga_notification_id, manga_link_id) VALUES  (?, ?);`, &mnn.MangaNotificationID, &mn1.MangaLinkID)
	if err != nil {
		log.Println(err)
		return
	}

	rs, err := mndb.GetMangaLinkByLink("manga.com/manga")	

	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(rs, mn1) {
		t.Error("Error with retrieving manga link.")
	}
}

func TestGetAllMangaLinks(t *testing.T) {
	db := initDB()
	defer db.Close()

	mndb := repositories.NewMangaLinkRepository(db)

	mnn := commands.MangaNotification{
		MangaNotificationID: 1,
		Guild: 1,
		Channel: 2,
		Role: 3,
		User: 1234,
	}
	mn1 := commands.MangaLink{
		MangaLinkID: 1,
		MangaLink: "manga.com/manga",
		MangaNotifications: []commands.MangaNotification{mnn},
	}
	mnn2 := commands.MangaNotification{
		MangaNotificationID: 2,
		Guild: 3,
		Channel: 4,
		Role: 5,
		User: 1234,
	}
	mn2 := commands.MangaLink{
		MangaLinkID: 2,
		MangaLink: "also.com/manga",
		MangaNotifications: []commands.MangaNotification{mnn2},
	}

	_, err := db.Exec(`INSERT INTO manga_links(manga_link) VALUES  (?);`, &mn1.MangaLink)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = db.Exec(`INSERT INTO manga_notification(guild, channel, role, author) VALUES  (?, ?, ?, ?);`, &mnn.Guild, &mnn.Channel, &mnn.Role, &mnn.User)
	_, err = db.Exec(`INSERT INTO manga_links(manga_link) VALUES  (?);`, &mn2.MangaLink)
	_, err = db.Exec(`INSERT INTO manga_notification(guild, channel, role, author) VALUES  (?, ?, ?, ?);`, &mnn2.Guild, &mnn2.Channel, &mnn2.Role, &mnn2.User)
	_, err = db.Exec(`INSERT INTO manga_notification_links(manga_notification_id, manga_link_id) VALUES  (?, ?);`, &mnn.MangaNotificationID, &mn1.MangaLinkID)
	_, err = db.Exec(`INSERT INTO manga_notification_links(manga_notification_id, manga_link_id) VALUES  (?, ?);`, &mnn2.MangaNotificationID, &mn2.MangaLinkID)
	if err != nil {
		log.Println(err)
		return
	}

	rs, err := mndb.GetAllMangaLinks()	

	if err != nil {
		t.Error(err)
		return
	}

	if len(rs) != 2 {
		t.Error("Wrong number of rows returned. Expected 2 received ", len(rs))
		return
	}

	if !reflect.DeepEqual(rs[0], mn1) {
		t.Error("Error with retrieving manga link.")
	}

	if !reflect.DeepEqual(rs[1], mn2) {
		t.Error("Error with retrieving manga link.")
	}
}