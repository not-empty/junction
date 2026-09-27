// Package bootstrap mounts domain modules on a delivery. Dependencies are
// opened by MountDeps, in deps.go.
package bootstrap

import (
	"net/http"

	"github.com/not-empty/bridge/platform/event"
	"github.com/not-empty/bridge/platform/queue"
)

type HTTPModule func(deps Deps, mux *http.ServeMux)

type QueueModule func(deps Deps, registry queue.Registry)

type EventModule func(deps Deps, registry event.Registry)

func MountHTTP(deps Deps, modules []HTTPModule) *http.ServeMux {
	mux := http.NewServeMux()

	for _, module := range modules {
		module(deps, mux)
	}

	return mux
}

func MountQueues(deps Deps, modules []QueueModule) queue.Registry {
	registry := queue.Registry{}

	for _, module := range modules {
		module(deps, registry)
	}

	return registry
}

func MountEvents(deps Deps, modules []EventModule) event.Registry {
	registry := event.Registry{}

	for _, module := range modules {
		module(deps, registry)
	}

	return registry
}
