package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"pijar/model"
	"pijar/model/dto"
	"slices"
	"time"

	"github.com/lib/pq"
)

type DailyGoalRepository interface {
	CreateGoal(ctx context.Context, goal *model.UserGoal, articlesToRead []int64) (model.UserGoal, error)
	GetGoalByID(ctx context.Context, id int, userID int) (model.UserGoal, error)
	GetGoalsByUserID(ctx context.Context, userID int) ([]model.UserGoal, error)
	GetGoalProgress(ctx context.Context, goalID int, userID int) ([]dto.ArticleProgress, error)
	UpdateGoal(ctx context.Context, goal *model.UserGoal, articlesToRead []int64, userID int) (model.UserGoal, error)
	UpdateGoalStatus(ctx context.Context, goalID int, userID int) error
	CompleteArticleProgress(ctx context.Context, goalID int, articleID int64, completed bool) error
	CountCompletedProgress(ctx context.Context, goalID int, userID int) (int, error)
	DeleteGoal(ctx context.Context, goalID int, userID int) error
	ValidateArticleIDs(ctx context.Context, articleIDs []int64) ([]int64, error)
}

type dailyGoalsRepository struct {
	db *sql.DB
}

func NewDailyGoalsRepository(db *sql.DB) DailyGoalRepository {
	return &dailyGoalsRepository{db: db}
}

func (r *dailyGoalsRepository) CreateGoal(ctx context.Context, goal *model.UserGoal, articleToRead []int64) (model.UserGoal, error) {
	// start db transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.UserGoal{}, fmt.Errorf("failed to begin transaction: %v", err)
	}

	// insert goal
	goalsQuery := `
        INSERT INTO user_goals
        (title, task, articles_to_read, user_id, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at
    `
	// execute insert goal and scan the ID and CreatedAt (RETURNING id, created_at)
	err = tx.QueryRow(
		goalsQuery,
		goal.Title,
		goal.Task,
		pq.Array(goal.ArticlesToRead),
		goal.UserID,
		time.Now(),
	).Scan(&goal.ID, &goal.CreatedAt)

	if err != nil {
		tx.Rollback()
		return model.UserGoal{}, fmt.Errorf("failed to create goal: %v", err)
	}

	// insert progress
	progressQuery := `INSERT INTO user_goals_progress 
        (id_goals, id_article, date_completed, completed)
        VALUES ($1, $2, $3, $4)`

	// insert progress records for articles if there are any
	for _, articleID := range goal.ArticlesToRead {
		_, err = tx.Exec(
			progressQuery,
			goal.ID,
			articleID,
			nil,
			false,
		)
		if err != nil {
			tx.Rollback() // rollback transaction if error
			return model.UserGoal{}, fmt.Errorf("failed to create progress: %v", err)
		}
	}

	// commit transaction if all operations succeed
	err = tx.Commit()
	if err != nil {
		return model.UserGoal{}, fmt.Errorf("failed to create progres: %v", err)
	}

	return *goal, nil

}

func (r *dailyGoalsRepository) UpdateGoal(
	ctx context.Context,
	goal *model.UserGoal,
	articleToRead []int64,
	userID int,
) (model.UserGoal, error) {
	// First, check if the goal exists and belongs to the user
	_, err := r.GetGoalByID(ctx, goal.ID, userID)
	if err != nil {
		return model.UserGoal{}, fmt.Errorf("failed to get existing goal: %v", err)
	}

	// start db transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.UserGoal{}, fmt.Errorf("failed to begin transaction: %v", err)
	}

	var existingID int
	err = tx.QueryRowContext(
		ctx,
		"SELECT id FROM user_goals WHERE id = $1 AND user_id = $2",
		goal.ID,
		userID,
	).Scan(&existingID)
	if err != nil {
		tx.Rollback()
		return model.UserGoal{}, fmt.Errorf("goal not found: %v", err)
	}

	updateGoalQuery := `
        UPDATE user_goals 
        SET title = $1, task = $2, completed = $3 
        WHERE id = $4 AND user_id = $5
        RETURNING id, title, task, completed, articles_to_read, user_id, created_at
    `
	err = tx.QueryRowContext(
		ctx,
		updateGoalQuery,
		goal.Title,
		goal.Task,
		goal.Completed,
		goal.ID,
		userID,
	).Scan(
		&goal.ID,
		&goal.Title,
		&goal.Task,
		&goal.Completed,
		pq.Array(&goal.ArticlesToRead),
		&goal.UserID,
		&goal.CreatedAt,
	)

	if err != nil {
		tx.Rollback()
		return model.UserGoal{}, fmt.Errorf("failed to update goal: %v", err)
	}

	insertProgressQuery := `
        INSERT INTO user_goals_progress (id_goals, id_article, completed, date_completed)
        SELECT $1, article_id, false, NULL
        FROM unnest($2::bigint[]) AS article_id
        ON CONFLICT (id_goals, id_article) DO NOTHING
    `
	_, err = tx.ExecContext(ctx, insertProgressQuery, goal.ID, pq.Array(articleToRead))
	if err != nil {
		tx.Rollback()
		return model.UserGoal{}, fmt.Errorf("gagal insert progress baru: %v", err)
	}

	if len(articleToRead) > 0 {
		updateArticlesQuery := `
            UPDATE user_goals 
            SET articles_to_read = $1
            WHERE id = $2 AND user_id = $3
            RETURNING articles_to_read
        `
		err = tx.QueryRowContext(
			ctx,
			updateArticlesQuery,
			pq.Array(articleToRead),
			goal.ID,
			userID,
		).Scan(pq.Array(&goal.ArticlesToRead))
		if err != nil {
			tx.Rollback()
			return model.UserGoal{}, fmt.Errorf("failed to update articles: %v", err)
		}
	}
	// delete article progress doesnt exists in new list
	if articleToRead != nil {
		deleteProgressQuery := `
        DELETE FROM user_goals_progress 
        WHERE id_goals = $1 
        AND id_article NOT IN (SELECT unnest($2::bigint[]))
    `
		_, err = tx.ExecContext(
			ctx,
			deleteProgressQuery,
			goal.ID,
			pq.Array(articleToRead),
		)
		if err != nil {
			tx.Rollback()
			return model.UserGoal{}, fmt.Errorf("failed to clean progress: %v", err)
		}

		// Update goal status based on artikel
		var completedCount int
		err = tx.QueryRowContext(
			ctx,
			`SELECT COUNT(*) FROM user_goals_progress 
             WHERE id_goals = $1 AND completed = true`,
			goal.ID,
		).Scan(&completedCount)
		if err != nil {
			tx.Rollback()
			return model.UserGoal{}, err
		}

		// use length articleToRead as total article
		totalArticles := len(articleToRead)

		// Hitung status baru dan update ke DB
		newStatus := (completedCount == totalArticles && totalArticles > 0)
		_, err = tx.ExecContext(
			ctx,
			"UPDATE user_goals SET completed = $1 WHERE id = $2",
			newStatus,
			goal.ID,
		)
		if err != nil {
			tx.Rollback()
			return model.UserGoal{}, err
		}
		// Sync juga ke struct goal
		goal.Completed = newStatus

	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return model.UserGoal{}, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return *goal, nil
}

func (r *dailyGoalsRepository) GetGoalsByUserID(ctx context.Context, userID int) ([]model.UserGoal, error) {
	query := `
        SELECT id, user_id, title, task, articles_to_read, completed, created_at 
        FROM user_goals 
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goals: %v", err)
	}
	defer rows.Close()

	var goals []model.UserGoal
	for rows.Next() {
		var goal model.UserGoal
		err := rows.Scan(
			&goal.ID,
			&goal.UserID,
			&goal.Title,
			&goal.Task,
			pq.Array(&goal.ArticlesToRead),
			&goal.Completed,
			&goal.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan goal: %v", err)
		}
		goals = append(goals, goal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating goals: %v", err)
	}

	return goals, nil
}

func (r *dailyGoalsRepository) GetGoalByID(ctx context.Context, goalID int, userID int) (model.UserGoal, error) {
	query := `
        SELECT id, user_id, title, task, articles_to_read, completed, created_at 
        FROM user_goals 
        WHERE id = $1 AND user_id = $2
    `

	log.Printf("Executing query: %s with goalID=%d, userID=%d", query, goalID, userID)

	var goal model.UserGoal
	err := r.db.QueryRowContext(ctx, query, goalID, userID).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.Title,
		&goal.Task,
		pq.Array(&goal.ArticlesToRead),
		&goal.Completed,
		&goal.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return model.UserGoal{}, fmt.Errorf("goal not found")
		}
		return model.UserGoal{}, fmt.Errorf("failed to get goal: %v", err)
	}

	return goal, nil
}

func (r *dailyGoalsRepository) CompleteArticleProgress(ctx context.Context, goalID int, articleID int64, completed bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Dapatkan data goal untuk memverifikasi artikel
	var articles []int64
	err = tx.QueryRowContext(
		ctx,
		`SELECT articles_to_read FROM user_goals WHERE id = $1`,
		goalID,
	).Scan(pq.Array(&articles))

	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return fmt.Errorf("goal not found")
		}
		return fmt.Errorf("failed to get goal: %v", err)
	}

	// Verifikasi artikel termasuk dalam goal
	found := slices.Contains(articles, articleID)

	if !found {
		tx.Rollback()
		return fmt.Errorf("article %d not found in goal %d", articleID, goalID)
	}

	// Update progress artikel
	query := `
        INSERT INTO user_goals_progress (id_goals, id_article, completed, date_completed)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id_goals, id_article) 
        DO UPDATE SET completed = $3, date_completed = $4
    `

	_, err = tx.ExecContext(
		ctx,
		query,
		goalID,
		articleID,
		completed,
		time.Now(),
	)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update article progress: %v", err)
	}

	// Hitung progress
	var completedCount int
	err = tx.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM user_goals_progress 
		  WHERE id_goals = $1 AND completed = true`,
		goalID,
	).Scan(&completedCount)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to count completed articles: %v", err)
	}

	// Update status completed di user_goals
	_, err = tx.ExecContext(
		ctx,
		`UPDATE user_goals 
		  SET completed = $1 
		  WHERE id = $2`,
		completedCount == len(articles),
		goalID,
	)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update goal status: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (r *dailyGoalsRepository) DeleteGoal(ctx context.Context, goalID int, userID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Delete progress (child table)
	deleteProgressQuery := `
        DELETE FROM user_goals_progress 
        WHERE id_goals = $1
    `
	_, err = tx.ExecContext(ctx, deleteProgressQuery, goalID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete progress: %v", err)
	}

	// Delete goal (parent table)
	deleteGoalQuery := `
        DELETE FROM user_goals 
        WHERE id = $1 AND user_id = $2
    `
	result, err := tx.ExecContext(ctx, deleteGoalQuery, goalID, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete goal: %v", err)
	}

	// Check if any row was affected
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("goal not found or access denied")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// HELPER FUNCTION =================================

func (r *dailyGoalsRepository) CountCompletedProgress(ctx context.Context, goalID int, userID int) (int, error) {
	// First verify the goal exists and belongs to the user
	_, err := r.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to verify goal: %v", err)
	}

	// Count completed progress records for this goal
	query := `
        SELECT COUNT(*) 
        FROM user_goals_progress 
        WHERE id_goals = $1 AND completed = true
    `

	var count int
	err = r.db.QueryRowContext(ctx, query, goalID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count completed progress: %v", err)
	}

	// Log untuk debugging
	log.Printf("Completed count for goal %d: %d", goalID, count)

	return count, nil
}

// konversi pq.Int64Array ke []int
func (r *dailyGoalsRepository) GetGoalProgress(ctx context.Context, goalID int, userID int) ([]dto.ArticleProgress, error) {
	query := `
        SELECT 
            COALESCE(p.id_article, article_id) as id_article,
            COALESCE(p.completed, false) as completed,
            p.date_completed
        FROM user_goals g
        CROSS JOIN UNNEST(g.articles_to_read) as article_id
        LEFT JOIN user_goals_progress p ON p.id_goals = g.id 
            AND p.id_article = article_id
        WHERE g.id = $1 AND g.user_id = $2
        ORDER BY article_id
    `

	rows, err := r.db.QueryContext(ctx, query, goalID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal progress: %v", err)
	}
	defer rows.Close()

	var progress []dto.ArticleProgress
	for rows.Next() {
		var p dto.ArticleProgress
		var dateCompleted sql.NullTime

		if err := rows.Scan(
			&p.ArticleID,
			&p.Completed,
			&dateCompleted,
		); err != nil {
			return nil, fmt.Errorf("failed to scan progress row: %v", err)
		}

		if dateCompleted.Valid {
			p.DateCompleted = dateCompleted.Time.Format("2006-01-02 15:04:05")
		}

		progress = append(progress, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating progress rows: %v", err)
	}

	return progress, nil
}

func (r *dailyGoalsRepository) ValidateArticleIDs(ctx context.Context, articleIDs []int64) ([]int64, error) {
	var invalidIDs []int64

	// find non existing id
	query := `
        SELECT id 
        FROM unnest($1::bigint[]) AS t(id)
        WHERE NOT EXISTS (
            SELECT 1 FROM articles WHERE id = t.id
        )
    `

	rows, err := r.db.QueryContext(ctx, query, pq.Array(articleIDs))
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		invalidIDs = append(invalidIDs, id)
	}

	return invalidIDs, nil
}

func (r *dailyGoalsRepository) UpdateGoalStatus(ctx context.Context, goalID int, userID int) error {
	// Hitung progress terbaru berdasarkan articles_to_read saat ini
	completedCount, err := r.CountCompletedProgress(ctx, goalID, userID)
	if err != nil {
		return err
	}

	// Dapatkan total artikel saat ini
	var totalArticles int
	err = r.db.QueryRowContext(
		ctx,
		"SELECT cardinality(articles_to_read) FROM user_goals WHERE id = $1",
		goalID,
	).Scan(&totalArticles)
	if err != nil {
		return err
	}

	// Update status completed
	_, err = r.db.ExecContext(
		ctx,
		"UPDATE user_goals SET completed = $1 WHERE id = $2",
		completedCount == totalArticles,
		goalID,
	)
	return err
}
