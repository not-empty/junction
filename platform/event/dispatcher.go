package event

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/not-empty/bridge/platform/apperror"
)

func Dispatcher(registry Registry, timeout time.Duration) Handler {
	return func(ctx context.Context, msg Message) error {
		handler, ok := registry[msg.Topic]
		if !ok {
			slog.Warn("event without handler", "topic", msg.Topic)
			return nil
		}

		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
		defer cancel()

		err := handler(ctx, msg)

		var appErr *apperror.Error
		if errors.As(err, &appErr) {
			slog.Warn("event discarded", "topic", msg.Topic, "key", msg.Key, "error", err)
			return nil
		}

		return err
	}
}
