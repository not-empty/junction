package consumer

import (
	"context"

	"github.com/not-empty/bridge/internal/modules/user/domain"
	"github.com/not-empty/bridge/internal/modules/user/service"
	"github.com/not-empty/bridge/platform/apperror"
	"github.com/not-empty/bridge/platform/database"
	"github.com/not-empty/bridge/platform/queue"
)

type UserConsumer struct {
	service service.UserServiceInterface
}

func NewUserConsumer(service service.UserServiceInterface) *UserConsumer {
	return &UserConsumer{
		service: service,
	}
}

func (c *UserConsumer) Create(ctx context.Context, job queue.Job) error {
	var userRegister domain.UserRegister

	err := queue.ValidateData(job, &userRegister)
	if err != nil {
		return err
	}

	if userRegister.ID == "" {
		return apperror.New(apperror.BadRequest, "invalid id")
	}

	userRegister.ID, err = database.VerifyULID(userRegister.ID)

	if err != nil {
		return apperror.New(apperror.BadRequest, "invalid id")
	}

	_, err = c.service.Create(ctx, userRegister)

	return err
}
