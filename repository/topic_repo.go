package repository

import (
	"context"
	"database/sql"
	"pijar/model"
)

type TopicUserRepository interface {
	CreateTopicUser(ctx context.Context, userID int, preference string) (int, error)
	GetTopicByID(ctx context.Context, userID int) ([]model.TopicUser, error)
	GetAllTopicUsers(ctx context.Context) ([]model.TopicUser, error)
	UpdateTopicUser(ctx context.Context, id int, preference string) error
	DeleteTopicUser(ctx context.Context, id int) error
}

type topicUserRepository struct {
	db *sql.DB
}

func (r *topicUserRepository) CreateTopicUser(ctx context.Context, userID int, preference string) (int, error) {
	var id int
	query := `INSERT INTO topics (user_id, preference) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, userID, preference).Scan(&id)
	return id, err
}

func (r *topicUserRepository) GetTopicByID(ctx context.Context, id int) ([]model.TopicUser, error) {
	var topicUser model.TopicUser
	query := `SELECT id, user_id, preference FROM topics WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&topicUser.ID,
		&topicUser.UserID,
		&topicUser.Preference,
	)

	if err == sql.ErrNoRows {
		return []model.TopicUser{}, nil
	}
	if err != nil {
		return nil, err
	}

	return []model.TopicUser{topicUser}, nil
}

func (r *topicUserRepository) GetAllTopicUsers(ctx context.Context) ([]model.TopicUser, error) {
	var topicUsers []model.TopicUser
	query := `SELECT id, user_id, preference FROM topics`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var topicUser model.TopicUser
		err := rows.Scan(&topicUser.ID, &topicUser.UserID, &topicUser.Preference)
		if err != nil {
			return nil, err
		}
		topicUsers = append(topicUsers, topicUser)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return topicUsers, nil
}

func (r *topicUserRepository) UpdateTopicUser(ctx context.Context, id int, preference string) error {
	query := `UPDATE topics SET preference = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, preference, id)
	return err
}

func (r *topicUserRepository) DeleteTopicUser(ctx context.Context, id int) error {
	query := `DELETE FROM topics WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func NewTopicRepository(db *sql.DB) TopicUserRepository {
	return &topicUserRepository{db: db}
}
