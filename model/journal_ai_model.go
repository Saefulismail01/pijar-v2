package model

import (
	"time"
)

// JournalAnalysis menyimpan hasil analisis AI terhadap journal entry
type JournalAnalysis struct {
	ID              int      `json:"id" gorm:"primaryKey"`
	UserID          int       `json:"user_id"`
	JournalID       int      `json:"journal_id"`      // Foreign key ke journal entry yang ada
	SentimentScore  float64   `json:"sentiment_score"` // -1.0 (negative) to 1.0 (positive)
	Emotions        string    `json:"emotions"`        // JSON array string
	Keywords        string    `json:"keywords"`        // JSON array string
	Themes          string    `json:"themes"`          // JSON array string
	Insights        string    `json:"insights"`        // AI-generated insights
	Recommendations string    `json:"recommendations"` // AI suggestions
	AnalyzedAt      time.Time `json:"analyzed_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TrendAnalysis untuk analisis trend jangka panjang
type TrendAnalysis struct {
	ID               int      `json:"id" gorm:"primaryKey"`
	UserID           int      `json:"user_id"`
	PeriodType       string    `json:"period_type"` // "weekly", "monthly"
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	AverageSentiment float64   `json:"average_sentiment"`
	TopEmotions      string    `json:"top_emotions"` // JSON array
	KeyThemes        string    `json:"key_themes"`   // JSON array
	MoodTrend        string    `json:"mood_trend"`   // "improving", "declining", "stable"
	TrendInsights    string    `json:"trend_insights"`
	ProgressNotes    string    `json:"progress_notes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AnalysisRequest untuk request analisis
type AnalysisRequest struct {
	JournalID int   `json:"journal_id"`
	UserID    int   `json:"user_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Feeling   string `json:"feeling"`
}

// AnalysisResponse untuk response analisis
type AnalysisResponse struct {
	JournalAnalysis *JournalAnalysis `json:"analysis"`
	Summary         string           `json:"summary"`
	ActionItems     []string         `json:"action_items"`
}

// TrendRequest untuk request trend analysis
type TrendRequest struct {
	UserID     int   `json:"user_id"`
	PeriodType string `json:"period_type"` // "weekly", "monthly"
	Days       int    `json:"days"`        // berapa hari ke belakang
}

// TrendResponse untuk response trend analysis
type TrendResponse struct {
	TrendAnalysis   *TrendAnalysis         `json:"trend_analysis"`
	ComparisonData  map[string]interface{} `json:"comparison_data"`
	Recommendations []string               `json:"recommendations"`
	Charts          map[string]interface{} `json:"charts"` // Data untuk chart visualization
}
