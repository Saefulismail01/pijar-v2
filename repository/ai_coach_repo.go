package repository

import (
	"context"
	"database/sql"
	"fmt"
	"pijar/model"
	"time"
)

type CoachRepository interface {
	CreateSession(ctx context.Context, userID int, input string) (int, error)
	UpdateSessionResponse(ctx context.Context, sessionID int, response string) error
	GetSessionByUserID(ctx context.Context, userID int) ([]model.CoachSession, error)
	DeleteSessionByUserID(ctx context.Context, id int) error
}

type coachRepository struct {
	db *sql.DB
}

func (r *coachRepository) CreateSession(ctx context.Context, userID int, input string) (int, error) {
	// Cek apakah user ada
	var exists bool
	checkUserQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := r.db.QueryRowContext(ctx, checkUserQuery, userID).Scan(&exists)
	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, fmt.Errorf("user dengan id %d tidak ditemukan", userID)
	}

	//input session ke database
	var id int
	query := `INSERT INTO coach_sessions (id_user, timestamp, user_input) VALUES ($1, $2, $3) RETURNING id`
	err = r.db.QueryRowContext(ctx, query, userID, time.Now(), input).Scan(&id)
	return id, err
}

func (r *coachRepository) UpdateSessionResponse(ctx context.Context, sessionID int, response string) error {
	query := `UPDATE coach_sessions SET ai_response=$1 WHERE id=$2`
	_, err := r.db.ExecContext(ctx, query, response, sessionID)
	return err
}

func (r *coachRepository) GetSessionByUserID(ctx context.Context, userID int) ([]model.CoachSession, error) {
	var sessions []model.CoachSession

	// Cek apakah user ada
	var exists bool
	checkUserQuery := `SELECT EXISTS(SELECT 1 FROM coach_sessions WHERE id_user = $1)`
	err := r.db.QueryRowContext(ctx, checkUserQuery, userID).Scan(&exists)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("user dengan id %d belum melakukan percakapan", userID)
	}

	// Query untuk mendapatkan semua sesi dari user
	query := `SELECT id, id_user, timestamp, user_input, ai_response FROM coach_sessions WHERE id_user=$1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterasi hasil query
	for rows.Next() {
		var session model.CoachSession
		var aiResponse sql.NullString

		err := rows.Scan(&session.ID, &session.UserID, &session.Timestamp, &session.UserInput, &aiResponse)
		if err != nil {
			return nil, err
		}

		session.AIResponse = aiResponse.String

		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *coachRepository) DeleteSessionByUserID(ctx context.Context, id int) error {
	query := `DELETE FROM coach_sessions WHERE id_user = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus data session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal memeriksa hasil delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user dengan ID %d tidak ditemukan", id)
	}

	return nil
}

func NewSession(db *sql.DB) CoachRepository {
	return &coachRepository{db: db}
}
