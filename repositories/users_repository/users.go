package users_repository

import (
	"database/sql"
	"discordbot/repositories"
	"log"
)


type UsersRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *UsersRepository {
	return &UsersRepository{
		db: db,
	}
}

func (r *UsersRepository) GetUserByDiscordId(user repositories.Snowflake) repositories.Users {
	const query = `SELECT * FROM users WHERE discord_users_id = ?;`

	row := r.db.QueryRow(query, user)
	if row.Err() != nil {
		log.Println(row.Err())
		return repositories.Users{}
	}

	result := repositories.Users{}

	row.Scan(
		&result.UsersID,
		&result.DiscordUsersID, 
		&result.IsAdmin)

	return result
}

func (r *UsersRepository) DoesUserExist(user repositories.Snowflake) bool {
	const query = `SELECT * FROM users WHERE discord_users_id = ?;`

	rows, err := r.db.Query(query, user)
	if err != nil {
		log.Println(err)
		return false
	}

	defer rows.Close()

	return rows.Next()
}

func (r *UsersRepository) SaveUser(user *repositories.Users) error {
	const query = `INSERT INTO users(discord_users_id, is_admin) VALUES (?, ?);`

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
		user.DiscordUsersID,
		user.IsAdmin)

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