package repositories

import (
	"database/sql"
	"discordbot/commands"
)

type mangaNotificationRepo struct {
	db *sql.DB
}

type mangaLinkRepo struct {
	db *sql.DB
}

func NewMangaNotificationRepository(db *sql.DB) *mangaNotificationRepo {
	return &mangaNotificationRepo{
		db: db,
	}
}

func NewMangaLinkRepository(db *sql.DB) *mangaLinkRepo {
	return &mangaLinkRepo{
		db: db,
	}
}

func (r *mangaNotificationRepo) SaveMangaNotification(m *commands.MangaNotification) error {
	const query = `INSERT INTO manga_notification (author, guild, channel, role) VALUES (?, ?, ?, ?);`
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

	m.MangaNotificationID = ID
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
			&row.Guild,
			&row.Channel,
			&row.Role)
		if err != nil {
			return []commands.MangaNotification{}, err
		}
		completedCommand = append(completedCommand, row)
	}

	return completedCommand, nil
}

func (r *mangaNotificationRepo) AddMangaLink(mangaNotificationId int64, mangaLinkId int64) error {
	const query = `INSERT INTO manga_notification_links(manga_notification_id, manga_link_id) VALUES (?, ?);`
	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(mangaNotificationId, mangaLinkId)

	if err != nil {
		return err
	}
	tx.Commit()

	return nil
}

func (r *mangaLinkRepo) SaveMangaLink(m *commands.MangaLink) error {
	const query = `INSERT INTO manga_links(manga_link) VALUES (?);`
	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	result, err := stmt.Exec(m.MangaLink)

	if err != nil {
		return err
	}

	ID, _ := result.LastInsertId()
	if err != nil {
		return err
	}
	tx.Commit()

	m.MangaLinkID = ID
	return nil
}

func (r *mangaLinkRepo) GetMangaLinkByLink(link string) (commands.MangaLink, error) {
	const query = `SELECT * FROM manga_links WHERE manga_link = ?;`

	row := r.db.QueryRow(query, link)
	if row.Err() != nil && row.Err().Error() != "sql: no rows in result set" { 
		return commands.MangaLink{}, nil
	} else if row.Err() != nil {
		return commands.MangaLink{}, row.Err()
	}

	completedCommand := commands.MangaLink{}

	err := row.Scan(
		&completedCommand.MangaLinkID,
		&completedCommand.MangaLink)

	if err != nil {
		return commands.MangaLink{}, err
	}

	const query2 = `SELECT mn.manga_notification_id, mn.author, mn.guild, mn.channel, mn.role FROM manga_notification_links as mnl 
					JOIN manga_notification as mn on mn.manga_notification_id = mnl.manga_notification_id
					WHERE mnl.manga_link_id = ?;`
	subqueryrows, _ := r.db.Query(query2, completedCommand.MangaLinkID)
	if subqueryrows.Err() != nil {
		return commands.MangaLink{}, subqueryrows.Err()
	}

	for subqueryrows.Next() {
		row := commands.MangaNotification{}
		err := subqueryrows.Scan(
			&row.MangaNotificationID,
			&row.User,
			&row.Guild,
			&row.Channel,
			&row.Role)
		if err != nil {
			return commands.MangaLink{}, err
		}
		completedCommand.MangaNotifications = append(completedCommand.MangaNotifications, row)
	}

	return completedCommand, nil
}

func (r *mangaLinkRepo) GetAllMangaLinks() ([]commands.MangaLink, error) {
	const query = `SELECT * FROM manga_links;`

	rows, _ := r.db.Query(query)
	if rows.Err() != nil {
		return []commands.MangaLink{}, rows.Err()
	}
	defer rows.Close()

	completedCommand := []commands.MangaLink{}

	for rows.Next() {
		row := commands.MangaLink{MangaNotifications: []commands.MangaNotification{}}
		err := rows.Scan(
			&row.MangaLinkID,
			&row.MangaLink)
		if err != nil {
			return []commands.MangaLink{}, err
		}
		completedCommand = append(completedCommand, row)
	}

	if rows.Err() != nil {
		return []commands.MangaLink{}, rows.Err()
	}

	const query2 = `SELECT mn.manga_notification_id, mn.author, mn.guild, mn.channel, mn.role FROM manga_notification_links as mnl 
					JOIN manga_notification as mn on mn.manga_notification_id = mnl.manga_notification_id
					WHERE mnl.manga_link_id = ?;`
	for i := range completedCommand {
		subqueryrows, _ := r.db.Query(query2, completedCommand[i].MangaLinkID)
		if subqueryrows.Err() != nil {
			return []commands.MangaLink{}, subqueryrows.Err()
		}

		for subqueryrows.Next() {
			row := commands.MangaNotification{}
			err := subqueryrows.Scan(
				&row.MangaNotificationID,
				&row.User,
				&row.Guild,
				&row.Channel,
				&row.Role)
			if err != nil {
				return []commands.MangaLink{}, err
			}
			completedCommand[i].MangaNotifications = append(completedCommand[i].MangaNotifications, row)
		}
		subqueryrows.Close()
	}

	return completedCommand, nil
}
