package tourneyrepo

import (
	"database/sql"
	"discordbot/commands"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) SaveTourney(t *commands.Tournament) error {
	if t.TournamentID == 0 {
		return r.saveNewTourney(t)
	}
	return r.updateTourney(t)
}

func (r *repository) updateTourney(t *commands.Tournament) error {
	const query = "UPDATE tournament SET (author, challonge_id, discord_server_id, current_match) = (?, ?, ?, ?);"

	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		t.User,
		t.ChallongeID,
		t.DiscordServerID,
		t.CurrentMatch,
	)

	if err != nil {
		return tx.Rollback()
	}

	tx.Commit()

	err = r.saveParticipants(t.TournamentID, t.Participants)

	if err != nil {
		return err
	}

	err = r.saveOrganizers(t.TournamentID, t.Organizers)

	if err != nil {
		return err
	}

	

	return nil
}

func (r *repository) saveNewTourney(t *commands.Tournament) error {
	const query = "INSERT INTO tournament (author, challonge_id, discord_server_id, current_match) VALUES (?, ?, ?, ?);"

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
		t.User,
		t.ChallongeID,
		t.DiscordServerID,
		t.CurrentMatch,
	)

	if err != nil {
		return err
	}

	tourneyID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	tx.Commit()

	t.TournamentID = tourneyID

	err = r.saveParticipants(tourneyID, t.Participants)

	if err != nil {
		return err
	}

	err = r.saveOrganizers(tourneyID, t.Organizers)

	if err != nil {
		return err
	}

	return nil
}

func (r *repository) saveOrganizers(tournamentID int64, os []commands.Users) error {
	const deletequery = `DELETE FROM tournament_organizer_xref WHERE tournament_id = ?;`
	
	_, err := r.db.Exec(deletequery, tournamentID)
	
	if err != nil {
		return err
	}

	const insertquery = `INSERT INTO tournament_organizer_xref (tournament_id, users_id) VALUES (?, ?);`
	tx, err := r.db.Begin()
	for _, o := range(os) {
		if o.UsersID == 0 {
			return tx.Rollback()
		}

		_, err = tx.Exec(insertquery, tournamentID, o.UsersID)

		if err != nil {
			return tx.Rollback()
		}
	}
	tx.Commit()
	return nil
}

func (r *repository) saveParticipants(tournamentID int64, ps []commands.TournamentParticipant) error {
	const deletequery = `DELETE FROM tournament_participant_xref WHERE tournament_id = ?;`
	
	_, err := r.db.Exec(deletequery, tournamentID)
	
	if err != nil {
		return err
	}

	const insertquery = `INSERT INTO tournament_participant_xref (tournament_id, tournament_participant_id) VALUES (?, ?);`
	for _, p := range(ps) {
		if p.TournamentParticipantID == 0 {
			r.saveNewParticipant(&p, tournamentID)
		}

		_, err = r.db.Exec(insertquery, tournamentID, p.TournamentParticipantID)

		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) saveNewParticipant(p *commands.TournamentParticipant, t int64) error {
	const query = `INSERT INTO tournament_participant (name, challonge_id) VALUES (?, ?);`

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
		p.Name,
		p.ChallongeID,
	)

	if err != nil {
		return err
	}

	pID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	tx.Commit()

	p.TournamentParticipantID = pID

	return nil
}

func (r *repository) GetTourneyByServer(t commands.Snowflake) (commands.Tournament, error) {
	const query = `SELECT * FROM tournament
	 WHERE discord_server_id = ?`

	row := r.db.QueryRow(query, t)
	if row.Err() != nil {
		return commands.Tournament{}, row.Err()
	}

	result := commands.Tournament{}

	row.Scan(
		&result.TournamentID,
		&result.User,
		&result.ChallongeID,
		&result.DiscordServerID,
		&result.CurrentMatch,
	)

	to, err := r.getTournamentOrganizers(result.TournamentID)

	if err != nil {
		return commands.Tournament{}, err
	}

	tp, err := r.getTournamentParticipants(result.TournamentID)

	if err != nil {
		return commands.Tournament{}, err
	}

	result.Participants = tp
	result.Organizers = to

	return result, nil
}

func (r *repository) getTournamentParticipants(tournamentID int64) ([]commands.TournamentParticipant, error) {
	const query = `SELECT tp.tournament_participant_id, tp.name, tp.challonge_id 
	FROM tournament_participant_xref as tpxref
	JOIN tournament_participant as tp ON tp.tournament_participant_id = tpxref.tournament_participant_id
	WHERE tpxref.tournament_id = ?;`
	
	rows, err := r.db.Query(query, tournamentID)

	if err != nil {
		return nil, err
	}

	var result []commands.TournamentParticipant
	for rows.Next() {
		t := commands.TournamentParticipant{}
		err = rows.Scan(
			&t.TournamentParticipantID,
			&t.Name,
			&t.ChallongeID,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, t)
	}
	rows.Close()

	return result, nil
}

func (r *repository) getTournamentOrganizers(tournamentID int64) ([]commands.Users, error) {
	const query = `SELECT users.users_id, users.discord_users_id, users.user_name, users.is_admin
	FROM tournament_organizer_xref as toxref
	JOIN users ON users.users_id = toxref.users_id
	WHERE toxref.tournament_id = ?;`

	rows, err := r.db.Query(query, tournamentID)

	if err != nil {
		return nil, err
	}

	var result []commands.Users
	for rows.Next() {
		u := commands.Users{}
		err = rows.Scan(
			&u.UsersID,
			&u.DiscordUsersID,
			&u.UserName,
			&u.IsAdmin,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, u)
	}
	rows.Close()

	return result, nil
}

func (r *repository) AddTourneyOrganizer(userID int64, tourneyID int64) error {
	return r.saveOrganizers(tourneyID, []commands.Users{{UsersID: userID}})
}

func (r *repository) IsUserTourneyOrganizer(userID int64, tourneyID int64) (bool, error) {
	const query = `SELECT * FROM tournament_organizer_xref WHERE users_id = ? AND tournament_id = ?;`

	row := r.db.QueryRow(query, userID, tourneyID)

	if row.Err() != nil {
		return false, row.Err()
	}

	return true, nil
}

func (r *repository) HasMatchInProgress(discordServerID int64) (bool, error) {
	const query = `SELECT * FROM tournament WHERE discord_server_id = ?;`

	row := r.db.QueryRow(query, discordServerID)

	if row.Err() != nil {
		return false, row.Err()
	}

	return true, nil
}

func (r *repository) RemoveTourney(discordServerID commands.Snowflake) error {
	const query = `DELETE FROM tournament WHERE discord_server_id = ?;`
	_, err := r.db.Exec(query, discordServerID)
	
	if err != nil {
		return err
	}
	
	return nil
}
