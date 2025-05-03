package repository

import (
	"database/sql"
	"fmt"
	"time"
	"pijar/model"
)

type CouchRepository interface {
	CreateSession(userID int, input string) (int, error)
	UpdateSessionResponse(sessionID int, response string) error
	GetSessionByUserID(userID int) ([]model.CoachSession, error)
}

type couchRepository struct {
	db *sql.DB
}

func (r *couchRepository) CreateSession(userID int, input string) (int, error) {
	// Cek apakah user ada
	var exists bool
	checkUserQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := r.db.QueryRow(checkUserQuery, userID).Scan(&exists)
	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, fmt.Errorf("user dengan id %d tidak ditemukan", userID)
	}

	//input session ke database
	var id int
	query := `INSERT INTO coach_sessions (id_user, timestamp, user_input) VALUES ($1, $2, $3) RETURNING id`
	err = r.db.QueryRow(query, userID, time.Now(), input).Scan(&id)
	return id, err
}

func (r *couchRepository) UpdateSessionResponse(sessionID int, response string) error {
	query := `UPDATE coach_sessions SET ai_response=$1 WHERE id=$2`
	_, err := r.db.Exec(query, response, sessionID)
	return err
}

func (r *couchRepository) GetSessionByUserID(userID int) ([]model.CoachSession, error) {
	var sessions []model.CoachSession

	// Cek apakah user ada
	var exists bool
	checkUserQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := r.db.QueryRow(checkUserQuery, userID).Scan(&exists)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("user dengan id %d tidak ditemukan", userID)
	}

	// Query untuk mendapatkan semua sesi dari user
	query := `SELECT id, id_user, timestamp, user_input, ai_response FROM coach_sessions WHERE id_user=$1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterasi hasil query
	for rows.Next() {
		var session model.CoachSession
		err := rows.Scan(&session.ID, &session.UserID, &session.Timestamp, &session.UserInput, &session.AIResponse)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}


func NewSession(db *sql.DB) CouchRepository {
	return &couchRepository{db: db}
}
