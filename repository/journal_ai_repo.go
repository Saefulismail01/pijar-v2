package repository

import (
	"database/sql"
	"fmt"
	"log"
	"pijar/model"
)

type JournalAnalysisRepository struct {
	db *sql.DB
}

func NewJournalAnalysisRepository(db *sql.DB) *JournalAnalysisRepository {
	return &JournalAnalysisRepository{db: db}
}

func (r *JournalAnalysisRepository) Save(analysis *model.JournalAnalysis) error {
	query := `
		INSERT INTO journal_analyses (journal_id, user_id, sentiment_score, analyzed_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	return r.db.QueryRow(query, analysis.JournalID, analysis.UserID, analysis.SentimentScore, analysis.AnalyzedAt).Scan(&analysis.ID)
}

func (r *JournalAnalysisRepository) SaveTrend(trend *model.TrendAnalysis) error {
	query := `
		INSERT INTO trend_analyses (user_id, period_type, period_start, period_end, average_sentiment, top_emotions, key_themes, mood_trend, trend_insights, progress_notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		trend.UserID,
		trend.PeriodType,
		trend.PeriodStart,
		trend.PeriodEnd,
		trend.AverageSentiment,
		trend.TopEmotions,
		trend.KeyThemes,
		trend.MoodTrend,
		trend.TrendInsights,
		trend.ProgressNotes,
	).Scan(&trend.ID, &trend.CreatedAt, &trend.UpdatedAt)
}

func (r *JournalAnalysisRepository) GetByJournalID(journalID int) (*model.JournalAnalysis, error) {
	query := `
		SELECT id, journal_id, user_id, sentiment_score, analyzed_at
		FROM journal_analyses
		WHERE journal_id = $1
		LIMIT 1
	`
	analysis := &model.JournalAnalysis{}
	err := r.db.QueryRow(query, journalID).Scan(&analysis.ID, &analysis.JournalID, &analysis.UserID, &analysis.SentimentScore, &analysis.AnalyzedAt)
	if err != nil {
		return nil, err
	}
	return analysis, nil
}

func (r *JournalAnalysisRepository) GetByUserID(userID int, limit int) ([]*model.JournalAnalysis, error) {
	log.Printf("Executing GetByUserID with userID: %d, limit: %d", userID, limit)
	
	query := `
		SELECT id, journal_id, user_id, sentiment_score, analyzed_at
		FROM journal_analyses
		WHERE user_id = $1
		ORDER BY analyzed_at DESC
		LIMIT $2
	`
	log.Printf("SQL Query: %s, params: [%d, %d]", query, userID, limit)
	
	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var analyses []*model.JournalAnalysis
	for rows.Next() {
		analysis := &model.JournalAnalysis{}
		err := rows.Scan(&analysis.ID, &analysis.JournalID, &analysis.UserID, &analysis.SentimentScore, &analysis.AnalyzedAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		log.Printf("Found analysis: ID=%d, JournalID=%d, UserID=%d, Score=%.2f, Date=%v", 
			analysis.ID, analysis.JournalID, analysis.UserID, analysis.SentimentScore, analysis.AnalyzedAt)
		analyses = append(analyses, analysis)
	}
	
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, err
	}
	
	log.Printf("Found %d analyses for user %d", len(analyses), userID)
	return analyses, nil
}

func (r *JournalAnalysisRepository) GetTrendsByUserID(userID int, periodType string) ([]*model.TrendAnalysis, error) {
	query := `
		SELECT id, user_id, period_type, period_start, period_end, average_sentiment, 
			top_emotions, key_themes, mood_trend, trend_insights, progress_notes, created_at, updated_at
		FROM trend_analyses
		WHERE user_id = $1 AND period_type = $2
		ORDER BY period_start DESC
	`
	rows, err := r.db.Query(query, userID, periodType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []*model.TrendAnalysis
	for rows.Next() {
		trend := &model.TrendAnalysis{}
		err := rows.Scan(
			&trend.ID,
			&trend.UserID,
			&trend.PeriodType,
			&trend.PeriodStart,
			&trend.PeriodEnd,
			&trend.AverageSentiment,
			&trend.TopEmotions,
			&trend.KeyThemes,
			&trend.MoodTrend,
			&trend.TrendInsights,
			&trend.ProgressNotes,
			&trend.CreatedAt,
			&trend.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		trends = append(trends, trend)
	}
	return trends, nil
}

func (r *JournalAnalysisRepository) GetAnalysisWithJournal(userID int, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT ja.id, ja.journal_id, ja.user_id, ja.sentiment_score, ja.analyzed_at,
		       j.title, j.content, j.feeling
		FROM journal_analyses ja
		LEFT JOIN journals j ON j.id = ja.journal_id
		WHERE ja.user_id = $1
		ORDER BY ja.analyzed_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var (
			id, journalID, userID   uint
			sentimentScore          float64
			analyzedAt              string
			title, content, feeling sql.NullString
		)

		err := rows.Scan(&id, &journalID, &userID, &sentimentScore, &analyzedAt, &title, &content, &feeling)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"id":              id,
			"journal_id":      journalID,
			"user_id":         userID,
			"sentiment_score": sentimentScore,
			"analyzed_at":     analyzedAt,
			"title":           title.String,
			"content":         content.String,
			"feeling":         feeling.String,
		}
		results = append(results, result)
	}
	return results, nil
}

func (r *JournalAnalysisRepository) GetSentimentTrend(userID int, days int) ([]map[string]interface{}, error) {
	query := `
		SELECT DATE(analyzed_at) AS date, 
		       AVG(sentiment_score) AS avg_sentiment, 
		       COUNT(*) AS entry_count
		FROM journal_analyses
		WHERE user_id = $1 AND analyzed_at >= NOW() - INTERVAL '$2 days'
		GROUP BY DATE(analyzed_at)
		ORDER BY date ASC
	`

	// PostgreSQL parameter untuk INTERVAL tidak bisa langsung $2,
	// jadi string formatting digunakan dengan hati-hati.
	query = fmt.Sprintf(`
		SELECT DATE(analyzed_at) AS date, 
		       AVG(sentiment_score) AS avg_sentiment, 
		       COUNT(*) AS entry_count
		FROM journal_analyses
		WHERE user_id = $1 AND analyzed_at >= NOW() - INTERVAL '%d days'
		GROUP BY DATE(analyzed_at)
		ORDER BY date ASC
	`, days)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var date string
		var avgSentiment float64
		var entryCount int

		err := rows.Scan(&date, &avgSentiment, &entryCount)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"date":          date,
			"avg_sentiment": avgSentiment,
			"entry_count":   entryCount,
		}
		results = append(results, result)
	}
	return results, nil
}

func (r *JournalAnalysisRepository) DeleteAnalysis(id int) error {
	query := `DELETE FROM journal_analyses WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *JournalAnalysisRepository) UpdateAnalysis(analysis *model.JournalAnalysis) error {
	query := `
		UPDATE journal_analyses
		SET journal_id = $1, user_id = $2, sentiment_score = $3, analyzed_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(query, analysis.JournalID, analysis.UserID, analysis.SentimentScore, analysis.AnalyzedAt, analysis.ID)
	return err
}
