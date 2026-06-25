package imapsync

import (
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

type mailbox struct {
	Name string
	Role string
}

// rolesToSync are the special-use roles synced automatically (besides INBOX).
var rolesToSync = map[string]bool{
	"sent": true, "junk": true, "trash": true, "drafts": true, "archive": true,
}

// listMailboxes returns the selectable mailboxes on the server with their
// special-use role resolved (from attributes, falling back to name heuristics).
func listMailboxes(c *imapclient.Client) []mailbox {
	data, err := c.List("", "*", &imap.ListOptions{ReturnSpecialUse: true}).Collect()
	if err != nil {
		return nil
	}
	out := make([]mailbox, 0, len(data))
	for _, d := range data {
		if hasAttr(d.Attrs, imap.MailboxAttrNonExistent) || hasAttr(d.Attrs, imap.MailboxAttrNoSelect) {
			continue
		}
		out = append(out, mailbox{Name: d.Mailbox, Role: roleOf(d.Mailbox, d.Attrs)})
	}
	return out
}

// foldersToSync picks INBOX plus the special-use folders to sync each poll.
func foldersToSync(mailboxes []mailbox) []string {
	set := []string{defaultFolder}
	seen := map[string]bool{strings.ToUpper(defaultFolder): true}
	for _, mb := range mailboxes {
		if rolesToSync[mb.Role] && !seen[strings.ToUpper(mb.Name)] {
			set = append(set, mb.Name)
			seen[strings.ToUpper(mb.Name)] = true
		}
	}
	return set
}

func hasAttr(attrs []imap.MailboxAttr, want imap.MailboxAttr) bool {
	for _, a := range attrs {
		if a == want {
			return true
		}
	}
	return false
}

// roleOf maps a mailbox to a normalized role using special-use attributes first,
// then common name conventions (covers servers without SPECIAL-USE).
func roleOf(name string, attrs []imap.MailboxAttr) string {
	switch {
	case hasAttr(attrs, imap.MailboxAttrSent):
		return "sent"
	case hasAttr(attrs, imap.MailboxAttrJunk):
		return "junk"
	case hasAttr(attrs, imap.MailboxAttrTrash):
		return "trash"
	case hasAttr(attrs, imap.MailboxAttrDrafts):
		return "drafts"
	case hasAttr(attrs, imap.MailboxAttrArchive):
		return "archive"
	}
	upper := strings.ToUpper(name)
	if upper == "INBOX" {
		return "inbox"
	}
	lower := strings.ToLower(name)
	switch {
	case containsAny(lower, "junk", "spam", "迷惑"):
		return "junk"
	case containsAny(lower, "sent", "送信"):
		return "sent"
	case containsAny(lower, "trash", "deleted", "ゴミ", "ごみ"):
		return "trash"
	case containsAny(lower, "draft", "下書き"):
		return "drafts"
	case containsAny(lower, "archive", "アーカイブ"):
		return "archive"
	}
	return ""
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
