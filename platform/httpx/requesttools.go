package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/not-empty/bridge/platform/apperror"
	"github.com/not-empty/bridge/platform/validation"
)

const unknownFieldPrefix = "json: unknown field "

var internalErrorBody = []byte(`{"message":"Internal Error"}`)

var statusByKind = map[apperror.Kind]int{
	apperror.BadRequest:      http.StatusBadRequest,
	apperror.Validation:      http.StatusUnprocessableEntity,
	apperror.Unauthorized:    http.StatusUnauthorized,
	apperror.Forbidden:       http.StatusForbidden,
	apperror.NotFound:        http.StatusNotFound,
	apperror.Conflict:        http.StatusConflict,
	apperror.PayloadTooLarge: http.StatusRequestEntityTooLarge,
}

type errorBody struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type ResponseData[T any] struct {
	Data       T      `json:"data"`
	PageCursor string `json:"page_cursor,omitempty"`
}

func Response(w http.ResponseWriter, statusCode int, content any, err error) {
	ResponsePage(w, statusCode, content, "", err)
}

func ResponsePage(w http.ResponseWriter, statusCode int, content any, cursor string, err error) {
	if err != nil {
		ResponseError(w, err)
		return
	}

	if content == nil || statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return
	}

	responseEncoded, err := json.Marshal(ResponseData[any]{Data: content, PageCursor: cursor})
	if err != nil {
		slog.Error("failed to encode response", "error", err)
		ResponseInternalError(w)
		return
	}

	execResponse(w, statusCode, responseEncoded)
}

func ResponseError(w http.ResponseWriter, err error) {
	var appErr *apperror.Error

	if !errors.As(err, &appErr) {
		slog.Error("unexpected error", "error", err)
		ResponseInternalError(w)
		return
	}

	statusCode, ok := statusByKind[appErr.Kind]
	if !ok {
		slog.Error("unmapped error kind", "kind", appErr.Kind, "error", err)
		ResponseInternalError(w)
		return
	}

	responseEncoded, err := json.Marshal(errorBody{Message: appErr.Message, Fields: appErr.Fields})
	if err != nil {
		slog.Error("failed to encode error response", "error", err)
		ResponseInternalError(w)
		return
	}

	execResponse(w, statusCode, responseEncoded)
}

func ResponseInternalError(w http.ResponseWriter) {
	execResponse(w, http.StatusInternalServerError, internalErrorBody)
}

func execResponse(w http.ResponseWriter, statusCode int, content []byte) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)
	w.Write(content)
}

func ValidateData[T any](requestBody io.ReadCloser, data *T) error {
	defer requestBody.Close()

	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	decoder := json.NewDecoder(requestBody)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(data)
	if err != nil {
		return decodeError(err)
	}

	var extra json.RawMessage
	err = decoder.Decode(&extra)
	if err == nil {
		return apperror.New(apperror.BadRequest, "Request body must contain a single JSON object")
	}
	if !errors.Is(err, io.EOF) {
		return decodeError(err)
	}

	return validation.Struct(data)
}

func decodeError(err error) error {
	var typeErr *json.UnmarshalTypeError
	var maxBytesErr *http.MaxBytesError

	switch {
	case errors.Is(err, io.EOF):
		return apperror.New(apperror.BadRequest, "Request body is empty")
	case errors.As(err, &maxBytesErr):
		return apperror.New(apperror.PayloadTooLarge, fmt.Sprintf("Request body must not exceed %d bytes", maxBytesErr.Limit))
	case errors.As(err, &typeErr) && typeErr.Field != "":
		return apperror.New(apperror.BadRequest, fmt.Sprintf("Invalid type for field %q", typeErr.Field))
	case strings.HasPrefix(err.Error(), unknownFieldPrefix):
		return apperror.New(apperror.BadRequest, "Unknown field "+strings.TrimPrefix(err.Error(), unknownFieldPrefix))
	default:
		return apperror.New(apperror.BadRequest, "Malformed JSON")
	}
}
