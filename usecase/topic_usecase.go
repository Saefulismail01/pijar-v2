package usecase

import (
	"context"
	"pijar/model"
	"pijar/repository"
)

type TopicUserUsecase interface {
	CreateTopicUser(ctx context.Context, userID int, preference string) (int, error)
	GetTopicByID(ctx context.Context, userID int) ([]model.TopicUser, error)
	GetAllTopicUsers(ctx context.Context) ([]model.TopicUser, error)
	UpdateTopicUser(ctx context.Context, id int, preference string) error
	DeleteTopicUser(ctx context.Context, id int) error
}

type topicUserUsecase struct {
	repo repository.TopicUserRepository
}

func (u *topicUserUsecase) CreateTopicUser(ctx context.Context, userID int, preference string) (int, error) {
	return u.repo.CreateTopicUser(ctx, userID, preference)
}

func (u *topicUserUsecase) GetTopicByID(ctx context.Context, userID int) ([]model.TopicUser, error) {
	return u.repo.GetTopicByID(ctx, userID)
}

func (u *topicUserUsecase) GetAllTopicUsers(ctx context.Context) ([]model.TopicUser, error) {
	return u.repo.GetAllTopicUsers(ctx)
}

func (u *topicUserUsecase) UpdateTopicUser(ctx context.Context, id int, preference string) error {
	return u.repo.UpdateTopicUser(ctx, id, preference)
}

func (u *topicUserUsecase) DeleteTopicUser(ctx context.Context, id int) error {
	return u.repo.DeleteTopicUser(ctx, id)
}

func NewTopicUsecase(repo repository.TopicUserRepository) TopicUserUsecase {
	return &topicUserUsecase{repo: repo}
}
