package main

import (
	healthcontroller "github.com/not-empty/junction/internal/modules/health/controller"
	"github.com/not-empty/junction/platform/bootstrap"
)

// `make api name=<domain>` appends to this list.
var modules = []bootstrap.HTTPModule{
	healthcontroller.Module,
}
