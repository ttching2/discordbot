package botcommands_test

import (
	"discordbot/botcommands"
	"discordbot/challonge"
	"discordbot/repositories/model"
	"log"
	"testing"

	"github.com/andersfylling/disgord"
	_ "github.com/mattn/go-sqlite3"
)

type onMessageCreateCommand interface {
	ExecuteMessageCreateCommand()
}

type mockTourneyDB struct {
	tourneys   map[model.Snowflake]*model.Tournament
	organizers map[int64]int64
}

func (r *mockTourneyDB) SaveTourney(t *model.Tournament) error {
	r.tourneys[t.DiscordServerID] = t
	return nil
}

func (r *mockTourneyDB) GetTourneyByServer(ID model.Snowflake) (model.Tournament, error) {
	return *r.tourneys[ID], nil
}

func (r *mockTourneyDB) AddTourneyOrganizer(userID int64, tourneyID int64) error {
	if r.organizers == nil {
		r.organizers = make(map[int64]int64)
	}
	r.organizers[userID] = tourneyID
	return nil
}

func (r *mockTourneyDB) IsUserTourneyOrganizer(ID int64, tournamentID int64) (bool, error) {
	_, ok := r.organizers[ID]
	return ok, nil
}

func (r *mockTourneyDB) HasMatchInProgress(ID int64) (bool, error) {
	return false, nil
}

func (r *mockTourneyDB) RemoveTourney(discordServerID model.Snowflake) error {
	delete(r.tourneys, discordServerID)
	return nil
}

type mockSession struct {
	message          string
	reactedMessageID model.Snowflake
}

func (s *mockSession) SendMessage(channel model.Snowflake, m *disgord.CreateMessageParams) {
	s.message = m.Content
}

func (s *mockSession) ReactToMessage(msg model.Snowflake, channel model.Snowflake, emoji interface{}) {
	s.reactedMessageID = msg
}

func (s *mockSession) getReactedMessage() model.Snowflake {
	return s.reactedMessageID
}

type mockChallongeClient struct {
	id           string
	participants []challonge.Participant
	matches      []challonge.Match
	WinnerID     int
	query        challonge.MatchQueryParams
}

func (c *mockChallongeClient) GetParticipants(tourneyID string) []challonge.Participant {
	if tourneyID != c.id {
		return nil
	}
	return c.participants
}

func (c *mockChallongeClient) GetMatches(tourneyID string) []challonge.Match {
	if tourneyID != c.id {
		return nil
	}
	return c.matches
}

func (c *mockChallongeClient) GetMatch(tourneyID string, matchID int) challonge.Match {
	if tourneyID != c.id {
		return challonge.Match{}
	}

	for _, m := range c.matches {
		if m.ID == matchID {
			return m
		}
	}
	return challonge.Match{}
}

func (c *mockChallongeClient) UpdateMatch(tourneyID string, matchID int, params challonge.MatchQueryParams) {
	if tourneyID != c.id {
		return
	}
	c.query = params
}

func (c *mockChallongeClient) setTourneyID(tourneyID string) {
	c.id = tourneyID
}

func (c *mockChallongeClient) addParticipant(p challonge.Participant) {
	c.participants = append(c.participants, p)
}

func (c *mockChallongeClient) addMatches(m ...challonge.Match) {
	c.matches = append(c.matches, m...)
}

func TestCreateTourneyFromChallongeLink(t *testing.T) {
	//Given: A link for a tourney from challonge.com with id testlink
	tourneyLink := "https://www.challonge.com/testlink"
	challongeClient := mockChallongeClient{}
	challongeClient.setTourneyID("testlink")
	//And: 2 participants
	p1 := challonge.Participant{ID: 432, Name: "test"}
	p2 := challonge.Participant{ID: 6543, Name: "Guy"}
	challongeClient.addParticipant(p1)
	challongeClient.addParticipant(p2)
	//And: A tourney command
	msg := disgord.MessageCreate{Message: &disgord.Message{
		Content: tourneyLink,
		GuildID: 123,
	}}
	user := model.Users{UsersID: 1, DiscordUsersID: 1}
	repo := &mockTourneyDB{tourneys: make(map[model.Snowflake]*model.Tournament)}
	s := mockSession{}
	factory := botcommands.NewTourneyCommandRequestFactory(&s, repo, &challongeClient)
	c := factory.CreateRequest(&msg, &user)
	//When: The command is executed
	c.(onMessageCreateCommand).ExecuteMessageCreateCommand()

	//Then: A tourney entry is made in the database
	to, err := repo.GetTourneyByServer(msg.Message.GuildID)

	if err != nil {
		t.Fail()
	}
	if to.DiscordServerID == 0 {
		log.Println("Failed to save tourney properly")
		t.Fail()
	}
	//And: The challonge id is correct
	if to.ChallongeID != "testlink" {
		log.Println("Empty challonge id")
		t.Fail()
	}
	//And: Participants were added
	if len(to.Participants) != 2 {
		log.Println("Wrong number of participants saved.")
		t.Fail()
	}
}

func TestRunCommandsWithoutTourneyStart(t *testing.T) {
	//Given: A command involving a tourney
	msg := disgord.MessageCreate{Message: &disgord.Message{
		Content: "testUser",
		GuildID: 123,
	}}
	challongeClient := mockChallongeClient{}
	user := model.Users{UsersID: 1, DiscordUsersID: 1}
	repo := &mockTourneyDB{tourneys: make(map[model.Snowflake]*model.Tournament)}
	s := mockSession{}
	factory := botcommands.NewTourneyCommandRequestFactory(&s, repo, &challongeClient)

	c := factory.CreateAddOrganizerCommand(&msg, &user)

	//When: The command is executed
	c.(onMessageCreateCommand).ExecuteMessageCreateCommand()

	//Then: The command does not complete and an error messasge is sent
	if s.message == "" {
		t.Fail()
	}
}

func TestAddTourneyOrganizer(t *testing.T) {
	//Given: A request to add a tourney organizer
	msg := disgord.MessageCreate{Message: &disgord.Message{
		Content: "testUser",
		GuildID: 123,
	}}
	challongeClient := mockChallongeClient{}
	user := model.Users{UsersID: 1, DiscordUsersID: 1}
	repo := &mockTourneyDB{tourneys: make(map[model.Snowflake]*model.Tournament)}
	repo.SaveTourney(&model.Tournament{
		TournamentID:    1,
		User:            1,
		DiscordServerID: 123,
	})
	s := mockSession{}
	factory := botcommands.NewTourneyCommandRequestFactory(&s, repo, &challongeClient)

	c := factory.CreateAddOrganizerCommand(&msg, &user)
	//When: The command is executed
	c.(onMessageCreateCommand).ExecuteMessageCreateCommand()

	//Then: The user is found as a tourney organizer
	b, _ := repo.IsUserTourneyOrganizer(user.UsersID, 1)

	if !b {
		t.Fail()
	}
}

func TestNextLosersMatch(t *testing.T) {
	//Given: A challonge bracket with a losers match
	msg := disgord.MessageCreate{Message: &disgord.Message{
		Content: "",
		GuildID: 123,
	}}

	cclient := mockChallongeClient{}
	wm1 := challonge.Match{ID: 100, Round: 1, Player1ID: 123, Player2ID: 234}
	wm2 := challonge.Match{ID: 200, Round: 1, Player1ID: 654, Player2ID: 789}
	lm1 := challonge.Match{ID: 300, Round: -1, Player1ID: 542, Player2ID: 987}
	cclient.addMatches(wm1, wm2, lm1)
	cclient.setTourneyID("test")

	user := model.Users{UsersID: 1, DiscordUsersID: 1}
	repo := &mockTourneyDB{tourneys: make(map[model.Snowflake]*model.Tournament)}
	//And: A tourney saved with matching participants
	repo.SaveTourney(&model.Tournament{
		TournamentID:    1,
		ChallongeID:     "test",
		User:            1,
		DiscordServerID: 123,
		Participants: []model.TournamentParticipant{
			{Name: "user1", ChallongeID: 542},
			{Name: "test user", ChallongeID: 987},
			{Name: "winner", ChallongeID: 123},
			{Name: "loser", ChallongeID: 234},
			{Name: "tse", ChallongeID: 654},
			{Name: "noit", ChallongeID: 789},
		},
	})
	s := mockSession{}
	factory := botcommands.NewTourneyCommandRequestFactory(&s, repo, &cclient)

	c := factory.CreateNextLosersCommnad(&msg, &user)

	//When: The command is executed
	c.(onMessageCreateCommand).ExecuteMessageCreateCommand()

	r, _ := repo.GetTourneyByServer(123)

	//Then: The correct match message is sent
	if s.message != "user1 vs test user" {
		log.Println("Message not sent")
		t.Fail()
	}
	if r.CurrentMatch != 300 {
		log.Println("Next losers match not set correctly")
		t.Fail()
	}

}

func TestNextLosersMatchMultipleLosers(t *testing.T) {
	//Given: A challonge bracket with 3 losers match and 2 waiting for opponents
	msg := disgord.MessageCreate{Message: &disgord.Message{
		Content: "",
		GuildID: 123,
	}}

	cclient := mockChallongeClient{}
	wm1 := challonge.Match{ID: 1, Round: -1, Player1ID: 123}
	wm2 := challonge.Match{ID: 2, Round: -1, Player1ID: 654}
	lm1 := challonge.Match{ID: 3, Round: -1, Player1ID: 542, Player2ID: 987}
	cclient.addMatches(wm1, wm2, lm1)
	cclient.setTourneyID("test")

	user := model.Users{UsersID: 1, DiscordUsersID: 1}
	repo := &mockTourneyDB{tourneys: make(map[model.Snowflake]*model.Tournament)}
	//And: A tourney saved with matching participants
	repo.SaveTourney(&model.Tournament{
		TournamentID:    1,
		ChallongeID:     "test",
		User:            1,
		DiscordServerID: 123,
		Participants: []model.TournamentParticipant{
			{Name: "user1", ChallongeID: 542},
			{Name: "test user", ChallongeID: 987},
			{Name: "winner", ChallongeID: 123},
			{Name: "loser", ChallongeID: 234},
			{Name: "tse", ChallongeID: 654},
			{Name: "noit", ChallongeID: 789},
		},
	})
	s := mockSession{}
	factory := botcommands.NewTourneyCommandRequestFactory(&s, repo, &cclient)

	c := factory.CreateNextLosersCommnad(&msg, &user)

	//When: The command is executed
	c.(onMessageCreateCommand).ExecuteMessageCreateCommand()

	tour := repo.tourneys[123]
	//Then: The correct match message is sent
	if s.message != "user1 vs test user" {
		log.Println("Message not sent")
		t.Fail()
	}
	//And: The losers match is set as the current match
	if tour.CurrentMatch != 3 {
		t.Fail()
		log.Println("Current Losers match not set.")
	}
}

func TestNextLosersNoMatchReady(t *testing.T) {
	//Given: A challonge bracket with 3 losers match and 2 waiting for opponents
	msg := disgord.MessageCreate{Message: &disgord.Message{
		Content: "",
		GuildID: 123,
	}}

	cclient := mockChallongeClient{}
	wm1 := challonge.Match{Round: -1, Player1ID: 123}
	wm2 := challonge.Match{Round: -1, Player1ID: 654}
	lm1 := challonge.Match{Round: -1, Player1ID: 542}
	cclient.addMatches(wm1, wm2, lm1)
	cclient.setTourneyID("test")

	user := model.Users{UsersID: 1, DiscordUsersID: 1}
	repo := &mockTourneyDB{tourneys: make(map[model.Snowflake]*model.Tournament)}
	//And: A tourney saved with matching participants
	repo.SaveTourney(&model.Tournament{
		TournamentID:    1,
		ChallongeID:     "test",
		User:            1,
		DiscordServerID: 123,
		Participants: []model.TournamentParticipant{
			{Name: "user1", ChallongeID: 542},
			{Name: "test user", ChallongeID: 987},
			{Name: "winner", ChallongeID: 123},
			{Name: "loser", ChallongeID: 234},
			{Name: "tse", ChallongeID: 654},
			{Name: "noit", ChallongeID: 789},
		},
	})
	s := mockSession{}
	factory := botcommands.NewTourneyCommandRequestFactory(&s, repo, &cclient)

	c := factory.CreateNextLosersCommnad(&msg, &user)

	//When: The command is executed
	c.(onMessageCreateCommand).ExecuteMessageCreateCommand()

	//Then: The correct match message is sent
	if s.message != "No losers match is ready to be played yet." {
		log.Println("Message not sent")
		t.Fail()
	}
}

func TestWinCommand(t *testing.T) {
	//Given: A challonge bracket with 3 losers match and 2 waiting for opponents
	msg := disgord.MessageCreate{Message: &disgord.Message{
		ID:      542,
		Content: "winner",
		GuildID: 123,
	}}

	cclient := mockChallongeClient{}
	wm1 := challonge.Match{ID: 1, Round: -1, Player1ID: 123, Player2ID: 987}
	wm2 := challonge.Match{ID: 2, Round: -1, Player1ID: 654}
	lm1 := challonge.Match{ID: 3, Round: -1, Player1ID: 542}
	cclient.addMatches(wm1, wm2, lm1)
	cclient.setTourneyID("test")

	user := model.Users{UsersID: 1, DiscordUsersID: 1}
	repo := &mockTourneyDB{tourneys: make(map[model.Snowflake]*model.Tournament)}
	//And: A tourney saved with matching participants
	repo.SaveTourney(&model.Tournament{
		TournamentID:    1,
		ChallongeID:     "test",
		User:            1,
		DiscordServerID: 123,
		Participants: []model.TournamentParticipant{
			{Name: "user1", ChallongeID: 542},
			{Name: "test user", ChallongeID: 987},
			{Name: "winner", ChallongeID: 123},
			{Name: "loser", ChallongeID: 234},
			{Name: "tse", ChallongeID: 654},
			{Name: "noit", ChallongeID: 789},
		},
		CurrentMatch: 1,
	})
	s := mockSession{}
	factory := botcommands.NewTourneyCommandRequestFactory(&s, repo, &cclient)

	c := factory.CreateWinnerCommand(&msg, &user)

	c.(onMessageCreateCommand).ExecuteMessageCreateCommand()

	if cclient.query.WinnerID != 123 {
		log.Println("Winner ID incorrect or not set.")
		t.Fail()
	}
	if s.getReactedMessage() != 542 {
		log.Println("Message not given a reaction.")
		t.Fail()
	}
	r, _ := repo.HasMatchInProgress(1)
	if r {
		t.Fail()
		log.Println("Tourney still has match in progress.")
	}
}

func TestFinishTourney(t *testing.T) {
	//Given: A challonge bracket with 3 losers match and 2 waiting for opponents
	msg := disgord.MessageCreate{Message: &disgord.Message{
		ID:      542,
		Content: "",
		GuildID: 123,
	}}

	cclient := mockChallongeClient{}
	wm1 := challonge.Match{ID: 1, Round: -1, Player1ID: 123, Player2ID: 987}
	wm2 := challonge.Match{ID: 2, Round: -1, Player1ID: 654}
	lm1 := challonge.Match{ID: 3, Round: -1, Player1ID: 542}
	cclient.addMatches(wm1, wm2, lm1)
	cclient.setTourneyID("test")

	user := model.Users{UsersID: 1, DiscordUsersID: 1}
	repo := &mockTourneyDB{tourneys: make(map[model.Snowflake]*model.Tournament)}
	//And: A tourney saved with matching participants
	repo.SaveTourney(&model.Tournament{
		TournamentID:    1,
		ChallongeID:     "test",
		User:            1,
		DiscordServerID: 123,
		Participants: []model.TournamentParticipant{
			{Name: "user1", ChallongeID: 542},
			{Name: "test user", ChallongeID: 987},
			{Name: "winner", ChallongeID: 123},
			{Name: "loser", ChallongeID: 234},
			{Name: "tse", ChallongeID: 654},
			{Name: "noit", ChallongeID: 789},
		},
		CurrentMatch: 1,
	})
	s := mockSession{}
	factory := botcommands.NewTourneyCommandRequestFactory(&s, repo, &cclient)

	c := factory.CreateTourneyCloseCommand(&msg, &user)

	c.(onMessageCreateCommand).ExecuteMessageCreateCommand()

	r, _ := repo.GetTourneyByServer(123)

	if r.TournamentID != 0 {
		log.Println("Tournament not removed.")
		t.Fail()
	}
}
