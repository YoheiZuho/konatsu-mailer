// Package ws implements a minimal per-user WebSocket fan-out hub used to push
// realtime events (new mail, analysis, sync status) to connected clients.
package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
)

// Subscription is a single client connection's outbound queue.
type Subscription struct {
	C      chan []byte
	hub    *Hub
	userID string
}

// Hub tracks active subscriptions grouped by user id.
type Hub struct {
	mu    sync.RWMutex
	users map[string]map[*Subscription]struct{}
}

func NewHub() *Hub {
	return &Hub{users: make(map[string]map[*Subscription]struct{})}
}

// Run blocks until ctx is cancelled. The hub itself is event-driven; this exists
// so it can be launched as a goroutine alongside other subsystems.
func (h *Hub) Run(ctx context.Context) { <-ctx.Done() }

// Register adds a subscription for a user and returns it.
func (h *Hub) Register(userID string) *Subscription {
	sub := &Subscription{C: make(chan []byte, 32), hub: h, userID: userID}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.users[userID] == nil {
		h.users[userID] = make(map[*Subscription]struct{})
	}
	h.users[userID][sub] = struct{}{}
	return sub
}

// Close removes the subscription from the hub.
func (s *Subscription) Close() {
	h := s.hub
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.users[s.userID]; ok {
		delete(set, s)
		if len(set) == 0 {
			delete(h.users, s.userID)
		}
	}
}

// Broadcast marshals msg and delivers it to every connection for the user.
// Slow consumers are skipped rather than blocking the sender.
func (h *Hub) Broadcast(userID string, msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Warn("ws: marshal event", slog.Any("error", err))
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for sub := range h.users[userID] {
		select {
		case sub.C <- data:
		default: // drop for a backed-up client
		}
	}
}
