package challonge

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)


const participantsURL = "/tournaments/%s/participants.json"

type ParticipantContainer struct {
	Participant Participant
}

type Participant struct {
	Active                                bool
	CheckedInAt                           *time.Time `json:"checked_in_at"`
	CreatedAt                             *time.Time `json:"created_at"`
	FinalRank                             int        `json:"final_rank"`
	GroupID                               int        `json:"group_id"`
	Icon                                  string
	ID                                    int
	InvitationID                          string `json:"invitation_id"`
	InviteEmail                           string `json:"invite_email"`
	Misc                                  string
	Name                                  string
	OnWaitingList                         bool `json:"on_waiting_list"`
	Seed                                  int
	TournamentID                          int        `json:"tournament_id"`
	UpdatedAt                             *time.Time `json:"updated_at"`
	ChallongeUsername                     string     `json:"challonge_username"`
	ChallongeEmailAddressVerified         string     `json:"challonge_email_address_verified"`
	Removable                             bool
	ParticipatableOrInvitationAttached    bool   `json:"participatable_or_invitation_attached"`
	ConfirmRemove                         bool   `json:"confirm_remove"`
	InvitationPending                     bool   `json:"invitation_pending"`
	DisplayNameWithInvitationEmailAddress string `json:"display_name_with_invitation_email_address"`
	EmailHash                             string `json:"email_hash"`
	Username                              string
	AttachedParticipatablePortraitURL     string `json:"attached_participatable_portrait_url"`
	CanCheckIn                            bool   `json:"can_check_in"`
	CheckedIn                             bool   `json:"checked_in"`
	Reactivatable                         bool
}

type participantsClient struct {
	baseClient
}

func getParticipantsIndexURL(t string) string {
	return fmt.Sprintf(participantsURL, t)
}

func (c *participantsClient) Index(tournamentID string) []*ParticipantContainer {
	body, err := c.getRequest(c.getAPIURL() + getParticipantsIndexURL(tournamentID))

	if err != nil {
		log.Error(err)
		return nil
	}

	t := []*ParticipantContainer{}
	err = json.Unmarshal([]byte(body), &t)
	if err != nil {
		log.Error(err)
		return nil
	}

	return t
}