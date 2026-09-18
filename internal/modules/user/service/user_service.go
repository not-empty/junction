package service

import (
	"context"
	"errors"

	"github.com/not-empty/bridge/internal/modules/user/domain"
	"github.com/not-empty/bridge/internal/modules/user/repository"
	"github.com/not-empty/bridge/platform/apperror"
	"github.com/not-empty/bridge/platform/database"
)

type UserServiceInterface interface {
	Create(ctx context.Context, input domain.UserRegister) (string, error)
	List(ctx context.Context) ([]domain.User, error)
	GetOne(ctx context.Context, id string) (domain.User, error)
	Update(ctx context.Context, id string, input domain.UserUpdate) error
	Delete(ctx context.Context, id string) error
}

type UserService struct {
	repository repository.UserRepositoryInterface
}

func NewUserService(repository repository.UserRepositoryInterface) *UserService {
	return &UserService{
		repository: repository,
	}
}

func (s *UserService) Create(ctx context.Context, input domain.UserRegister) (string, error) {
	id, err := s.repository.Create(ctx, input)

	switch {
	case errors.Is(err, database.ErrInvalidID):
		return "", apperror.New(apperror.BadRequest, "invalid id")
	case errors.Is(err, database.ErrDuplicate):
		return "", apperror.New(apperror.Conflict, "register already exists")
	}

	return id, err
}

func (s *UserService) GetOne(ctx context.Context, id string) (domain.User, error) {
	userData, err := s.repository.GetOne(ctx, id)

	if errors.Is(err, database.ErrNotFound) {
		return domain.User{}, apperror.New(apperror.NotFound, "register not found")
	}

	return userData, err
}

func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.repository.List(ctx)
}

func (s *UserService) Update(ctx context.Context, id string, input domain.UserUpdate) error {
	err := s.repository.Update(ctx, id, input)

	if errors.Is(err, database.ErrNotFound) {
		return apperror.New(apperror.NotFound, "register not found")
	}

	return err
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	err := s.repository.Delete(ctx, id)

	if errors.Is(err, database.ErrNotFound) {
		return apperror.New(apperror.NotFound, "register not found")
	}

	return err
}
