package commands

import (
	"discordbot/challonge"
	"os"

	"github.com/sirupsen/logrus"
)

const CommandPrefix = "$"

var log = &logrus.Logger{
	Out:          os.Stderr,
	Formatter:    new(logrus.TextFormatter),
	Hooks:        make(logrus.LevelHooks),
	Level:        logrus.InfoLevel,
	ReportCaller: true,
}

type challongeClient interface {
	GetParticipants(tourneyID string) []challonge.Participant
	GetMatches(tourneyID string) []challonge.Match
	GetMatch(tourneyID string, matchID int) challonge.Match
	UpdateMatch(tourneyID string, matchID int, params challonge.MatchQueryParams)
}