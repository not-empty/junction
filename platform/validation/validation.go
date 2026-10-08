package validation

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/not-empty/junction/platform/apperror"
)

var validate = newValidator()

func Struct(data any) error {
	return validationError(validate.Struct(data))
}

func validationError(err error) error {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return err
	}

	fields := make(map[string]string, len(validationErrs))
	for _, fieldErr := range validationErrs {
		rule := fieldErr.Tag()
		if fieldErr.Param() != "" {
			rule += "=" + fieldErr.Param()
		}

		_, path, found := strings.Cut(fieldErr.Namespace(), ".")
		if !found {
			path = fieldErr.Field()
		}

		fields[path] = rule
	}

	return &apperror.Error{Kind: apperror.Validation, Message: "Invalid payload", Fields: fields}
}

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}

		return name
	})

	return v
}
