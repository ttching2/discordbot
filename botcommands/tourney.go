package botcommands

import (
	"discordbot/challonge"
	"discordbot/repositories/model"
	"net/url"

	"github.com/andersfylling/disgord"
	log "github.com/sirupsen/logrus"
)

/*
$tourney {link (optional?)} - done
$add_organizer {discord_name} - done
$next_losers_match - done
$ammend_participant {tourney_name} {discord_name} - not rn
$win {optional - participant}  {optional - score format "1-1"} - also sends results to person if specified (done)
$finish_tourney - done
$organizer-list - list organizers
*/

const TournamentCommandString = "tournament"
const TournamentAddOrganizerString = "add-organizer"
const TournamentNextLosersMatchString = "next-losers-match"
const TournamentMatchWinString = "match-win"
const TournamentFinishString = "end-tournament"

type tourneyCommandRequestFactory struct {
	repo            TournamentRepository
	session         DiscordSession
	challongeClient challongeClient
}

type challongeClient interface {
	GetParticipants(tourneyID string) []challonge.Participant
	GetMatches(tourneyID string) []challonge.Match
	GetMatch(tourneyID string, matchID int) challonge.Match
	UpdateMatch(tourneyID string, matchID int, params challonge.MatchQueryParams)
}

func NewTourneyCommandRequestFactory(s DiscordSession, repo TournamentRepository, client challongeClient) *tourneyCommandRequestFactory {
	return &tourneyCommandRequestFactory{
		session:         s,
		repo:            repo,
		challongeClient: client,
	}
}

func (c *tourneyCommandRequestFactory) PrintHelp() string {
	return ""
}

func (c *tourneyCommandRequestFactory) CreateRequest(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &tourneyCommand{
		tourneyCommandRequestFactory: c,
		data:                         data,
		user:                         user,
	}
}

func (c *tourneyCommandRequestFactory) CreateAddOrganizerCommand(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &addOrganizerCommand{
		tourneyCommandRequestFactory: c,
		data:                         data,
		user:                         user,
	}
}

func (c *tourneyCommandRequestFactory) CreateNextLosersCommnad(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &nextLosersMatchCommand{
		tourneyCommandRequestFactory: c,
		data:                         data,
		user:                         user,
	}
}

func (c *tourneyCommandRequestFactory) CreateWinnerCommand(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &matchWinnerCommand{
		tourneyCommandRequestFactory: c,
		data:                         data,
		user:                         user,
	}
}

func (c *tourneyCommandRequestFactory) CreateTourneyCloseCommand(data *disgord.MessageCreate, user *model.Users) interface{} {
	return &closeTourney{
		tourneyCommandRequestFactory: c,
		data:                         data,
		user:                         user,
	}
}

type tourneyCommand struct {
	*tourneyCommandRequestFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c *tourneyCommand) ExecuteMessageCreateCommand() {
	con := c.data.Message.Content
	u, err := url.Parse(con)
	if err != nil {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("Unable to parse url"))
		log.Error(err)
		return
	}

	tourneyID := u.Path[1:]
	ps := c.challongeClient.GetParticipants(tourneyID)
	var tourneyParticipants []model.TournamentParticipant
	for _, p := range ps {
		participant := model.TournamentParticipant{
			Name:        p.Name,
			ChallongeID: p.ID,
		}
		tourneyParticipants = append(tourneyParticipants, participant)
	}
	t := model.Tournament{
		DiscordServerID: c.data.Message.GuildID,
		User:            c.user.UsersID,
		ChallongeID:     tourneyID,
		Participants:    tourneyParticipants,
	}
	err = c.repo.SaveTourney(&t)
	if err != nil {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("Something went wrong, tournament unable to start."))
		log.Error(err)
		return
	}
	c.session.ReactToMessage(c.data.Message.ID, c.data.Message.ChannelID, "👍")
}

type addOrganizerCommand struct {
	*tourneyCommandRequestFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c *addOrganizerCommand) ExecuteMessageCreateCommand() {
	t, err := c.repo.GetTourneyByServer(c.data.Message.GuildID)

	if t.DiscordServerID == 0 || err != nil {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("Command unable to be used. Tournament not started in this server."))
		return
	}

	err = c.repo.AddTourneyOrganizer(c.user.UsersID, t.TournamentID)
	if err != nil {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("Something went wrong, organizer unable to be added."))
		log.Error(err)
		return
	}
	c.session.ReactToMessage(c.data.Message.ID, c.data.Message.ChannelID, "👍" )
}

type nextLosersMatchCommand struct {
	*tourneyCommandRequestFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c *nextLosersMatchCommand) ExecuteMessageCreateCommand() {
	t, err := c.repo.GetTourneyByServer(c.data.Message.GuildID)

	if t.DiscordServerID == 0 || err != nil {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("Command unable to be used. Tournament not started in this server."))
		return
	}

	mcontainer := c.challongeClient.GetMatches(t.ChallongeID)

	var nextMatch challonge.Match
	for _, m := range mcontainer {
		if m.Round > 0 || m.WinnerID != 0 {
			continue
		}

		if m.Player1ID == 0 || m.Player2ID == 0 {
			continue
		}

		nextMatch = m
		break
	}

	if nextMatch.ID == 0 {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("No losers match is ready to be played yet."))
		return
	}

	t.CurrentMatch = nextMatch.ID
	c.repo.SaveTourney(&t)

	p1 := findParticipant(&t.Participants, nextMatch.Player1ID)
	p2 := findParticipant(&t.Participants, nextMatch.Player2ID)

	c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage(p1.Name+" vs "+p2.Name))
}

func findParticipant(ps *[]model.TournamentParticipant, id int) *model.TournamentParticipant {
	for _, p := range *ps {
		if p.ChallongeID == id {
			return &p
		}
	}
	return nil
}

func findParticipantByName(ps *[]model.TournamentParticipant, name string) *model.TournamentParticipant {
	for _, p := range *ps {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

type matchWinnerCommand struct {
	*tourneyCommandRequestFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c *matchWinnerCommand) ExecuteMessageCreateCommand() {
	t, err := c.repo.GetTourneyByServer(c.data.Message.GuildID)

	if t.DiscordServerID == 0 || err != nil || t.CurrentMatch == 0 {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("Command unable to be used. Tournament not started in this server."))
		return
	}

	if c.data.Message.Content == "" {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("Missing winner's name."))
		return
	}

	w := findParticipantByName(&t.Participants, c.data.Message.Content)

	if w == nil {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("Winner's name not found."))
		return
	}

	m := c.challongeClient.GetMatch(t.ChallongeID, t.CurrentMatch)

	var score challonge.MatchScore
	if w.ChallongeID == m.Player1ID {
		score = challonge.MatchScore{Player1Score: 1}
	} else if w.ChallongeID == m.Player2ID {
		score = challonge.MatchScore{Player2Score: 1}
	} else {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("Winner not found?"))
		log.Error("Winner ID does not match current participants in match.")
		return
	}

	q := challonge.MatchQueryParams{WinnerID: w.ChallongeID, MatchScore: score}

	c.challongeClient.UpdateMatch(t.ChallongeID, t.CurrentMatch, q)
	c.session.ReactToMessage(c.data.Message.ID, c.data.Message.ChannelID, "👍")
	t.CurrentMatch = 0
	err = c.repo.SaveTourney(&t)
	if err != nil {
		log.Error(err)
	}
}

type closeTourney struct {
	*tourneyCommandRequestFactory
	data *disgord.MessageCreate
	user *model.Users
}

func (c *closeTourney) ExecuteMessageCreateCommand() {
	t, err := c.repo.GetTourneyByServer(c.data.Message.GuildID)

	if t.DiscordServerID == 0 || err != nil {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("Command unable to be used. Tournament not started in this server."))
		return
	}

	err = c.repo.RemoveTourney(c.data.Message.GuildID)

	if err != nil {
		c.session.SendMessage(c.data.Message.ChannelID, createSimpleDisgordMessage("An error occurred ending tournament."))
		log.Error(err)
		return
	}
	c.session.ReactToMessage(c.data.Message.ID, c.data.Message.ChannelID, "👍")
}