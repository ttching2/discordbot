package twitterfollow

import (
	"database/sql"
	"discordbot/databaseclient"
	"log"
)


type TwitterFollowRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *TwitterFollowRepository {
	return &TwitterFollowRepository{
		db: db,
	}
}

func (r *TwitterFollowRepository) GetFollowedUser(screenName string) []databaseclient.TwitterFollowCommand {
	const query = `SELECT * FROM twitter_follow_command WHERE screen_name = ?;`

	rows, err := r.db.Query(query, screenName)
	if err != nil {
		log.Println("Error: ", err)
		return []databaseclient.TwitterFollowCommand{}
	}

	completedCommand := []databaseclient.TwitterFollowCommand{}

	for rows.Next() {
		row := databaseclient.TwitterFollowCommand{}
		err := rows.Scan(
			&row.TwitterFollowCommandID,
			&row.User,
			&row.ScreenName,
			&row.Channel,
			&row.Guild,
			&row.ScreenNameID)
		if err != nil {
			log.Println(err)
			continue
		}
		completedCommand = append(completedCommand, row)
	}

	return completedCommand
}

func (r *TwitterFollowRepository) SaveUserToFollow(twitterFollow *databaseclient.TwitterFollowCommand) error {
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

	result , err := stmt.Exec(
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

func (r *TwitterFollowRepository) DeleteFollowedUser(screenName string, guild databaseclient.Snowflake) error {
	const query = `DELETE FROM twitter_follow_command WHERE screen_name = ? AND guild = ?;`

	result, err := r.db.Exec(query, screenName, guild)

	if err != nil {
		return err
	}

	if num, _ := result.RowsAffected(); num < 1 {
		return err
	}
	return nil
}

func (r *TwitterFollowRepository) GetAllFollowedUsersInServer(guild databaseclient.Snowflake) []databaseclient.TwitterFollowCommand {
	const query = `SELECT * FROM twitter_follow_command WHERE guild = ?;`

	rows, err := r.db.Query(query, guild)
	if err != nil {
		log.Println("Error: ", err)
		return []databaseclient.TwitterFollowCommand{}
	}

	completedCommand := []databaseclient.TwitterFollowCommand{}

	for rows.Next() {
		row := databaseclient.TwitterFollowCommand{}
		err := rows.Scan(
			&row.TwitterFollowCommandID,
			&row.User,
			&row.ScreenName,
			&row.Channel,
			&row.Guild,
			&row.ScreenNameID)
		if err != nil {
			log.Println(err)
			continue
		}
		completedCommand = append(completedCommand, row)
	}

	return completedCommand
}

func (r *TwitterFollowRepository) GetAllUniqueFollowedUsers() []databaseclient.TwitterFollowCommand {
	const query = `SELECT * FROM twitter_follow_command WHERE screen_name_id IS NOT NULL GROUP BY screen_name_id;`

	rows, err := r.db.Query(query)
	if err != nil {
		log.Println("Error: ", err)
		return []databaseclient.TwitterFollowCommand{}
	}

	completedCommand := []databaseclient.TwitterFollowCommand{}

	for rows.Next() {
		row := databaseclient.TwitterFollowCommand{}
		err := rows.Scan(
			&row.TwitterFollowCommandID,
			&row.User,
			&row.ScreenName,
			&row.Channel,
			&row.Guild,
			&row.ScreenNameID)
		if err != nil {
			log.Println(err)
			continue
		}
		completedCommand = append(completedCommand, row)
	}

	return completedCommand
}