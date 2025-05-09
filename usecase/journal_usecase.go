package usecase

import (
	"context"
	"pijar/model"
	"pijar/repository"
)

type JournalUsecase interface {
	Create(ctx context.Context, journal *model.Journal) error
	FindAll(ctx context.Context) ([]model.Journal, error)
	FindByUserID(ctx context.Context, userID int) ([]model.Journal, error)
	FindByID(ctx context.Context, id int) (*model.Journal, error)
	Update(ctx context.Context, journal *model.Journal) error
	Delete(ctx context.Context, id int) error
}

type journalUsecase struct {
	repo repository.JournalRepository
}

func NewJournalUsecase(repo repository.JournalRepository) JournalUsecase {
	return &journalUsecase{repo: repo}
}

func (u *journalUsecase) Create(ctx context.Context, journal *model.Journal) error {
	return u.repo.Create(ctx, journal)
}

func (u *journalUsecase) FindAll(ctx context.Context) ([]model.Journal, error) {
	return u.repo.FindAll(ctx)
}

func (u *journalUsecase) FindByUserID(ctx context.Context, userID int) ([]model.Journal, error) {
	return u.repo.FindByUserID(ctx, userID)
}

func (u *journalUsecase) FindByID(ctx context.Context, id int) (*model.Journal, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *journalUsecase) Update(ctx context.Context, journal *model.Journal) error {
	return u.repo.Update(ctx, journal)
}

func (u *journalUsecase) Delete(ctx context.Context, id int) error {
	return u.repo.Delete(ctx, id)
}
