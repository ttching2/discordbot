package repositories

import (
	"database/sql"
	"discordbot/repositories/model"
)

type tourneyRepo struct {
	db *sql.DB
}

func NewTourneyRepo(db *sql.DB) *tourneyRepo {
	return &tourneyRepo{
		db: db,
	}
}

func (r *tourneyRepo) SaveNewTourney(model.Tournament) error {
	return nil
}

func (r *tourneyRepo) GetTourneyByServer(model.Snowflake) (model.Tournament, error) {
	return model.Tournament{}, nil
}
