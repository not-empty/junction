package main

import (
	healthcontroller "github.com/not-empty/bridge/internal/modules/health/controller"
	"github.com/not-empty/bridge/platform/bootstrap"
)

// `make api name=<domain>` appends to this list.
var modules = []bootstrap.HTTPModule{
	healthcontroller.Module,
}
