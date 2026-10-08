package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/not-empty/junction/platform/event"
)

type Config struct {
	Brokers     []string
	GroupID     string
	MaxAttempts int
	Backoff     time.Duration
	DLQSuffix   string
}

type Subscriber struct {
	cfg    Config
	topics []string
	reader *kafkago.Reader
	dlq    *kafkago.Writer
}

func NewSubscriber(cfg Config, topics []string) *Subscriber {
	return &Subscriber{
		cfg:    cfg,
		topics: topics,
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:     cfg.Brokers,
			GroupID:     cfg.GroupID,
			GroupTopics: topics,
		}),
		dlq: &kafkago.Writer{
			Addr:     kafkago.TCP(cfg.Brokers...),
			Balancer: &kafkago.Hash{},
		},
	}
}

func (s *Subscriber) Run(ctx context.Context, handler event.Handler) error {
	err := s.ensureDLQTopics(ctx)
	if err != nil {
		return err
	}

	for {
		raw, err := s.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			return fmt.Errorf("fetching message: %w", err)
		}

		err = s.process(ctx, handler, toMessage(raw))
		if errors.Is(err, context.Canceled) {
			return nil
		}

		if err != nil {
			return err
		}

		err = s.reader.CommitMessages(context.WithoutCancel(ctx), raw)
		if err != nil {
			return fmt.Errorf("committing offset: %w", err)
		}
	}
}

func (s *Subscriber) process(ctx context.Context, handler event.Handler, msg event.Message) error {
	var err error

	for attempt := 1; attempt <= s.cfg.MaxAttempts; attempt++ {
		err = handler(ctx, msg)
		if err == nil {
			return nil
		}

		slog.Error("event failed", "topic", msg.Topic, "key", msg.Key, "attempt", attempt, "error", err)

		if attempt == s.cfg.MaxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(s.cfg.Backoff * time.Duration(attempt)):
		}
	}

	dlqErr := s.dlq.WriteMessages(context.WithoutCancel(ctx), kafkago.Message{
		Topic:   msg.Topic + s.cfg.DLQSuffix,
		Key:     []byte(msg.Key),
		Value:   msg.Payload,
		Headers: []kafkago.Header{{Key: "error", Value: []byte(err.Error())}},
	})

	if dlqErr != nil {
		return fmt.Errorf("writing to dlq: %w", dlqErr)
	}

	return nil
}

func (s *Subscriber) ensureDLQTopics(ctx context.Context) error {
	configs := make([]kafkago.TopicConfig, 0, len(s.topics))
	for _, topic := range s.topics {
		configs = append(configs, kafkago.TopicConfig{
			Topic:             topic + s.cfg.DLQSuffix,
			NumPartitions:     -1,
			ReplicationFactor: -1,
		})
	}

	client := &kafkago.Client{Addr: kafkago.TCP(s.cfg.Brokers...)}

	res, err := client.CreateTopics(ctx, &kafkago.CreateTopicsRequest{Topics: configs})
	if err != nil {
		return fmt.Errorf("creating dlq topics: %w", err)
	}

	for topic, topicErr := range res.Errors {
		if topicErr != nil && !errors.Is(topicErr, kafkago.TopicAlreadyExists) {
			return fmt.Errorf("creating dlq topic %q: %w", topic, topicErr)
		}
	}

	return nil
}

func (s *Subscriber) Close() error {
	return errors.Join(s.reader.Close(), s.dlq.Close())
}

func toMessage(raw kafkago.Message) event.Message {
	headers := make(map[string]string, len(raw.Headers))
	for _, header := range raw.Headers {
		headers[header.Key] = string(header.Value)
	}

	return event.Message{
		Topic:   raw.Topic,
		Key:     string(raw.Key),
		Headers: headers,
		Payload: raw.Value,
	}
}
