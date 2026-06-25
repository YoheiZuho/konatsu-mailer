package ws

import "context"

// Hub manages WebSocket connections per user.
type Hub struct{}

func NewHub() *Hub { return &Hub{} }

func (h *Hub) Run(ctx context.Context) {}
func (h *Hub) Broadcast(userID string, msg any) {}
