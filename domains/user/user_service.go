package user

import (
	"context"
	"time"
)

type ServiceInterface interface {
	Create(ctx context.Context, userRegister UserRegister) (User, error)
}

type Service struct {
	repository RepositoryInterface
}

func NewService(repository RepositoryInterface) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, userRegister UserRegister) (User, error) {
	now := time.Now().UTC().Truncate(time.Microsecond)

	return s.repository.Create(ctx, User{
		Name:      userRegister.Name,
		CreatedAt: now,
		UpdatedAt: now,
	})
}
