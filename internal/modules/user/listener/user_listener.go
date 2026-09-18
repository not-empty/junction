package listener

import (
	"context"

	"github.com/not-empty/bridge/internal/modules/user/domain"
	"github.com/not-empty/bridge/internal/modules/user/service"
	"github.com/not-empty/bridge/platform/apperror"
	"github.com/not-empty/bridge/platform/event"
)

type UserListener struct {
	service service.UserServiceInterface
}

func NewUserListener(service service.UserServiceInterface) *UserListener {
	return &UserListener{
		service: service,
	}
}

func (l *UserListener) Registered(ctx context.Context, msg event.Message) error {
	var userRegister domain.UserRegister

	err := event.ValidateData(msg, &userRegister)
	if err != nil {
		return err
	}

	if userRegister.ID == "" {
		return apperror.New(apperror.BadRequest, "id is required")
	}

	_, err = l.service.Create(ctx, userRegister)

	return err
}
