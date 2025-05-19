package service

import (
	"encoding/json"
	"fmt"
	"pijar/model"
	"sort"
	"strings"
	"time"
)

type JournalAnalysisService struct {
	aiClient AIClient // Interface untuk AI client (Gemini/DeepSeek/etc)
	repo     JournalAnalysisRepository
}

// Interface untuk AI client
type AIClient interface {
	GetAIResponse(prompt string) (string, error)
}

// Interface untuk repository
type JournalAnalysisRepository interface {
	Save(analysis *model.JournalAnalysis) error
	SaveTrend(trend *model.TrendAnalysis) error
	GetByJournalID(journalID int) (*model.JournalAnalysis, error)
	GetByUserID(userID int, limit int) ([]*model.JournalAnalysis, error)
	GetTrendsByUserID(userID int, periodType string) ([]*model.TrendAnalysis, error)
}

func NewJournalAnalysisService(aiClient AIClient, repo JournalAnalysisRepository) *JournalAnalysisService {
	return &JournalAnalysisService{
		aiClient: aiClient,
		repo:     repo,
	}
}

// AnalyzeJournalEntry menganalisis single journal entry
func (j *JournalAnalysisService) AnalyzeJournalEntry(req *model.AnalysisRequest) (*model.AnalysisResponse, error) {
	// Cek apakah sudah pernah dianalisis
	existing, _ := j.repo.GetByJournalID(req.JournalID)
	if existing != nil {
		return &model.AnalysisResponse{
			JournalAnalysis: existing,
			Summary:         "Analysis already exists",
		}, nil
	}

	// Prompt untuk AI analysis
	prompt := j.buildAnalysisPrompt(req)

	// Kirim ke AI
	response, err := j.aiClient.GetAIResponse(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI analysis: %w", err)
	}

	// Parse AI response
	analysis, err := j.parseAIResponse(response, req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Simpan ke database
	if err := j.repo.Save(analysis); err != nil {
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	// Generate action items
	actionItems := j.generateActionItems(analysis)

	return &model.AnalysisResponse{
		JournalAnalysis: analysis,
		Summary:         j.generateSummary(analysis),
		ActionItems:     actionItems,
	}, nil
}

// buildAnalysisPrompt membuat prompt untuk AI analysis
func (j *JournalAnalysisService) buildAnalysisPrompt(req *model.AnalysisRequest) string {
	return fmt.Sprintf(`
Analyze this journal entry and provide structured insights:

Title: %s
Content: %s
User's Feeling: %s

Please provide your analysis in the following JSON format:
{
  "sentiment_score": -0.5 to 1.0,
  "emotions": ["emotion1", "emotion2", "emotion3"],
  "keywords": ["keyword1", "keyword2", "keyword3"],
  "themes": ["theme1", "theme2"],
  "insights": "Brief psychological insight about this entry",
  "recommendations": "Specific actionable suggestions for the user"
}

Focus on:
1. Overall emotional tone and sentiment
2. Dominant emotions (max 3)
3. Key themes and topics
4. Psychological patterns or concerns
5. Constructive suggestions for mental wellness

Keep insights empathetic, non-judgmental, and focused on growth.`,
		req.Title, req.Content, req.Feeling)
}

// parseAIResponse parsing response dari AI menjadi JournalAnalysis
func (j *JournalAnalysisService) parseAIResponse(response string, req *model.AnalysisRequest) (*model.JournalAnalysis, error) {
	// Extract JSON dari response (AI mungkin menambahkan text lain)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}") + 1

	if jsonStart == -1 || jsonEnd == 0 {
		return nil, fmt.Errorf("no valid JSON found in AI response")
	}

	jsonResponse := response[jsonStart:jsonEnd]

	// Parse JSON
	var aiResult struct {
		SentimentScore  float64  `json:"sentiment_score"`
		Emotions        []string `json:"emotions"`
		Keywords        []string `json:"keywords"`
		Themes          []string `json:"themes"`
		Insights        string   `json:"insights"`
		Recommendations string   `json:"recommendations"`
	}

	if err := json.Unmarshal([]byte(jsonResponse), &aiResult); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON: %w", err)
	}

	// Convert arrays to JSON strings for storage
	emotionsJSON, _ := json.Marshal(aiResult.Emotions)
	keywordsJSON, _ := json.Marshal(aiResult.Keywords)
	themesJSON, _ := json.Marshal(aiResult.Themes)

	return &model.JournalAnalysis{
		UserID:          req.UserID,
		JournalID:       req.JournalID,
		SentimentScore:  aiResult.SentimentScore,
		Emotions:        string(emotionsJSON),
		Keywords:        string(keywordsJSON),
		Themes:          string(themesJSON),
		Insights:        aiResult.Insights,
		Recommendations: aiResult.Recommendations,
		AnalyzedAt:      time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

// generateActionItems membuat action items berdasarkan analysis
func (j *JournalAnalysisService) generateActionItems(analysis *model.JournalAnalysis) []string {
	var actions []string

	// Based on sentiment score
	if analysis.SentimentScore < -0.3 {
		actions = append(actions, "Consider practicing gratitude journaling")
		actions = append(actions, "Reach out to a trusted friend or counselor")
	} else if analysis.SentimentScore > 0.5 {
		actions = append(actions, "Reflect on what contributed to this positive state")
		actions = append(actions, "Document successful coping strategies")
	}

	// Based on emotions
	var emotions []string
	json.Unmarshal([]byte(analysis.Emotions), &emotions)

	for _, emotion := range emotions {
		switch strings.ToLower(emotion) {
		case "anxiety", "worry", "stress":
			actions = append(actions, "Try deep breathing or meditation exercises")
		case "sadness", "depression":
			actions = append(actions, "Engage in physical activity or creative expression")
		case "anger", "frustration":
			actions = append(actions, "Practice mindful observation of triggers")
		}
	}

	return actions
}

// generateSummary membuat summary singkat
func (j *JournalAnalysisService) generateSummary(analysis *model.JournalAnalysis) string {
	sentiment := "neutral"
	if analysis.SentimentScore > 0.3 {
		sentiment = "positive"
	} else if analysis.SentimentScore < -0.3 {
		sentiment = "negative"
	}

	var emotions []string
	json.Unmarshal([]byte(analysis.Emotions), &emotions)
	emotionStr := strings.Join(emotions, ", ")

	return fmt.Sprintf("Overall sentiment: %s (%.2f). Dominant emotions: %s",
		sentiment, analysis.SentimentScore, emotionStr)
}

// GenerateTrendAnalysis membuat analisis trend untuk periode tertentu
func (j *JournalAnalysisService) GenerateTrendAnalysis(req *model.TrendRequest) (*model.TrendResponse, error) {
	// Ambil analyses dalam periode yang diminta
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -req.Days)

	analyses, err := j.repo.GetByUserID(req.UserID, req.Days)
	if err != nil {
		return nil, fmt.Errorf("failed to get analyses: %w", err)
	}

	if len(analyses) == 0 {
		return nil, fmt.Errorf("no journal analyses found for trend analysis")
	}

	// Calculate trend metrics
	trendAnalysis := j.calculateTrendMetrics(analyses, req)
	trendAnalysis.UserID = req.UserID
	trendAnalysis.PeriodStart = startDate
	trendAnalysis.PeriodEnd = endDate
	trendAnalysis.CreatedAt = time.Now()
	trendAnalysis.UpdatedAt = time.Now()

	// Save trend analysis
	if err := j.repo.SaveTrend(trendAnalysis); err != nil {
		return nil, fmt.Errorf("failed to save trend analysis: %w", err)
	}

	// Generate comparison data dan charts
	comparisonData := j.generateComparisonData(analyses)
	charts := j.generateChartData(analyses)
	recommendations := j.generateTrendRecommendations(trendAnalysis)

	return &model.TrendResponse{
		TrendAnalysis:   trendAnalysis,
		ComparisonData:  comparisonData,
		Recommendations: recommendations,
		Charts:          charts,
	}, nil
}

// calculateTrendMetrics menghitung metrics untuk trend analysis
func (j *JournalAnalysisService) calculateTrendMetrics(analyses []*model.JournalAnalysis, req *model.TrendRequest) *model.TrendAnalysis {
	// Calculate average sentiment
	var totalSentiment float64
	emotionCounts := make(map[string]int)
	themeCounts := make(map[string]int)

	for _, analysis := range analyses {
		totalSentiment += analysis.SentimentScore

		// Count emotions
		var emotions []string
		json.Unmarshal([]byte(analysis.Emotions), &emotions)
		for _, emotion := range emotions {
			emotionCounts[emotion]++
		}

		// Count themes
		var themes []string
		json.Unmarshal([]byte(analysis.Themes), &themes)
		for _, theme := range themes {
			themeCounts[theme]++
		}
	}

	avgSentiment := totalSentiment / float64(len(analyses))

	// Get top emotions and themes
	topEmotions := j.getTopItems(emotionCounts, 3)
	topThemes := j.getTopItems(themeCounts, 3)

	// Determine mood trend
	moodTrend := j.calculateMoodTrend(analyses)

	// Generate insights
	insights := j.generateTrendInsights(avgSentiment, moodTrend, topEmotions, topThemes)

	topEmotionsJSON, _ := json.Marshal(topEmotions)
	topThemesJSON, _ := json.Marshal(topThemes)

	return &model.TrendAnalysis{
		PeriodType:       req.PeriodType,
		AverageSentiment: avgSentiment,
		TopEmotions:      string(topEmotionsJSON),
		KeyThemes:        string(topThemesJSON),
		MoodTrend:        moodTrend,
		TrendInsights:    insights,
	}
}

// getTopItems mendapatkan top N items dari map count
func (j *JournalAnalysisService) getTopItems(counts map[string]int, limit int) []string {
	type kv struct {
		key   string
		value int
	}

	var sorted []kv
	for k, v := range counts {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].value > sorted[j].value
	})

	var result []string
	for i := 0; i < limit && i < len(sorted); i++ {
		result = append(result, sorted[i].key)
	}

	return result
}

// calculateMoodTrend menentukan trend mood (improving/declining/stable)
func (j *JournalAnalysisService) calculateMoodTrend(analyses []*model.JournalAnalysis) string {
	if len(analyses) < 2 {
		return "stable"
	}

	// Sort by date (assuming newer entries have higher ID)
	sort.Slice(analyses, func(i, j int) bool {
		return analyses[i].ID < analyses[j].ID
	})

	// Compare first half with second half
	mid := len(analyses) / 2
	firstHalf := analyses[:mid]
	secondHalf := analyses[mid:]

	var firstAvg, secondAvg float64
	for _, a := range firstHalf {
		firstAvg += a.SentimentScore
	}
	firstAvg /= float64(len(firstHalf))

	for _, a := range secondHalf {
		secondAvg += a.SentimentScore
	}
	secondAvg /= float64(len(secondHalf))

	diff := secondAvg - firstAvg
	if diff > 0.1 {
		return "improving"
	} else if diff < -0.1 {
		return "declining"
	}
	return "stable"
}

// generateTrendInsights membuat insights untuk trend
func (j *JournalAnalysisService) generateTrendInsights(avgSentiment float64, moodTrend string, topEmotions, topThemes []string) string {
	emotionStr := strings.Join(topEmotions, ", ")
	themeStr := strings.Join(topThemes, ", ")

	return fmt.Sprintf(
		"Your average sentiment score is %.2f (%s trend). "+
			"Most common emotions: %s. Key themes: %s.",
		avgSentiment, moodTrend, emotionStr, themeStr)
}

// generateComparisonData membuat data perbandingan
func (j *JournalAnalysisService) generateComparisonData(analyses []*model.JournalAnalysis) map[string]interface{} {
	if len(analyses) == 0 {
		return nil
	}

	var sentiments []float64
	for _, a := range analyses {
		sentiments = append(sentiments, a.SentimentScore)
	}

	return map[string]interface{}{
		"total_entries":     len(analyses),
		"highest_sentiment": j.maxFloat64(sentiments),
		"lowest_sentiment":  j.minFloat64(sentiments),
		"avg_sentiment":     j.avgFloat64(sentiments),
	}
}

// generateChartData membuat data untuk chart visualization
func (j *JournalAnalysisService) generateChartData(analyses []*model.JournalAnalysis) map[string]interface{} {
	// Sentiment over time
	var sentimentData []map[string]interface{}
	for i, a := range analyses {
		sentimentData = append(sentimentData, map[string]interface{}{
			"entry":     i + 1,
			"sentiment": a.SentimentScore,
			"date":      a.AnalyzedAt.Format("2006-01-02"),
		})
	}

	// Emotion frequency
	emotionCounts := make(map[string]int)
	for _, a := range analyses {
		var emotions []string
		json.Unmarshal([]byte(a.Emotions), &emotions)
		for _, emotion := range emotions {
			emotionCounts[emotion]++
		}
	}

	return map[string]interface{}{
		"sentiment_timeline": sentimentData,
		"emotion_frequency":  emotionCounts,
	}
}

// generateTrendRecommendations membuat rekomendasi berdasarkan trend
func (j *JournalAnalysisService) generateTrendRecommendations(trend *model.TrendAnalysis) []string {
	var recommendations []string

	if trend.MoodTrend == "declining" {
		recommendations = append(recommendations, "Consider scheduling time with a counselor or therapist")
		recommendations = append(recommendations, "Focus on self-care activities that have helped you before")
		recommendations = append(recommendations, "Practice daily mindfulness or meditation")
	} else if trend.MoodTrend == "improving" {
		recommendations = append(recommendations, "Continue with current positive habits")
		recommendations = append(recommendations, "Document what strategies are working well")
		recommendations = append(recommendations, "Consider sharing your progress with supportive people")
	}

	if trend.AverageSentiment < 0 {
		recommendations = append(recommendations, "Include more gratitude practices in your routine")
		recommendations = append(recommendations, "Engage in activities that bring you joy")
	}

	return recommendations
}

// Helper functions
func (j *JournalAnalysisService) maxFloat64(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	max := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		}
	}
	return max
}

func (j *JournalAnalysisService) minFloat64(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	min := arr[0]
	for _, v := range arr {
		if v < min {
			min = v
		}
	}
	return min
}

func (j *JournalAnalysisService) avgFloat64(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range arr {
		sum += v
	}
	return sum / float64(len(arr))
}
