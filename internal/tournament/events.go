package tournament

import (
	"sync"
	"time"
)

// Event represents a broadcast tournament state notification.
type Event struct {
	Type      string      `json:"type"` // "match_start", "match_finish", "round_complete", "elo_update"
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// EventBus provides pub-sub event distribution for live server-sent events.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[chan Event]struct{}
}

// NewEventBus initializes a central event hub.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[chan Event]struct{}),
	}
}

// Subscribe opens a new event listener channel.
func (b *EventBus) Subscribe() chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, 64)
	b.subscribers[ch] = struct{}{}
	return ch
}

// Unsubscribe closes and unregisters an event channel.
func (b *EventBus) Unsubscribe(ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}

// Publish broadcasts an event to all active subscribers.
func (b *EventBus) Publish(eventType string, payload interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	event := Event{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Non-blocking drop for slow subscribers to preserve engine throughput
		}
	}
}
