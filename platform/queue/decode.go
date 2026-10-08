package queue

import "github.com/not-empty/junction/platform/payload"

func ValidateData[T any](job Job, data *T) error {
	return payload.Decode(job.Payload, data)
}
