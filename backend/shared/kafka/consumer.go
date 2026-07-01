package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Message is the canonical message type exchanged between producer and consumer.
// It wraps the underlying library message so callers never import kafka-go directly.
type Message struct {
	Key       []byte
	Value     []byte
	Topic     string
	Partition int
	Offset    int64
	Timestamp time.Time

	// raw holds the original kafka message for manual offset commit.
	// Unexported: keeps the kafka-go type internal to this package.
	raw kafka.Message
}

// ConsumerConfig configures a Kafka consumer group reader for a single topic.
type ConsumerConfig struct {
	Brokers     []string
	Topic       string
	GroupID     string
	StartOffset int64 // kafka.LastOffset or kafka.FirstOffset; defaults to LastOffset
}

// Consumer wraps a kafka.Reader with a simplified consume interface.
// Close() must be called when done to release resources.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a new Kafka consumer group reader.
func NewConsumer(cfg ConsumerConfig) *Consumer {
	startOffset := cfg.StartOffset
	if startOffset == 0 {
		startOffset = kafka.LastOffset
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10 MB
		MaxWait:        500 * time.Millisecond,
		CommitInterval: time.Second, // auto-commit when using Read()
		StartOffset:    startOffset,
	})

	return &Consumer{reader: r}
}

// Read blocks until a message is available or ctx is cancelled.
// Offsets are committed automatically via CommitInterval.
// Use FetchMessage + Commit for at-least-once manual control.
func (c *Consumer) Read(ctx context.Context) (*Message, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("kafka read from %s: %w", c.reader.Config().Topic, err)
	}
	return fromRaw(msg), nil
}

// FetchMessage fetches a message WITHOUT auto-committing the offset.
// The caller must call Commit(ctx, msg) after successful processing
// to advance the consumer group offset.
func (c *Consumer) FetchMessage(ctx context.Context) (*Message, error) {
	raw, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("kafka fetch from %s: %w", c.reader.Config().Topic, err)
	}
	return fromRaw(raw), nil
}

// Commit commits the offset of a message that was fetched via FetchMessage.
// Must not be called on messages obtained via Read (those are auto-committed).
func (c *Consumer) Commit(ctx context.Context, msg *Message) error {
	if err := c.reader.CommitMessages(ctx, msg.raw); err != nil {
		return fmt.Errorf("kafka commit offset %d on %s: %w", msg.Offset, msg.Topic, err)
	}
	return nil
}

// Close closes the consumer reader and releases resources.
func (c *Consumer) Close() error {
	return c.reader.Close()
}

func fromRaw(raw kafka.Message) *Message {
	return &Message{
		Key:       raw.Key,
		Value:     raw.Value,
		Topic:     raw.Topic,
		Partition: raw.Partition,
		Offset:    raw.Offset,
		Timestamp: raw.Time,
		raw:       raw,
	}
}
