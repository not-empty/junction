package event

import (
	"context"
	"fmt"
	"maps"
	"slices"
)

type Message struct {
	Topic   string
	Key     string
	Headers map[string]string
	Payload []byte
}

type Handler func(ctx context.Context, msg Message) error

type Registry map[string]Handler

func (r Registry) Register(topic string, handler Handler) {
	if _, exists := r[topic]; exists {
		panic(fmt.Sprintf("topic %q already registered", topic))
	}

	r[topic] = handler
}

func (r Registry) Topics() []string {
	return slices.Sorted(maps.Keys(r))
}

type Subscriber interface {
	Run(ctx context.Context, handler Handler) error
	Close() error
}
