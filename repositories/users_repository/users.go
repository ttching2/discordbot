package users_repository

import (
	"database/sql"
	"discordbot/repositories/model"

	log "github.com/sirupsen/logrus"
)

type UsersRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *UsersRepository {
	return &UsersRepository{
		db: db,
	}
}

func (r *UsersRepository) GetUserByDiscordId(user model.Snowflake) (model.Users, error) {
	const query = `SELECT * FROM users WHERE discord_users_id = ?;`

	row := r.db.QueryRow(query, user)
	if row.Err() != nil {
		log.WithField("user", user).Error(row.Err())
		return model.Users{}, row.Err()
	}

	result := model.Users{}

	row.Scan(
		&result.UsersID,
		&result.DiscordUsersID,
		&result.UserName,
		&result.IsAdmin)

	return result, nil
}

func (r *UsersRepository) DoesUserExist(user model.Snowflake) bool {
	const query = `SELECT * FROM users WHERE discord_users_id = ?;`

	rows, err := r.db.Query(query, user)
	if err != nil {
		log.WithField("user", user).Error(err)
		return false
	}

	defer rows.Close()

	return rows.Next()
}

func (r *UsersRepository) SaveUser(user *model.Users) error {
	const query = `INSERT INTO users(discord_users_id, is_admin, user_name) VALUES (?, ?, ?);`

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
		user.DiscordUsersID,
		user.IsAdmin,
		user.UserName)

	if err != nil {
		return err
	}

	ID, _ := result.LastInsertId()
	if err != nil {
		return err
	}
	tx.Commit()

	user.UsersID = ID

	return nil
}
