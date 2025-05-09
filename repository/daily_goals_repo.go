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
	CompleteArticleProgress(ctx context.Context, goalID int, articleID int64, completed bool) error
	CountCompletedProgress(ctx context.Context, goalID int, userID int) (int, error)
	DeleteGoal(ctx context.Context, goalID int, userID int) error
	ValidateArticleIDs(ctx context.Context, articleIDs []int64) ([]int64, error)
	GetPendingArticlesCount(userID int) (int, error)
	GetAllUsersWithPendingArticles() ([]model.Users, error)
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
			time.Now(),
			false,
		)
		if err != nil {
			tx.Rollback() // rollback transaction if error
			return model.UserGoal{}, fmt.Errorf("failed to create progress: %v", err)
		}
	}
	// If there are no articles, we don't create any progress records

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

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.UserGoal{}, fmt.Errorf("failed to begin transaction: %v", err)
	}

	// 1. Update goal in user_goals table
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
		userID, // Ensure user owns the goal
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

	// 2. Update articles_to_read in user_goals table
	if articleToRead != nil {
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

func (r *dailyGoalsRepository) GetPendingArticlesCount(userID int) (int, error) {
	var count int
	err := r.db.QueryRow(`
        SELECT COUNT(*) 
        FROM (
            SELECT ug.id AS goal_id, unnest(ug.articles_to_read) AS article_id
            FROM user_goals ug
            WHERE ug.user_id = $1
        ) goals
        LEFT JOIN user_goals_progress ugp 
            ON goals.goal_id = ugp.id_goals 
            AND goals.article_id = ugp.id_article
        WHERE ugp.completed IS NULL OR ugp.completed = false`,
		userID).Scan(&count)
	return count, err
}

func (r *dailyGoalsRepository) GetAllUsersWithPendingArticles() ([]model.Users, error) {
	rows, err := r.db.Query(`
        SELECT DISTINCT u.id, u.name 
        FROM users u
        JOIN user_goals ug ON u.id = ug.user_id
        WHERE EXISTS (
            SELECT 1
            FROM unnest(ug.articles_to_read) article_id
            LEFT JOIN user_goals_progress ugp 
                ON ug.id = ugp.id_goals 
                AND article_id = ugp.id_article
            WHERE ugp.completed IS NULL OR ugp.completed = false
        )`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.Users
	for rows.Next() {
		var user model.Users
		if err := rows.Scan(&user.ID, &user.Name); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
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
