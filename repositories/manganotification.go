package repositories

import (
	"database/sql"
	"discordbot/commands"
)

type mangaNotificationRepo struct {
	db *sql.DB
}

func NewMangaNotificationRepository(db *sql.DB) *mangaNotificationRepo {
	return &mangaNotificationRepo{
		db: db,
	}
}

func (r *mangaNotificationRepo) SaveMangaNotification(m *commands.MangaNotification) error {
	const query = `INSERT INTO manga_notification (author, manga_url, guild, channel, role) VALUES (?, ?, ?, ?, ?);`
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
		m.User,
		m.MangaURL,
		m.Guild,
		m.Channel,
		m.Role)

	if err != nil {
		return err
	}

	ID, _ := result.LastInsertId()
	if err != nil {
		return err
	}
	tx.Commit()

	m.MangaNotificationID= ID
	return nil
}

func (r *mangaNotificationRepo) GetAllMangaNotifications() ([]commands.MangaNotification, error) {
	const query = `SELECT * FROM manga_notification;`

	rows, _ := r.db.Query(query)
	if rows.Err() != nil {
		return []commands.MangaNotification{}, rows.Err()
	}

	completedCommand := []commands.MangaNotification{}

	for rows.Next() {
		row := commands.MangaNotification{}
		err := rows.Scan(
			&row.MangaNotificationID,
			&row.User,
			&row.MangaURL,
			&row.Guild,
			&row.Channel,
			&row.Role)
		if err != nil {
			return []commands.MangaNotification{}, rows.Err()
		}
		completedCommand = append(completedCommand, row)
	}

	return completedCommand, nil
}
