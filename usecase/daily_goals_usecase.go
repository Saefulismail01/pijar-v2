package usecase

import (
	"context"
	"fmt"
	"pijar/model"
	"pijar/model/dto"
	"pijar/repository"
)

type DailyGoalUseCase interface {
	CreateGoal(
		ctx context.Context,
		userID int,
		title string,
		task string,
		articlesToRead []int64,
	) (model.UserGoal, error)
	GetUserGoals(ctx context.Context, userID int) ([]model.UserGoal, error)
	GetGoalByID(ctx context.Context, userID int, goalID int) (model.UserGoal, error)
	UpdateGoal(
		ctx context.Context,
		userID int,
		goalID int,
		title string,
		task string,
		completed bool,
		articlesToRead []int64,
	) (dto.GoalProgressInfo, error)
	CompleteArticleProgress(
		ctx context.Context,
		goalID int,
		articleID int,
		userID int,
	) (dto.GoalProgressInfo, error)
	DeleteGoal(ctx context.Context, userID int, goalID int) error
}

type dailyGoalUseCase struct {
	repo repository.DailyGoalRepository
}

func NewGoalUseCase(repo repository.DailyGoalRepository) DailyGoalUseCase {
	return &dailyGoalUseCase{repo: repo}
}

func (uc *dailyGoalUseCase) CreateGoal(ctx context.Context, userID int, title string, task string, articlesToRead []int64) (model.UserGoal, error) {
	// validate article id
	if len(articlesToRead) > 0 {
		invalidIDs, err := uc.repo.ValidateArticleIDs(ctx, articlesToRead)
		if err != nil {
			return model.UserGoal{}, fmt.Errorf("failed to validate articles: %v", err)
		}

		if len(invalidIDs) > 0 {
			return model.UserGoal{}, fmt.Errorf("invalid article ID(s): %v", invalidIDs)
		}
	}
	// create goals fields
	newGoal := model.UserGoal{
		UserID:         userID,
		Title:          title,
		Task:           task,
		ArticlesToRead: articlesToRead,
		Completed:      false,
	}

	// create goal
	createdGoal, err := uc.repo.CreateGoal(ctx, &newGoal, articlesToRead)
	if err != nil {
		return model.UserGoal{}, fmt.Errorf("usecase error: %v", err)
	}

	return createdGoal, nil
}

func (uc *dailyGoalUseCase) CompleteArticleProgress(
	ctx context.Context,
	goalID int,
	articleID int,
	userID int,
) (dto.GoalProgressInfo, error) {
	// Get the goal first to verify it exists and belongs to the user
	goal, err := uc.repo.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return dto.GoalProgressInfo{}, fmt.Errorf("failed to get goal: %v", err)
	}

	// Complete the article progress
	err = uc.repo.CompleteArticleProgress(ctx, goalID, int64(articleID), true)
	if err != nil {
		return dto.GoalProgressInfo{}, fmt.Errorf("failed to complete article progress: %v", err)
	}

	err = uc.repo.UpdateGoalStatus(ctx, goalID, userID)
	if err != nil {
		return dto.GoalProgressInfo{}, fmt.Errorf("failed to update goal status: %v", err)
	}

	// Get the updated goal
	updatedGoal, err := uc.repo.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return dto.GoalProgressInfo{}, fmt.Errorf("failed to get updated goal: %v", err)
	}

	// Check if we need to update the main goal status
	completedCount, err := uc.repo.CountCompletedProgress(ctx, goalID, userID)
	if err != nil {
		return dto.GoalProgressInfo{}, fmt.Errorf("failed to count completed progress: %v", err)
	}

	// Log untuk debugging
	fmt.Printf("Goal %d has %d completed articles. Goal completed status: %v\n", goalID, completedCount, goal.Completed)

	err = uc.repo.UpdateGoalStatus(ctx, goalID, userID)
	if err != nil {
		return dto.GoalProgressInfo{}, fmt.Errorf("failed to update goal status: %v", err)
	}
	// Get the updated progress information
	progress, err := uc.repo.GetGoalProgress(ctx, goalID, userID)
	if err != nil {
		return dto.GoalProgressInfo{}, fmt.Errorf("failed to get goal progress: %v", err)
	}

	return dto.GoalProgressInfo{
		Goal:     updatedGoal,
		Progress: progress,
	}, nil
}

func (uc *dailyGoalUseCase) GetUserGoals(ctx context.Context, userID int) ([]model.UserGoal, error) {
	goals, err := uc.repo.GetGoalsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user goals: %v", err)
	}
	return goals, nil
}

func (uc *dailyGoalUseCase) GetGoalByID(ctx context.Context, userID int, goalID int) (model.UserGoal, error) {
	goal, err := uc.repo.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return model.UserGoal{}, fmt.Errorf("failed to get goal: %v", err)
	}
	return goal, nil
}

func (uc *dailyGoalUseCase) UpdateGoal(
	ctx context.Context,
	userID int,
	goalID int,
	title string,
	task string,
	completed bool,
	newArticlesToRead []int64,
) (dto.GoalProgressInfo, error) {
	// validate new article
	if newArticlesToRead != nil {
		invalidIDs, err := uc.repo.ValidateArticleIDs(ctx, newArticlesToRead)
		if err != nil {
			return dto.GoalProgressInfo{}, fmt.Errorf("failed to validate articles: %v", err)
		}
		if len(invalidIDs) > 0 {
			return dto.GoalProgressInfo{}, fmt.Errorf("invalid article ID(s): %v", invalidIDs)
		}
	}

	// get an existing data
	existingGoal, err := uc.repo.GetGoalByID(ctx, goalID, userID)
	if err != nil {
		return dto.GoalProgressInfo{}, fmt.Errorf("failed to get existing goal: %v", err)
	}

	// full replacement logic
	var articlesToRead []int64
	if newArticlesToRead != nil {
		articlesToRead = newArticlesToRead // new article
	} else {
		articlesToRead = existingGoal.ArticlesToRead // old article
	}

	updatedGoal := model.UserGoal{
		ID:             goalID,
		Title:          title,
		Task:           task,
		ArticlesToRead: articlesToRead,
		Completed:      completed,
	}

	result, err := uc.repo.UpdateGoal(ctx, &updatedGoal, newArticlesToRead, userID)
	if err != nil {
		return dto.GoalProgressInfo{}, fmt.Errorf("usecase error: %v", err)
	}

	// Get the progress information
	progress, err := uc.repo.GetGoalProgress(ctx, goalID, userID)
	if err != nil {
		return dto.GoalProgressInfo{}, fmt.Errorf("failed to get goal progress: %v", err)
	}

	return dto.GoalProgressInfo{
		Goal:     result,
		Progress: progress,
	}, nil
}

func (uc *dailyGoalUseCase) DeleteGoal(ctx context.Context, userID int, goalID int) error {
	// Validate input
	if userID <= 0 || goalID <= 0 {
		return fmt.Errorf("invalid userID or goalID")
	}

	err := uc.repo.DeleteGoal(ctx, goalID, userID)
	if err != nil {
		return fmt.Errorf("usecase error: %v", err)
	}

	return nil
}
