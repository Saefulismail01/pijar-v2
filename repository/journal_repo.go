package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pijar/model"
	"time"
)

type JournalRepository interface {
	Create(ctx context.Context, journal *model.Journal) error
	FindAll(ctx context.Context) ([]model.Journal, error)
	FindByUserID(ctx context.Context, userID int) ([]model.Journal, error)
	FindByID(ctx context.Context, id int) (*model.Journal, error)
	Update(ctx context.Context, journal *model.Journal) error
	Delete(ctx context.Context, id int) error
}

type journalRepository struct {
	db *sql.DB
}

func NewJournalRepository(db *sql.DB) JournalRepository {
	return &journalRepository{db: db}
}

func (r *journalRepository) Create(ctx context.Context, journal *model.Journal) error {
	// Check if user exists
	query := `SELECT id FROM users WHERE id = $1`
	var userID int
	err := r.db.QueryRowContext(ctx, query, journal.UserID).Scan(&userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user dengan id %d tidak ditemukan", journal.UserID)
		}
		return fmt.Errorf("gagal memeriksa keberadaan user: %w", err)
	}

	// Set created_at dan updated_at sama saat pertama kali dibuat
	now := time.Now()
	journal.CreatedAt = now
	journal.UpdatedAt = now

	query = `INSERT INTO journals (user_id, judul, isi, perasaan, created_at, updated_at) 
	        VALUES ($1, $2, $3, $4, $5, $6) 
	        RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		journal.UserID,
		journal.Judul,
		journal.Isi,
		journal.Perasaan,
		journal.CreatedAt,
		journal.UpdatedAt,
	).Scan(&journal.ID, &journal.CreatedAt, &journal.UpdatedAt)
}

func (r *journalRepository) FindAll(ctx context.Context) ([]model.Journal, error) {
	var journals []model.Journal
	query := `SELECT id, user_id, judul, isi, perasaan, created_at, updated_at 
	         FROM journals`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var journal model.Journal
		if err := rows.Scan(
			&journal.ID,
			&journal.UserID,
			&journal.Judul,
			&journal.Isi,
			&journal.Perasaan,
			&journal.CreatedAt,
			&journal.UpdatedAt,
		); err != nil {
			return nil, err
		}
		journals = append(journals, journal)
	}

	return journals, nil
}


func (r *journalRepository) FindByUserID(ctx context.Context, userID int) ([]model.Journal, error) {
	var journals []model.Journal
	query := `SELECT id, user_id, judul, isi, perasaan, created_at, updated_at 
	         FROM journals 
	         WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var journal model.Journal
		if err := rows.Scan(
			&journal.ID,
			&journal.UserID,
			&journal.Judul,
			&journal.Isi,
			&journal.Perasaan,
			&journal.CreatedAt,
			&journal.UpdatedAt,
		); err != nil {
			return nil, err
		}
		journals = append(journals, journal)
	}

	return journals, nil
}

func (r *journalRepository) FindByID(ctx context.Context, id int) (*model.Journal, error) {
	var journal model.Journal
	query := `SELECT id, user_id, judul, isi, perasaan, created_at, updated_at 
	         FROM journals 
	         WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(
		&journal.ID,
		&journal.UserID,
		&journal.Judul,
		&journal.Isi,
		&journal.Perasaan,
		&journal.CreatedAt,
		&journal.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &journal, nil
}

func (r *journalRepository) Update(ctx context.Context, journal *model.Journal) error {
	// Set updated_at ke waktu sekarang
	journal.UpdatedAt = time.Now()

	query := `UPDATE journals 
	         SET judul = $1, isi = $2, perasaan = $3, updated_at = $4 
	         WHERE id = $5
	         RETURNING created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		journal.Judul,
		journal.Isi,
		journal.Perasaan,
		journal.UpdatedAt,
		journal.ID,
	).Scan(&journal.CreatedAt, &journal.UpdatedAt)
}

func (r *journalRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM journals WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
