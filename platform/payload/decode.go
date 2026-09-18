package payload

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/not-empty/bridge/platform/apperror"
	"github.com/not-empty/bridge/platform/validation"
)

func Decode[T any](raw []byte, data *T) error {
	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	if len(raw) == 0 {
		return apperror.New(apperror.BadRequest, "Payload is empty")
	}

	err := json.Unmarshal(raw, data)
	if err != nil {
		return decodeError(err)
	}

	return validation.Struct(data)
}

func decodeError(err error) error {
	var typeErr *json.UnmarshalTypeError

	if errors.As(err, &typeErr) && typeErr.Field != "" {
		return apperror.New(apperror.BadRequest, fmt.Sprintf("Invalid type for field %q", typeErr.Field))
	}

	return apperror.New(apperror.BadRequest, "Malformed JSON payload")
}
