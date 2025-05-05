package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pijar/model"
)

type JournalRepository interface {
	Create(ctx context.Context, journal *model.Journal) error
	FindAll(ctx context.Context, userID int) ([]model.Journal, error)
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

	query = `INSERT INTO journals (user_id, judul, isi, perasaan) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		journal.UserID,
		journal.Judul,
		journal.Isi,
		journal.Perasaan,
	).Scan(&journal.ID)
}

func (r *journalRepository) FindAll(ctx context.Context, userID int) ([]model.Journal, error) {
	var journals []model.Journal
	query := `SELECT id, user_id, judul, isi, perasaan FROM journals WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var journal model.Journal
		if err := rows.Scan(&journal.ID, &journal.UserID, &journal.Judul, &journal.Isi, &journal.Perasaan); err != nil {
			return nil, err
		}
		journals = append(journals, journal)
	}

	return journals, nil
}

func (r *journalRepository) FindByID(ctx context.Context, id int) (*model.Journal, error) {
	var journal model.Journal
	query := `SELECT id, user_id, judul, isi, perasaan FROM journals WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(&journal.ID, &journal.UserID, &journal.Judul, &journal.Isi, &journal.Perasaan); err != nil {
		return nil, err
	}

	return &journal, nil
}

func (r *journalRepository) Update(ctx context.Context, journal *model.Journal) error {
	query := `UPDATE journals SET judul = $1, isi = $2, perasaan = $3 WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query,
		journal.Judul,
		journal.Isi,
		journal.Perasaan,
		journal.ID,
	)
	return err
}

func (r *journalRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM journals WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
