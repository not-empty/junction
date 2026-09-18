package consumer

import "github.com/not-empty/bridge/platform/queue"

func RegisterQueues(registry queue.Registry, userConsumer *UserConsumer) {
	registry.Register("user.create", userConsumer.Create)
}
