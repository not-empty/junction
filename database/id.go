package database

import (
	"fmt"

	"github.com/oklog/ulid/v2"
)

func NewID() string {
	return ulid.Make().String()
}

func resolveID(id string) (string, error) {
	if id == "" {
		return NewID(), nil
	}

	parsed, err := ulid.ParseStrict(id)
	if err != nil {
		return "", fmt.Errorf("%w: %q", ErrInvalidID, id)
	}

	return parsed.String(), nil
}
