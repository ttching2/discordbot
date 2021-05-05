package botcommands

import "discordbot/repositories/model"

type TournamentRepository interface {
	SaveTourney(*model.Tournament) error
	GetTourneyByServer(model.Snowflake) (model.Tournament, error)
	AddTourneyOrganizer(userID int64, tourneyID int64) error
	IsUserTourneyOrganizer(userID int64, tourneyID int64) (bool, error)
	RemoveTourney(discordServerID model.Snowflake) error
}
