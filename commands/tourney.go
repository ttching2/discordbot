package commands

import (
	"discordbot/challonge"
	"net/url"

	"github.com/andersfylling/disgord"
)

type tourneyCommandRequestFactory struct {
	repo            TournamentRepository
	session         DiscordSession
	challongeClient challongeClient
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

func (c *tourneyCommandRequestFactory) CreateRequest(data *disgord.MessageCreate, user *Users) interface{} {
	return &tourneyCommand{
		tourneyCommandRequestFactory: c,
		data:                         data,
		user:                         user,
	}
}

func (c *tourneyCommandRequestFactory) CreateAddOrganizerCommand(data *disgord.MessageCreate, user *Users) interface{} {
	return &addOrganizerCommand{
		tourneyCommandRequestFactory: c,
		data:                         data,
		user:                         user,
	}
}

func (c *tourneyCommandRequestFactory) CreateNextLosersCommnad(data *disgord.MessageCreate, user *Users) interface{} {
	return &nextLosersMatchCommand{
		tourneyCommandRequestFactory: c,
		data:                         data,
		user:                         user,
	}
}

func (c *tourneyCommandRequestFactory) CreateWinnerCommand(data *disgord.MessageCreate, user *Users) interface{} {
	return &matchWinnerCommand{
		tourneyCommandRequestFactory: c,
		data:                         data,
		user:                         user,
	}
}

func (c *tourneyCommandRequestFactory) CreateTourneyCloseCommand(data *disgord.MessageCreate, user *Users) interface{} {
	return &closeTourney{
		tourneyCommandRequestFactory: c,
		data:                         data,
		user:                         user,
	}
}

type tourneyCommand struct {
	*tourneyCommandRequestFactory
	data *disgord.MessageCreate
	user *Users
}

func (c *tourneyCommand) ExecuteMessageCreateCommand() {
	con := c.data.Message.Content
	u, err := url.Parse(con)
	if err != nil {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Unable to parse url")
		log.Error(err)
		return
	}

	tourneyID := u.Path[1:]
	ps := c.challongeClient.GetParticipants(tourneyID)
	var tourneyParticipants []TournamentParticipant
	for _, p := range ps {
		participant := TournamentParticipant{
			Name:        p.Name,
			ChallongeID: p.ID,
		}
		tourneyParticipants = append(tourneyParticipants, participant)
	}
	t := Tournament{
		DiscordServerID: c.data.Message.GuildID,
		User:            c.user.UsersID,
		ChallongeID:     tourneyID,
		Participants:    tourneyParticipants,
	}
	err = c.repo.SaveTourney(&t)
	if err != nil {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Something went wrong, tournament unable to start.")
		log.Error(err)
		return
	}
	c.session.ReactToMessage(c.data.Message.ID, c.data.Message.ChannelID, "👍")
}

type addOrganizerCommand struct {
	*tourneyCommandRequestFactory
	data *disgord.MessageCreate
	user *Users
}

func (c *addOrganizerCommand) ExecuteMessageCreateCommand() {
	t, err := c.repo.GetTourneyByServer(c.data.Message.GuildID)

	if t.DiscordServerID == 0 || err != nil {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Command unable to be used. Tournament not started in this server.")
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Command unable to be used. Tournament not started in this server.")
		return
	}

	err = c.repo.AddTourneyOrganizer(c.user.UsersID, t.TournamentID)
	if err != nil {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Something went wrong, organizer unable to be added.")
		log.Error(err)
		return
	}
	c.session.ReactToMessage(c.data.Message.ID, c.data.Message.ChannelID, "👍" )
}

type nextLosersMatchCommand struct {
	*tourneyCommandRequestFactory
	data *disgord.MessageCreate
	user *Users
}

func (c *nextLosersMatchCommand) ExecuteMessageCreateCommand() {
	t, err := c.repo.GetTourneyByServer(c.data.Message.GuildID)

	if t.DiscordServerID == 0 || err != nil {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Command unable to be used. Tournament not started in this server.")
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
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "No losers match is ready to be played yet.")
		return
	}

	t.CurrentMatch = nextMatch.ID
	c.repo.SaveTourney(&t)

	p1 := findParticipant(&t.Participants, nextMatch.Player1ID)
	p2 := findParticipant(&t.Participants, nextMatch.Player2ID)

	c.session.SendSimpleMessage(c.data.Message.ChannelID, p1.Name+" vs "+p2.Name)
}

func findParticipant(ps *[]TournamentParticipant, id int) *TournamentParticipant {
	for _, p := range *ps {
		if p.ChallongeID == id {
			return &p
		}
	}
	return nil
}

func findParticipantByName(ps *[]TournamentParticipant, name string) *TournamentParticipant {
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
	user *Users
}

func (c *matchWinnerCommand) ExecuteMessageCreateCommand() {
	t, err := c.repo.GetTourneyByServer(c.data.Message.GuildID)

	if t.DiscordServerID == 0 || err != nil || t.CurrentMatch == 0 {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Command unable to be used. Tournament not started in this server.")
		return
	}

	if c.data.Message.Content == "" {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Missing winner's name.")
		return
	}

	w := findParticipantByName(&t.Participants, c.data.Message.Content)

	if w == nil {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Winner's name not found.")
		return
	}

	m := c.challongeClient.GetMatch(t.ChallongeID, t.CurrentMatch)

	var score challonge.MatchScore
	if w.ChallongeID == m.Player1ID {
		score = challonge.MatchScore{Player1Score: 1}
	} else if w.ChallongeID == m.Player2ID {
		score = challonge.MatchScore{Player2Score: 1}
	} else {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Winner not found?")
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
	user *Users
}

func (c *closeTourney) ExecuteMessageCreateCommand() {
	t, err := c.repo.GetTourneyByServer(c.data.Message.GuildID)

	if t.DiscordServerID == 0 || err != nil {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "Command unable to be used. Tournament not started in this server.")
		return
	}

	err = c.repo.RemoveTourney(c.data.Message.GuildID)

	if err != nil {
		c.session.SendSimpleMessage(c.data.Message.ChannelID, "An error occurred ending tournament.")
		log.Error(err)
		return
	}
	c.session.ReactToMessage(c.data.Message.ID, c.data.Message.ChannelID, "👍")
}