package challonge

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

const matchIndexURL = "/tournaments/%s/matches.json"
const matchUpdateURL = "/tournaments/%s/matches/%s.json"

type MatchContainer struct {
	Match Match
}

type Match struct {
	AttachmentCount           int        `json:"attachment_count"`
	CreatedAt                 *time.Time `json:"created_at"`
	GroupID                   int        `json:"group_id"`
	HasAttachment             bool       `json:"has_attachment"`
	ID                        int
	Identifier                string
	Location                  string
	LoserID                   int  `json:"loser_id"`
	Player1ID                 int  `json:"player1_id"`
	Player1IsPrereqMatchLoser bool `json:"player1_is_prereq_match_loser"`
	Player1PrereqMatchID      int  `json:"player1_prereq_match_id"`
	Player1Votes              int  `json:"player1_votes"`
	Player2ID                 int  `json:"player2_id"`
	Player2IsPrereqMatchLoser bool `json:"player2_is_prereq_match_loser"`
	Player2PrereqMatchID      int  `json:"player2_prereq_match_id"`
	Player2Votes              int  `json:"player2_votes"`
	Round                     int
	ScheduledTime             *time.Time `json:"scheduled_time"`
	StartedAt                 *time.Time `json:"started_at"`
	State                     string
	TournamentID              int        `json:"tournament_id"`
	UnderwayAt                string     `json:"underway_at"`
	UpdatedAt                 *time.Time `json:"updated_at"`
	WinnerID                  int        `json:"winner_id"`
	PrerequisiteMatchIdsCsv   string     `json:"prerequisite_match_ids_csv"`
	ScoresCsv                 string     `json:"scores_csv"`
}

type MatchQueryParams struct {
	MatchScore  MatchScore
	WinnerID    int
	Player1Vote int
	Player2Vote int
}

func createMatchParams(m MatchQueryParams) string {
	p := "?"
	if m.WinnerID != 0 {
		p = fmt.Sprintf("%smatch[winner_id]=%v&", p, m.WinnerID)
		p = fmt.Sprintf("%smatch[scores_csv]=\"%v-%v\"&", p, m.MatchScore.Player1Score, m.MatchScore.Player2Score)
	}
	p = fmt.Sprintf("%smatch[player1_votes]=%v&", p, m.Player1Vote)
	p = fmt.Sprintf("%smatch[player2_votes]=%v", p, m.Player2Vote)
	return p
}

type MatchScore struct {
	Player1Score int
	Player2Score int
}

type matchClient struct {
	baseClient
}

func createMatchIndexURL(t string) string {
	return fmt.Sprintf(matchIndexURL, t)
}

//Index - matches do not show until tournament has started
func (c *matchClient) Index(tournamentID string) []*MatchContainer {
	body, err := c.getRequest(c.getAPIURL() + createMatchIndexURL(tournamentID))
	if err != nil {
		log.Error(err)
		return nil
	}

	t := []*MatchContainer{}
	err = json.Unmarshal([]byte(body), &t)
	if err != nil {
		log.Error(err)
		return nil
	}

	return t
}

func (c *matchClient) Show(tournamentID string, matchID string) MatchContainer {
	body, err := c.getRequest(c.getAPIURL() + fmt.Sprintf(matchUpdateURL, tournamentID, matchID))
	if err != nil {
		log.Error(err)
		return MatchContainer{}
	}

	t := MatchContainer{}
	err = json.Unmarshal([]byte(body), &t)
	if err != nil {
		log.Error(err)
		return MatchContainer{}
	}

	return t
}

func (c *matchClient) Update(tournamentID string, matchID string, params MatchQueryParams) error {
	p := createMatchParams(params)
	url := c.getAPIURL() + fmt.Sprintf(matchUpdateURL, tournamentID, matchID) + p
	return c.putRequest(url)
}
