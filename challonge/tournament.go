package challonge

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

const tournamentURL = "/tournaments/%s.json"

type Tournaments struct {
	Tournament Tournament
}

type Tournament struct {
	AcceptAttachments                bool `json:"accept_attachments"`
	AllowParticipantMatchReporting   bool `json:"allow_participant_match_reporting"`
	AnonymousVoting                  bool `json:"anonymous_voting"`
	Category                         string
	CheckInDuration                  string    `json:"check_in_duration"`
	CompletedAt                      string    `json:"completed_at"`
	CreatedAt                        time.Time `json:"created_at"`
	CreatedByAPI                     bool      `json:"created_by_api"`
	CreditCapped                     bool      `json:"credit_capped"`
	Description                      string
	GameID                           int  `json:"game_id"`
	GroupStagesEnabled               bool `json:"group_stages_enabled"`
	HideForum                        bool `json:"hide_forum"`
	HideSeeds                        bool `json:"hide_seeds"`
	HoldThirdPlaceMatch              bool `json:"hold_third_place_match"`
	ID                               int
	MaxPredictionsPerUser            int `json:"max_predictions_per_user"`
	Name                             string
	NotifyUsersWhenMatchesOpen       bool   `json:"notify_users_when_matches_open"`
	NotifyUsersWhenTheTournamentEnds bool   `json:"notify_users_when_the_tournaments_ends"`
	OpenSignup                       bool   `json:"open_signup"`
	ParticipantsCount                int    `json:"participants_count"`
	PredictionMethod                 int    `json:"prediction_method"`
	PredictionsOpenedAt              string `json:"predictions_opened_at"`
	Private                          bool
	ProgressMeter                    int       `json:"progress_meter"`
	PtsForBye                        string    `json:"pts_for_bye"`
	PtsForGameTie                    string    `json:"pts_for_game_tie"`
	PtsForGameWin                    string    `json:"pts_for_game_win"`
	PtsForMatchTie                   string    `json:"pts_for_match_tie"`
	PtsForMatchWin                   string    `json:"pts_for_match_win"`
	QuickAdvance                     bool      `json:"quick_advance"`
	RankedBy                         string    `json:"ranked_by"`
	RequireScoreAgreement            bool      `json:"require_score_agreement"`
	RrPtsForGameTie                  string    `json:"rr_pts_for_game_tie"`
	RrPtsForGameWin                  string    `json:"rr_pts_for_game_win"`
	RrPtsForMatchTie                 string    `json:"rr_pts_for_match_tie"`
	RrPtsForMatchWin                 string    `json:"rr_pts_for_match_win"`
	SequentialPairings               bool      `json:"sequential_pairings"`
	ShowRounds                       bool      `json:"show_rounds"`
	SignupCap                        string    `json:"signup_cap"`
	StartAt                          string    `json:"start_at"`
	StartedAt                        time.Time `json:"started_at"`
	StartedCheckingInAt              string    `json:"started_checking_in_at"`
	State                            string
	SwissRounds                      int `json:"swiss_rounds"`
	Teams                            bool
	TieBreaks                        []string  `json:"tie_breaks"`
	TournamentType                   string    `json:"tournament_type"`
	UpdatedAt                        time.Time `json:"updated_at"`
	URL                              string
	DescriptionSource                string `json:"description_source"`
	Subdomain                        string
	FullChallongeURL                 string `json:"full_challonge_url"`
	LiveImageURL                     string `json:"live_image_url"`
	SignUpURL                        string `json:"sign_up_url"`
	ReviewBeforeFinalizing           bool   `json:"review_before_finalizing"`
	AcceptingPredictions             bool   `json:"accepting_predictions"`
	ParticipantsLocked               bool   `json:"participants_locked"`
	GameName                         string `json:"game_name"`
	ParticipantsSwappable            bool   `json:"participants_swappable"`
	TeamConvertable                  bool   `json:"team_convertable"`
	GroupStagesWereStarted           bool   `json:"group_stages_were_started"`
}

type tournamentClient struct {
	baseClient
}

func (c *tournamentClient) createTournamentURL(ID string) string {
	return fmt.Sprintf(tournamentURL, ID)
}

func (c *tournamentClient) Show(ID string) *Tournaments {
	body, err := c.getRequest(c.getAPIURL()+c.createTournamentURL(ID))
	if err != nil {
		log.Error(err)
		return nil
	}

	t := Tournaments{}
	err = json.Unmarshal([]byte(body), &t)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &t
}
