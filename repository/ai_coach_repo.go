package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"pijar/model"
	"time"
	"context"
	"github.com/google/uuid"
)

type CoachSessionRepository interface {
	// Session Management
	CreateSession(c context.Context, userID int, input string) (string, error)
	UpdateSessionResponse(c context.Context, sessionID string, response string) error
	GetOrCreateConversationContext(c context.Context, userID int, sessionID string) (*model.ConversationContext, error)
	SaveConversationContext(c context.Context, ctx *model.ConversationContext) error
	SaveConversation(c context.Context, userID int, sessionID, userInput, aiResponse string) error
	GetSessionHistory(c context.Context, userID int, sessionID string, limit int) ([]model.Message, error)
	GetUserSessions(c context.Context, userID int) ([]model.CoachSession, error)
	DeleteSession(c context.Context, userID int, sessionID string) error
}

type coachSessionRepository struct {
	db *sql.DB
}

func (r *coachSessionRepository) CreateSession(c context.Context, userID int, input string) (string, error) {
	// Cek apakah user ada
	var exists bool
	checkUserQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := r.db.QueryRow(checkUserQuery, userID).Scan(&exists)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", fmt.Errorf("user dengan id %d tidak ditemukan", userID)
	}

	// Generate session ID baru
	sessionID := uuid.New().String()
	now := time.Now()

	// Input session ke database
	query := `INSERT INTO coach_sessions 
		  (user_id, session_id, timestamp, user_input) 
		  VALUES ($1, $2, $3, $4) 
		  RETURNING session_id`

	_, err = r.db.Exec(query, userID, sessionID, now, input)
	if err != nil {
		return "", fmt.Errorf("gagal membuat sesi: %w", err)
	}

	return sessionID, nil
}

func (r *coachSessionRepository) UpdateSessionResponse(c context.Context, sessionID string, response string) error {
	query := `UPDATE coach_sessions 
	         SET ai_response = $1, updated_at = $2 
	         WHERE session_id = $3`
	_, err := r.db.Exec(query, response, time.Now(), sessionID)
	return err
}

func (r *coachSessionRepository) GetOrCreateConversationContext(c context.Context, userID int, sessionID string) (*model.ConversationContext, error) {
	// Cek apakah session ada
	var exists bool
	checkSessionQuery := `SELECT EXISTS(SELECT 1 FROM coach_sessions WHERE session_id = $1 AND user_id = $2)`
	err := r.db.QueryRow(checkSessionQuery, sessionID, userID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa sesi: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("sesi tidak ditemukan")
	}

	// Ambil konteks dari database
	var contextJSON []byte
	query := `SELECT context FROM conversation_contexts WHERE session_id = $1`
	err = r.db.QueryRow(query, sessionID).Scan(&contextJSON)

	if err == sql.ErrNoRows {
		// Buat konteks baru jika belum ada
		ctx := &model.ConversationContext{
			SessionID: sessionID,
			Messages:  []model.Message{},
			Metadata:  make(map[string]interface{}),
		}
		return ctx, nil
	}

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil konteks: %w", err)
	}

	// Parse konteks yang ada
	var ctx model.ConversationContext
	if err := json.Unmarshal(contextJSON, &ctx); err != nil {
		return nil, fmt.Errorf("gagal mem-parse konteks: %w", err)
	}

	return &ctx, nil
}

func (r *coachSessionRepository) SaveConversation(c context.Context, userID int, sessionID, userInput, aiResponse string) error {
	// Cek apakah sesi sudah ada
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM coach_sessions WHERE session_id = $1 AND user_id = $2)",
		sessionID, userID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("gagal memeriksa sesi: %w", err)
	}

	if exists {
		// Update sesi yang sudah ada
		query := `
			UPDATE coach_sessions 
			SET user_input = $1, 
				ai_response = $2, 
				timestamp = $3,
				updated_at = $3
			WHERE session_id = $4 AND user_id = $5`

		_, err = r.db.Exec(query, userInput, aiResponse, time.Now(), sessionID, userID)
	} else {
		// Buat sesi baru
		query := `
			INSERT INTO coach_sessions 
			(user_id, session_id, user_input, ai_response, timestamp, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $6)`

		now := time.Now()
		_, err = r.db.Exec(query, userID, sessionID, userInput, aiResponse, now, now)
	}

	if err != nil {
		return fmt.Errorf("gagal menyimpan percakapan: %w", err)
	}

	return nil
}

func (r *coachSessionRepository) SaveConversationContext(c context.Context, ctx *model.ConversationContext) error {
	// Konversi konteks ke JSON
	contextJSON, err := json.Marshal(ctx)
	if err != nil {
		return fmt.Errorf("gagal mengkonversi konteks: %w", err)
	}

	// Simpan atau update konteks
	query := `
		INSERT INTO conversation_contexts (session_id, context, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (session_id) 
		DO UPDATE SET context = $2, updated_at = $3`

	_, err = r.db.Exec(query, ctx.SessionID, contextJSON, time.Now())
	if err != nil {
		return fmt.Errorf("gagal menyimpan konteks: %w", err)
	}

	return nil
}

func (r *coachSessionRepository) GetSessionHistory(c context.Context, userID int, sessionID string, limit int) ([]model.Message, error) {
	// Cek apakah session ada dan milik user yang benar
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM coach_sessions WHERE session_id = $1 AND user_id = $2)",
		sessionID, userID,
	).Scan(&exists)

	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa sesi: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("sesi tidak ditemukan atau tidak dapat diakses")
	}

	// Ambil konteks percakapan
	var contextJSON []byte
	err = r.db.QueryRow(
		"SELECT context FROM conversation_contexts WHERE session_id = $1",
		sessionID,
	).Scan(&contextJSON)

	if err == sql.ErrNoRows {
		return []model.Message{}, nil // Return empty array if no conversation context exists yet
	}

	if err != nil {
		return nil, fmt.Errorf("gagal mengambil konteks percakapan: %w", err)
	}

	// Parse konteks yang ada
	var ctx model.ConversationContext
	if err := json.Unmarshal(contextJSON, &ctx); err != nil {
		return nil, fmt.Errorf("gagal mem-parse konteks: %w", err)
	}

	// Jika ada limit, potong array messages sesuai limit
	if limit > 0 && len(ctx.Messages) > limit {
		return ctx.Messages[len(ctx.Messages)-limit:], nil
	}

	return ctx.Messages, nil
}

func (r *coachSessionRepository) GetUserSessions(c context.Context, userID int) ([]model.CoachSession, error) {
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
	query := `SELECT id, user_id,session_id, timestamp, user_input, ai_response FROM coach_sessions WHERE user_id=$1 ORDER BY timestamp DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterasi hasil query
	for rows.Next() {
		var session model.CoachSession
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.SessionID,
			&session.Timestamp,
			&session.UserInput,
			&session.AIResponse,
		)
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

func (r *coachSessionRepository) DeleteSession(c context.Context, userID int, sessionID string) error {
	query := `DELETE FROM coach_sessions WHERE user_id = $1 AND session_id = $2 `
	result, err := r.db.Exec(query, userID, sessionID)
	if err != nil {
		return err
	}

	// Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found or not owned by user")
	}

	return nil
}

func NewSession(db *sql.DB) CoachSessionRepository {
	return &coachSessionRepository{db: db}
}
