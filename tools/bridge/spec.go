package main

type fileSpec struct {
	Template string
	Path     string
}

type bridgeSpec struct {
	Layer  string
	Files  []fileSpec
	Target string
}

var coreFiles = []fileSpec{
	{Template: "domain.go.tmpl", Path: "domain/{{.Snake}}.go"},
	{Template: "repository.go.tmpl", Path: "repository/{{.Snake}}_repository.go"},
	{Template: "repository_test.go.tmpl", Path: "repository/{{.Snake}}_repository_test.go"},
	{Template: "service.go.tmpl", Path: "service/{{.Snake}}_service.go"},
	{Template: "service_test.go.tmpl", Path: "service/{{.Snake}}_service_test.go"},
}

var bridges = map[string]bridgeSpec{
	"api": {
		Layer: "controller",
		Files: []fileSpec{
			{Template: "controller.go.tmpl", Path: "controller/{{.Snake}}_controller.go"},
			{Template: "controller_test.go.tmpl", Path: "controller/{{.Snake}}_controller_test.go"},
			{Template: "routes.go.tmpl", Path: "controller/{{.Snake}}_routes.go"},
		},
		Target: "cmd/api/modules.go",
	},
	"worker": {
		Layer: "consumer",
		Files: []fileSpec{
			{Template: "consumer.go.tmpl", Path: "consumer/{{.Snake}}_consumer.go"},
			{Template: "consumer_test.go.tmpl", Path: "consumer/{{.Snake}}_consumer_test.go"},
			{Template: "queues.go.tmpl", Path: "consumer/{{.Snake}}_queues.go"},
		},
		Target: "cmd/worker/modules.go",
	},
	"event": {
		Layer: "listener",
		Files: []fileSpec{
			{Template: "listener.go.tmpl", Path: "listener/{{.Snake}}_listener.go"},
			{Template: "listener_test.go.tmpl", Path: "listener/{{.Snake}}_listener_test.go"},
			{Template: "events.go.tmpl", Path: "listener/{{.Snake}}_events.go"},
		},
		Target: "cmd/event/modules.go",
	},
}

var bridgeOrder = []string{"api", "worker", "event"}
