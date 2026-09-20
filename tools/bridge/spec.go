package main

type fileSpec struct {
	Template string
	Path     string
}

type bridgeSpec struct {
	Layer    string
	Files    []fileSpec
	Target   string
	Fn       string
	Register string
	Wiring   string
}

var coreFiles = []fileSpec{
	{Template: "domain.go.tmpl", Path: "domain/{{.Snake}}.go"},
	{Template: "repository.go.tmpl", Path: "repository/{{.Snake}}_repository.go"},
	{Template: "service.go.tmpl", Path: "service/{{.Snake}}_service.go"},
}

var bridges = map[string]bridgeSpec{
	"api": {
		Layer: "controller",
		Files: []fileSpec{
			{Template: "controller.go.tmpl", Path: "controller/{{.Snake}}_controller.go"},
			{Template: "routes.go.tmpl", Path: "controller/{{.Snake}}_routes.go"},
		},
		Target:   "cmd/api/router.go",
		Fn:       "newRouter",
		Register: "RegisterRoutes",
		Wiring:   "wiring_api.tmpl",
	},
	"worker": {
		Layer: "consumer",
		Files: []fileSpec{
			{Template: "consumer.go.tmpl", Path: "consumer/{{.Snake}}_consumer.go"},
			{Template: "queues.go.tmpl", Path: "consumer/{{.Snake}}_queues.go"},
		},
		Target:   "cmd/worker/queues.go",
		Fn:       "newRegistry",
		Register: "RegisterQueues",
		Wiring:   "wiring_worker.tmpl",
	},
	"event": {
		Layer: "listener",
		Files: []fileSpec{
			{Template: "listener.go.tmpl", Path: "listener/{{.Snake}}_listener.go"},
			{Template: "events.go.tmpl", Path: "listener/{{.Snake}}_events.go"},
		},
		Target:   "cmd/event/events.go",
		Fn:       "newRegistry",
		Register: "RegisterEvents",
		Wiring:   "wiring_event.tmpl",
	},
}

var bridgeOrder = []string{"api", "worker", "event"}
