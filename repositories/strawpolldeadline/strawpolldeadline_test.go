// build +integration

package strawpolldeadline_test

import (
	"database/sql"
	"discordbot/repositories/model"
	"discordbot/repositories/strawpolldeadline"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() *sql.DB {
	client, _ := sql.Open("sqlite3", ":memory:?_foreign_keys=on")

	query, _ := ioutil.ReadFile("../../dbscript.sql")

	if _, err := client.Exec(string(query)); err != nil {
		log.Fatal(err)
	}

	client.Exec(`INSERT INTO users(users_id, discord_users_id) VALUES (1234, 5678);`)

	return client
}

func TestSaveStrawpollDeadline(t *testing.T) {
	db := initDB()
	defer db.Close()

	spDB := strawpolldeadline.New(db)

	strawpolldeadline := model.StrawpollDeadline{
		User: 1234,
		StrawpollID: "abc",
		Guild: 1234,
		Channel: 5678,
		Role: 1357,
	}
	err := spDB.SaveStrawpollDeadline(&strawpolldeadline)

	if err != nil {
		t.Error(err)
		return
	}

	row := db.QueryRow(`SELECT * FROM strawpoll_deadline WHERE strawpoll_deadline_id = 1`)
	result := model.StrawpollDeadline{}
	err = row.Scan(
		&result.StrawpollDeadlineID,
		&result.User,
		&result.StrawpollID,
		&result.Guild,
		&result.Channel,
		&result.Role,
	)

	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(strawpolldeadline, result) {
		t.Error("Mismatched structs found on save.")
	}
}

func TestSaveWithNonExistantUser(t *testing.T) {
	db := initDB()
	defer db.Close()

	spDB := strawpolldeadline.New(db)

	strawpolldeadline := model.StrawpollDeadline{
		User: 2,
		StrawpollID: "abc",
		Guild: 1234,
		Channel: 5678,
		Role: 1357,
	}
	
	err := spDB.SaveStrawpollDeadline(&strawpolldeadline)

	if err == nil {
		t.Error("No error found on failed save.")
	}
}

func TestGetAllStrawpolls(t *testing.T) {
	db := initDB()
	defer db.Close()

	spDB := strawpolldeadline.New(db)

	s1 := model.StrawpollDeadline{
		StrawpollDeadlineID: 1,
		User: 1234,
		StrawpollID: "abc",
		Guild: 1234,
		Channel: 5678,
		Role: 1357,
	}
	s2 := model.StrawpollDeadline{
		StrawpollDeadlineID: 2,
		User: 1234,
		StrawpollID: "absd",
		Guild: 1543,
		Channel: 345623,
		Role: 64352,
	}

	_, err := db.Exec(`INSERT INTO strawpoll_deadline(strawpoll_id, author, guild, channel, role) VALUES  (?, ?, ?, ?, ?);`, &s1.StrawpollID, &s1.User, &s1.Guild, &s1.Channel, &s1.Role)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = db.Exec(`INSERT INTO strawpoll_deadline(strawpoll_id, author, guild, channel, role) VALUES  (?, ?, ?, ?, ?);`, &s2.StrawpollID, &s2.User, &s2.Guild, &s2.Channel, &s2.Role)
	if err != nil {
		log.Println(err)
		return
	}

	rs, err := spDB.GetAllStrawpollDeadlines()	

	if err != nil {
		t.Error(err)
		return
	}

	if len(rs) != 2 {
		t.Error("Wrong number of rows returned. Expected 2 received ", len(rs))
		return
	}

	if !reflect.DeepEqual(rs[0], s1) {
		t.Error("Error with retrieving strawpoll deadline.")
	}

	if !reflect.DeepEqual(rs[1], s2) {
		t.Error("Error with retrieving strawpoll deadline.")
	}
}

func TestDeleteStrawpollDeadline(t *testing.T) {
	db := initDB()
	defer db.Close()

	spDB := strawpolldeadline.New(db)

	s1 := model.StrawpollDeadline{
		StrawpollDeadlineID: 1,
		User: 1234,
		StrawpollID: "abc",
		Guild: 1234,
		Channel: 5678,
		Role: 1357,
	}

	err := spDB.SaveStrawpollDeadline(&s1)
	if err != nil {
		log.Println(err)
		return
	}

	spDB.DeleteStrawpollDeadlineByID(s1.StrawpollDeadlineID)

	row := db.QueryRow(`SELECT * FROM strawpoll_deadline WHERE strawpoll_deadline_id = 1`)
	result := model.StrawpollDeadline{}
	err = row.Scan(
		&result.StrawpollDeadlineID,
		&result.User,
		&result.StrawpollID,
		&result.Guild,
		&result.Channel,
		&result.Role,
	)

	if err == nil {
		t.Error("Error in deleting strawpoll deadline.")
		return
	}

	if err.Error() != "sql: no rows in result set" {
		t.Error("Error in deleting strawpoll deadline. ", err)
		return
	}
}