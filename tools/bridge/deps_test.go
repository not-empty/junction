package main

// Keeps in go.mod the dependencies that only the generated code imports, so
// go mod tidy does not drop them.
import (
	_ "github.com/DATA-DOG/go-sqlmock"
)
