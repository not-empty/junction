package event

import "github.com/not-empty/bridge/platform/payload"

func ValidateData[T any](msg Message, data *T) error {
	return payload.Decode(msg.Payload, data)
}
