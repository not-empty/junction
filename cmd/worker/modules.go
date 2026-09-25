package main

import (
	"github.com/not-empty/bridge/platform/bootstrap"
)

// `make worker name=<domain>` appends to this list.
var modules = []bootstrap.QueueModule{}
