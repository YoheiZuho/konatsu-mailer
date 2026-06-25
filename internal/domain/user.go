package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type UUID string

// Recipient represents an email address with display name.
type Recipient struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
	Type string `json:"type"` // to, cc, bcc
}

type Recipients []Recipient

func (r Recipients) Value() (driver.Value, error) {
	return json.Marshal(r)
}
func (r *Recipients) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T", value)
	}
	return json.Unmarshal(b, r)
}

// SyncState tracks per-folder IMAP sync progress.
type SyncState map[string]FolderSyncState

type FolderSyncState struct {
	UIDValidity uint32 `json:"uidvalidity"`
	LastUID     uint32 `json:"last_uid"`
}

func (s SyncState) Value() (driver.Value, error) { return json.Marshal(s) }
func (s *SyncState) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T", value)
	}
	return json.Unmarshal(b, s)
}

// User represents a single application user.
type User struct {
	ID          UUID      `db:"id" json:"id"`
	Email       string    `db:"email" json:"email"`
	DisplayName string    `db:"display_name" json:"display_name"`
	Theme       string    `db:"theme" json:"theme"`
	BrandColor  string    `db:"brand_color" json:"brand_color"`
	Prefs       UserPrefs `db:"prefs" json:"prefs"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// UserPrefs holds UI-level preferences persisted per user.
type UserPrefs map[string]interface{}

func (p UserPrefs) Value() (driver.Value, error) { return json.Marshal(p) }
func (p *UserPrefs) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T", value)
	}
	return json.Unmarshal(b, p)
}
