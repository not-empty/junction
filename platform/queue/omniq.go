package queue

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/not-empty/junction/platform/apperror"
	"github.com/not-empty/omniq-go/src/omniq"
)

func OmniqHandler(handler Handler, timeout time.Duration) omniq.ConsumeHandler {
	return func(jobCtx omniq.JobCtx) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err := handler(ctx, Job{
			ID:          jobCtx.JobID,
			Attempt:     jobCtx.Attempt,
			MaxAttempts: jobCtx.MaxAttempts,
			Payload:     []byte(jobCtx.PayloadRaw),
		})

		if err == nil {
			return
		}

		var appErr *apperror.Error
		if errors.As(err, &appErr) {
			slog.Warn("job discarded", "queue", jobCtx.Queue, "job_id", jobCtx.JobID, "error", err)
			return
		}

		panic(err)
	}
}
