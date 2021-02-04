package strawpolldeadline

import (
	"database/sql"
	"discordbot/databaseclient"
	"log"
)


type StrawpollDeadlineRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *StrawpollDeadlineRepository {
	return &StrawpollDeadlineRepository{
		db: db,
	}
}

func (r *StrawpollDeadlineRepository) SaveStrawpollDeadline(strawpollDeadline *databaseclient.StrawpollDeadline) error {
	const query = `INSERT INTO strawpoll_deadline(strawpoll_id, author, guild, channel, role) VALUES (?, ?, ?, ?, ?);`

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
		strawpollDeadline.StrawpollID,
		strawpollDeadline.User,
		strawpollDeadline.Guild,
		strawpollDeadline.Channel,
		strawpollDeadline.Role)

	if err != nil {
		return err
	}

	strawpollCommandId, _ := result.LastInsertId()
	if err != nil {
		return err
	}
	tx.Commit()

	strawpollDeadline.StrawpollDeadlineID = strawpollCommandId

	return nil
}

func (r *StrawpollDeadlineRepository) GetAllStrawpollDeadlines() []databaseclient.StrawpollDeadline {
	const query = `SELECT * FROM strawpoll_deadline;`

	rows, _ := r.db.Query(query)
	if rows.Err() != nil {
		log.Println("Error: ", rows.Err())
		return []databaseclient.StrawpollDeadline{}
	}

	completedCommand := []databaseclient.StrawpollDeadline{}

	for rows.Next() {
		row := databaseclient.StrawpollDeadline{}
		err := rows.Scan(
			&row.StrawpollDeadlineID,
			&row.User,
			&row.StrawpollID,
			&row.Guild,
			&row.Channel,
			&row.Role)
		if err != nil {
			log.Println(err)
			continue
		}
		completedCommand = append(completedCommand, row)
	}

	return completedCommand
}

func (r *StrawpollDeadlineRepository) DeleteStrawpollDeadlineByID(ID int64) error{
	const query = `DELETE FROM strawpoll_deadline WHERE strawpoll_deadline_id = ?;`

	result, err := r.db.Exec(query, ID)

	if err != nil {
		return err
	}

	if num, _ := result.RowsAffected(); num < 1 {
		return err
	}
	return nil
}