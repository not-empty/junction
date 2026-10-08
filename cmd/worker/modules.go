package main

import (
	"github.com/not-empty/junction/platform/bootstrap"
)

// `make worker name=<domain>` appends to this list.
var modules = []bootstrap.QueueModule{}
