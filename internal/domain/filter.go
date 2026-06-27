package domain

// FilterCondition is one condition in a message filter.
// Field: subject | from | to | cc | recipient | body
// Op:    contains | not_contains | is | is_not | starts_with | ends_with
type FilterCondition struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

// FilterAction is one action applied when a filter matches.
// Type: move_folder | add_label | mark_read | star | set_category
// Value: target folder / label name / category, depending on Type.
type FilterAction struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Filter is a user-defined auto-classification rule (Thunderbird-style).
type Filter struct {
	ID         UUID              `json:"id"`
	UserID     UUID              `json:"-"`
	Name       string            `json:"name"`
	Enabled    bool              `json:"enabled"`
	MatchType  string            `json:"match_type"` // all | any
	Conditions []FilterCondition `json:"conditions"`
	Actions    []FilterAction    `json:"actions"`
	Position   int               `json:"position"`
}
