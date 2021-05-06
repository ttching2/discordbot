// build +integration

package tourneyrepo_test

import (
	"database/sql"
	"discordbot/repositories/model"
	"discordbot/repositories/tourneyrepo"
	"io/ioutil"
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() *sql.DB {
	client, _ := sql.Open("sqlite3", ":memory:?_foreign_keys=on")

	query, _ := ioutil.ReadFile("..\\..\\dbscript.sql")

	if _, err := client.Exec(string(query)); err != nil {
		log.Fatal(err)
	}

	client.Exec(`INSERT INTO users(users_id, discord_users_id, user_name) VALUES (1234, 5678, 'person');`)

	return client
}

func TestSaveNewTourney(t *testing.T) {
	db := initDB()
	repo := tourneyrepo.NewRepository(db)

	tourney := &model.Tournament{
		User:            1234,
		DiscordServerID: 123,
		ChallongeID:     "ABC",
		Participants:    []model.TournamentParticipant{{Name: "test", ChallongeID: 1}, {Name: "person", ChallongeID: 2}},
		Organizers:      []model.Users{{UsersID: 1234, DiscordUsersID: 5678, UserName: "person"}}}

	err := repo.SaveTourney(tourney)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	r, err := repo.GetTourneyByServer(123)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	if r.TournamentID != 1 {
		log.Println("Tournament not saved correctly.")
		t.FailNow()
	}
	if len(r.Participants) != 2 || r.Participants[0].ChallongeID == 0 {
		log.Println("Failed to save participants.")
		t.FailNow()
	}
	if len(r.Organizers) != 1 || r.Organizers[0].DiscordUsersID == 0 {
		log.Println("Failed to save organizers.")
		t.Fail()
	}
}

func TestUpdateTourney(t *testing.T) {
	db := initDB()
	repo := tourneyrepo.NewRepository(db)
	db.Exec("INSERT INTO tournament_participant (name, challonge_id) VALUES ('person', 1), ('test', 2), ('hey', 3);")
	tourney := &model.Tournament{User: 1234, DiscordServerID: 123, ChallongeID: "ABC", Participants: []model.TournamentParticipant{{TournamentParticipantID: 2, Name: "test", ChallongeID: 2}}}

	err := repo.SaveTourney(tourney)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	tourney.CurrentMatch = 567
	tourney.Participants = []model.TournamentParticipant{{TournamentParticipantID: 1, Name: "person", ChallongeID: 1}, {Name: "test1", ChallongeID: 123}, {Name: "second", ChallongeID: 567}}
	tourney.Organizers = []model.Users{{UsersID: 1234, DiscordUsersID: 5678, UserName: "person"}}
	err = repo.SaveTourney(tourney)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	r, err := repo.GetTourneyByServer(123)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	if r.TournamentID != 1 || r.CurrentMatch != 567 {
		log.Println("Tournament not updated correctly.")
		t.FailNow()
	}
	if len(r.Participants) != 3 || r.Participants[0].TournamentParticipantID != 1 {
		log.Println("Error updating Participants")
		t.FailNow()
	}
	if len(r.Organizers) != 1 || r.Organizers[0].UsersID != 1234 {
		log.Println("Error updating Organizers")
		t.FailNow()
	}
 	
}

func TestGetTourney(t *testing.T) {
	db := initDB()
	repo := tourneyrepo.NewRepository(db)

	db.Exec("INSERT INTO tournament (author,challonge_id,discord_server_id,current_match) VALUES (1234, 'abc', 567, 890);")
	db.Exec("INSERT INTO tournament_participant (name, challonge_id) VALUES ('person', 1), ('test', 2), ('hey', 3);")
	db.Exec("INSERT INTO tournament_participant_xref (tournament_id, tournament_participant_id) VALUES (1, 1), (1,2), (1,3);")
	db.Exec("INSERT INTO tournament_organizer_xref (tournament_id, users_id) VALUES (1, 1234);")

	r, err := repo.GetTourneyByServer(567)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	if r.ChallongeID != "abc" {
		log.Println("Tournament was fetched incorrectly.")
		t.FailNow()
	}
	if len(r.Participants) != 3 || r.Participants[0].ChallongeID == 0 {
		log.Println("Participants were fetched incorrectly.")
		t.FailNow()
	}
	if len(r.Organizers) != 1 || r.Organizers[0].UsersID == 0 {
		log.Println("Wrong number of organizers returned.")
		t.FailNow()
	}
}

func TestAddTourneyOrganizer(t *testing.T) {
	db := initDB()
	repo := tourneyrepo.NewRepository(db)

	tourney := &model.Tournament{User: 1234, DiscordServerID: 123, ChallongeID: "ABC"}

	err := repo.SaveTourney(tourney)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	repo.AddTourneyOrganizer(1234, 1)

	r, err := repo.GetTourneyByServer(123)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	if len(r.Organizers) != 1 || r.Organizers[0].UsersID == 0 {
		log.Println("Wrong number of organizers returned.")
		t.FailNow()
	}
}

func TestIsUserTourneyOrganizer(t *testing.T) {
	db := initDB()
	repo := tourneyrepo.NewRepository(db)

	tourney := &model.Tournament{User: 1234, DiscordServerID: 123, ChallongeID: "ABC", Organizers: []model.Users{{UsersID: 1234, DiscordUsersID: 5678, UserName: "person"}}}

	err := repo.SaveTourney(tourney)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	result, err := repo.IsUserTourneyOrganizer(1234, 1)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	if !result {
		log.Println("User is not organizer when they should be an organizer")
		t.Fail()
	}
}

func TestRemoveTournament(t *testing.T) {
	db := initDB()
	repo := tourneyrepo.NewRepository(db)

	tourney := &model.Tournament{User: 1234, DiscordServerID: 123, ChallongeID: "ABC", Organizers: []model.Users{{UsersID: 1234, DiscordUsersID: 5678, UserName: "person"}}}

	err := repo.SaveTourney(tourney)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	err = repo.RemoveTourney(123)

	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	r, _ := repo.GetTourneyByServer(123)

	if r.TournamentID != 0 {
		log.Println("Tournament not deleted correctly")
		t.Fail()
	}
}