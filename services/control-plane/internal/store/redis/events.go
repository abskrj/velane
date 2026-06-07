package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// eventStreamKey returns the Redis Stream key carrying live invocation events.
func eventStreamKey(invocationID string) string {
	return "velane:inv:" + invocationID + ":events"
}

// Defaults bounding the per-invocation event stream.
const (
	eventStreamMaxLen = 10000           // approximate cap on retained events
	eventStreamTTL    = 5 * time.Minute // retention past completion
)

// StreamEvent is one event on a per-invocation Redis Stream. ID is the Redis
// stream entry ID (e.g. "1700000000000-0") and doubles as the SSE event id for
// resume. Payload is the raw JSON event body.
type StreamEvent struct {
	ID      string
	Payload json.RawMessage
}

// PublishEvent appends a single JSON event to the invocation's event stream and
// refreshes the key TTL. payload must be valid JSON.
func (c *Client) PublishEvent(ctx context.Context, invocationID string, payload []byte) error {
	key := eventStreamKey(invocationID)
	if err := c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		MaxLen: eventStreamMaxLen,
		Approx: true,
		Values: map[string]any{"d": payload},
	}).Err(); err != nil {
		return fmt.Errorf("publish event xadd: %w", err)
	}
	// Best-effort TTL refresh; ignore errors (the stream still functions).
	_ = c.rdb.Expire(ctx, key, eventStreamTTL).Err()
	return nil
}

// ReadEvents blocks up to blockFor for new events after lastID. Pass "0" to read
// from the beginning (replay) or "$" to read only new events. Returns the events
// and the ID to pass on the next call. When no events arrive before the timeout
// it returns (nil, lastID, nil).
func (c *Client) ReadEvents(ctx context.Context, invocationID, lastID string, blockFor time.Duration) ([]StreamEvent, string, error) {
	key := eventStreamKey(invocationID)
	if lastID == "" {
		lastID = "0"
	}

	res, err := c.rdb.XRead(ctx, &redis.XReadArgs{
		Streams: []string{key, lastID},
		Count:   256,
		Block:   blockFor,
	}).Result()
	if err != nil {
		// No new data within the block window.
		if errors.Is(err, redis.Nil) {
			return nil, lastID, nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, lastID, ctx.Err()
		}
		return nil, lastID, fmt.Errorf("read events xread: %w", err)
	}

	var events []StreamEvent
	next := lastID
	for _, stream := range res {
		for _, msg := range stream.Messages {
			next = msg.ID
			raw, _ := msg.Values["d"].(string)
			events = append(events, StreamEvent{ID: msg.ID, Payload: json.RawMessage(raw)})
		}
	}
	return events, next, nil
}

// DeleteEventStream removes the invocation event stream. Best-effort cleanup.
func (c *Client) DeleteEventStream(ctx context.Context, invocationID string) error {
	return c.rdb.Del(ctx, eventStreamKey(invocationID)).Err()
}
