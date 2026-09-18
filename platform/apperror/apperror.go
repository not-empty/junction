package apperror

type Kind int

const (
	_ Kind = iota
	BadRequest
	Validation
	Unauthorized
	Forbidden
	NotFound
	Conflict
	PayloadTooLarge
)

type Error struct {
	Kind    Kind
	Message string
	Fields  map[string]string
}

func New(kind Kind, message string) *Error {
	return &Error{Kind: kind, Message: message}
}

func (e *Error) Error() string {
	return e.Message
}
