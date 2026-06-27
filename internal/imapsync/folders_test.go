package imapsync

import (
	"testing"

	"github.com/emersion/go-imap/v2"
)

func TestRoleOf_Attributes(t *testing.T) {
	cases := []struct {
		name  string
		attrs []imap.MailboxAttr
		want  string
	}{
		{"Whatever", []imap.MailboxAttr{imap.MailboxAttrSent}, "sent"},
		{"Whatever", []imap.MailboxAttr{imap.MailboxAttrJunk}, "junk"},
		{"Whatever", []imap.MailboxAttr{imap.MailboxAttrTrash}, "trash"},
		{"Whatever", []imap.MailboxAttr{imap.MailboxAttrDrafts}, "drafts"},
		{"Whatever", []imap.MailboxAttr{imap.MailboxAttrArchive}, "archive"},
	}
	for _, c := range cases {
		if got := roleOf(c.name, c.attrs); got != c.want {
			t.Errorf("roleOf(%q, %v) = %q, want %q", c.name, c.attrs, got, c.want)
		}
	}
}

func TestRoleOf_NameHeuristics(t *testing.T) {
	cases := map[string]string{
		"INBOX":         "inbox",
		"Sent Messages": "sent",
		"迷惑メール":         "junk",
		"Spam":          "junk",
		"Trash":         "trash",
		"ゴミ箱":           "trash",
		"Drafts":        "drafts",
		"下書き":           "drafts",
		"Archive":       "archive",
		"Projects":      "",
	}
	for name, want := range cases {
		if got := roleOf(name, nil); got != want {
			t.Errorf("roleOf(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestFoldersToSync(t *testing.T) {
	mbs := []mailbox{
		{Name: "INBOX", Role: "inbox"},
		{Name: "Sent", Role: "sent"},
		{Name: "Junk", Role: "junk"},
		{Name: "Projects", Role: ""}, // custom, not auto-synced
	}
	got := foldersToSync(mbs)
	want := map[string]bool{"INBOX": true, "Sent": true, "Junk": true}
	if len(got) != len(want) {
		t.Fatalf("got %v, want keys %v", got, want)
	}
	for _, f := range got {
		if !want[f] {
			t.Errorf("unexpected folder synced: %q", f)
		}
	}
	if got[0] != "INBOX" {
		t.Errorf("INBOX should be first, got %q", got[0])
	}
}

func TestContainsAny(t *testing.T) {
	if !containsAny("hello spam world", "spam") {
		t.Error("should contain spam")
	}
	if containsAny("clean", "spam", "junk") {
		t.Error("should not match")
	}
}
