package twitterfollow

import (
	"database/sql"
	"discordbot/repositories"

	log "github.com/sirupsen/logrus"
)

type TwitterFollowRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *TwitterFollowRepository {
	return &TwitterFollowRepository{
		db: db,
	}
}

func (r *TwitterFollowRepository) GetFollowedUser(screenName string) ([]repositories.TwitterFollowCommand, error) {
	const query = `SELECT * FROM twitter_follow_command WHERE screen_name = ?;`

	rows, err := r.db.Query(query, screenName)
	if err != nil {
		return []repositories.TwitterFollowCommand{}, err
	}

	completedCommand := []repositories.TwitterFollowCommand{}

	for rows.Next() {
		row := repositories.TwitterFollowCommand{}
		err := rows.Scan(
			&row.TwitterFollowCommandID,
			&row.User,
			&row.ScreenName,
			&row.Channel,
			&row.Guild,
			&row.ScreenNameID)
		if err != nil {
			log.Error(err)
			continue
		}
		completedCommand = append(completedCommand, row)
	}

	return completedCommand, nil
}

func (r *TwitterFollowRepository) SaveUserToFollow(twitterFollow *repositories.TwitterFollowCommand) error {
	const query = `INSERT INTO twitter_follow_command(author, screen_name, channel, guild, screen_name_id) VALUES (?, ?, ?, ?, ?);`

	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	result, err := stmt.Exec(
		twitterFollow.User,
		twitterFollow.ScreenName,
		twitterFollow.Channel,
		twitterFollow.Guild,
		twitterFollow.ScreenNameID)

	if err != nil {
		return err
	}

	twitterFollowID, _ := result.LastInsertId()
	if err != nil {
		return err
	}
	tx.Commit()

	twitterFollow.TwitterFollowCommandID = twitterFollowID

	return nil
}

func (r *TwitterFollowRepository) DeleteFollowedUser(screenName string, guild repositories.Snowflake) error {
	const query = `DELETE FROM twitter_follow_command WHERE screen_name = ? AND guild = ?;`

	result, err := r.db.Exec(query, screenName, guild)

	if err != nil {
		return err
	}

	if num, err := result.RowsAffected(); num < 1 {
		return err
	}
	return nil
}

func (r *TwitterFollowRepository) GetAllFollowedUsersInServer(guild repositories.Snowflake) ([]repositories.TwitterFollowCommand, error) {
	const query = `SELECT * FROM twitter_follow_command WHERE guild = ?;`

	rows, err := r.db.Query(query, guild)
	if err != nil {
		return []repositories.TwitterFollowCommand{}, err
	}

	completedCommand := []repositories.TwitterFollowCommand{}

	for rows.Next() {
		row := repositories.TwitterFollowCommand{}
		err := rows.Scan(
			&row.TwitterFollowCommandID,
			&row.User,
			&row.ScreenName,
			&row.Channel,
			&row.Guild,
			&row.ScreenNameID)
		if err != nil {
			return []repositories.TwitterFollowCommand{}, err
		}
		completedCommand = append(completedCommand, row)
	}

	return completedCommand, nil
}

func (r *TwitterFollowRepository) GetAllUniqueFollowedUsers() ([]repositories.TwitterFollowCommand, error) {
	const query = `SELECT * FROM twitter_follow_command WHERE screen_name_id IS NOT NULL GROUP BY screen_name_id;`

	rows, err := r.db.Query(query)
	if err != nil {
		return []repositories.TwitterFollowCommand{}, err
	}

	completedCommand := []repositories.TwitterFollowCommand{}

	for rows.Next() {
		row := repositories.TwitterFollowCommand{}
		err := rows.Scan(
			&row.TwitterFollowCommandID,
			&row.User,
			&row.ScreenName,
			&row.Channel,
			&row.Guild,
			&row.ScreenNameID)
		if err != nil {
			return []repositories.TwitterFollowCommand{}, err
		}
		completedCommand = append(completedCommand, row)
	}

	return completedCommand, nil
}
