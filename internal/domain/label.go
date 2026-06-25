package domain

// Label is a tag that can be applied to emails.
type Label struct {
	ID        UUID   `db:"id" json:"id"`
	AccountID UUID   `db:"account_id" json:"account_id"`
	Name      string `db:"name" json:"name"`
	Color     string `db:"color" json:"color"`
	IsSystem  bool   `db:"is_system" json:"is_system"`
}

// EmailLabel links an email to a label with provenance.
type EmailLabel struct {
	EmailID    UUID    `db:"email_id" json:"email_id"`
	LabelID    UUID    `db:"label_id" json:"label_id"`
	Source     string  `db:"source" json:"source"`      // ai | user
	Confidence *float32 `db:"confidence" json:"confidence,omitempty"`
}
