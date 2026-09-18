package queue

import (
	"context"
	"fmt"
)

type Job struct {
	ID          string
	Attempt     int
	MaxAttempts int
	Payload     []byte
}

type Handler func(ctx context.Context, job Job) error

type Registry map[string]Handler

func (r Registry) Register(queue string, handler Handler) {
	if _, exists := r[queue]; exists {
		panic(fmt.Sprintf("queue %q already registered", queue))
	}

	r[queue] = handler
}
