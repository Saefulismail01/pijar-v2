package usecase

import (
	"context"
	"pijar/model"
	"pijar/repository"
)

type TopicUsecase interface {
	CreateTopic(ctx context.Context, userID int, preference string) (int, error)
	GetTopicByID(ctx context.Context, id int) (*model.TopicUser, error)
	GetAllTopics(ctx context.Context) ([]model.TopicUser, error)
	UpdateTopic(ctx context.Context, id int, preference string) error
	DeleteTopic(ctx context.Context, id int) error
}

type topicUsecase struct {
	topicRepo repository.TopicUserRepository
}

func NewTopicUsecase(topicRepo repository.TopicUserRepository) TopicUsecase {
	return &topicUsecase{topicRepo: topicRepo}
}

func (uc *topicUsecase) CreateTopic(ctx context.Context, userID int, preference string) (int, error) {
	return uc.topicRepo.CreateTopicUser(ctx, userID, preference)
}

func (uc *topicUsecase) GetTopicByID(ctx context.Context, id int) (*model.TopicUser, error) {
	topics, err := uc.topicRepo.GetTopicByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(topics) == 0 {
		return nil, nil
	}
	return &topics[0], nil
}

func (uc *topicUsecase) GetAllTopics(ctx context.Context) ([]model.TopicUser, error) {
	return uc.topicRepo.GetAllTopicUsers(ctx)
}

func (uc *topicUsecase) UpdateTopic(ctx context.Context, id int, preference string) error {
	return uc.topicRepo.UpdateTopicUser(ctx, id, preference)
}

func (uc *topicUsecase) DeleteTopic(ctx context.Context, id int) error {
	return uc.topicRepo.DeleteTopicUser(ctx, id)
}
