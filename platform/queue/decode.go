package queue

import "github.com/not-empty/bridge/platform/payload"

func ValidateData[T any](job Job, data *T) error {
	return payload.Decode(job.Payload, data)
}
