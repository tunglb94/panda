// Package kafka provides Kafka producer and consumer primitives for FAIRRIDE services.
package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// ProducerConfig configures a Kafka producer for a single topic.
type ProducerConfig struct {
	Brokers      []string
	Topic        string
	Async        bool
	BatchTimeout time.Duration

	// RequireAllAcks enables RequireAll acknowledgement mode.
	// Set true for financial topics (wallet, payment) to prevent message loss
	// on broker failure. Default (false) uses RequireOne — faster but less safe.
	RequireAllAcks bool
}

// Producer wraps a kafka.Writer with a simplified publish interface.
type Producer struct {
	writer *kafka.Writer
	topic  string
}

// NewProducer creates a new Kafka producer. Close() must be called when done.
func NewProducer(cfg ProducerConfig) *Producer {
	batchTimeout := cfg.BatchTimeout
	if batchTimeout == 0 {
		batchTimeout = 10 * time.Millisecond
	}

	acks := kafka.RequireOne
	if cfg.RequireAllAcks {
		acks = kafka.RequireAll
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: acks,
		Async:        cfg.Async,
		Compression:  kafka.Snappy,
		WriteTimeout: 10 * time.Second,
		BatchTimeout: batchTimeout,
		MaxAttempts:  3,
	}

	return &Producer{writer: w, topic: cfg.Topic}
}

// Publish sends a keyed message to the configured Kafka topic.
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("kafka publish to %s: %w", p.topic, err)
	}
	return nil
}

// PublishBatch sends multiple messages in a single call to the same topic.
func (p *Producer) PublishBatch(ctx context.Context, messages []Message) error {
	kmsgs := make([]kafka.Message, len(messages))
	for i, m := range messages {
		kmsgs[i] = kafka.Message{
			Key:   m.Key,
			Value: m.Value,
			Time:  time.Now().UTC(),
		}
	}
	if err := p.writer.WriteMessages(ctx, kmsgs...); err != nil {
		return fmt.Errorf("kafka batch publish to %s: %w", p.topic, err)
	}
	return nil
}

// Close flushes pending messages and closes the underlying writer.
func (p *Producer) Close() error {
	return p.writer.Close()
}
