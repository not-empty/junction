package listener

import "github.com/not-empty/bridge/platform/event"

func RegisterEvents(registry event.Registry, userListener *UserListener) {
	registry.Register("user.registered", userListener.Registered)
}
