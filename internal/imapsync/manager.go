package imapsync

import "context"

// SyncManager manages IMAP IDLE connections per account.
type SyncManager struct{}

func NewManager() *SyncManager { return &SyncManager{} }

func (m *SyncManager) Start(ctx context.Context) error { return nil }
func (m *SyncManager) Stop()                           {}
