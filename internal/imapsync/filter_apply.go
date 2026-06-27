package imapsync

import (
	"context"
	"log/slog"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

// socialDomains are senders treated as the "social" inbox category.
var socialDomains = []string{
	"facebook", "twitter", "x.com", "linkedin", "instagram", "line.me",
	"slack", "discord", "github", "gitlab", "notion", "youtube", "tiktok",
}

// categorize assigns an inbox category from classification headers + sender,
// approximating Gmail-style tabs (primary / promotions / social / newsletters).
func categorize(headerBlob, sender string) string {
	h := strings.ToLower(headerBlob)
	s := strings.ToLower(sender)

	for _, d := range socialDomains {
		if strings.Contains(s, d) {
			return "social"
		}
	}
	if strings.Contains(h, "list-id:") {
		return "newsletters"
	}
	if strings.Contains(h, "list-unsubscribe:") ||
		strings.Contains(h, "precedence: bulk") ||
		strings.Contains(h, "precedence:bulk") ||
		strings.Contains(h, "auto-submitted:") {
		return "promotions"
	}
	return "primary"
}

// applyFilters evaluates a user's filters against a newly-synced email and
// performs the matching actions (move/label/read/star/category). The IMAP
// client must be selected on the email's folder.
func (m *SyncManager) applyFilters(ctx context.Context, c *imapclient.Client, a *domain.Account, id domain.UUID, email *domain.Email, filters []domain.Filter) {
	for _, f := range filters {
		if !filterMatches(f, email) {
			continue
		}
		moved := false
		for _, action := range f.Actions {
			switch action.Type {
			case "move_folder":
				if action.Value == "" {
					continue
				}
				if _, err := c.Move(imap.UIDSetNum(imap.UID(email.IMAPUID)), action.Value).Wait(); err != nil {
					slog.Warn("filter: move failed", slog.String("to", action.Value), slog.Any("error", err))
					continue
				}
				// The message now lives in the target folder; drop the source row
				// (it will be re-synced there).
				_ = m.db.DeleteEmail(ctx, id)
				moved = true
			case "add_label":
				if labelID, err := m.db.GetOrCreateLabel(ctx, a.ID, action.Value); err == nil {
					_ = m.db.LinkEmailLabel(ctx, id, labelID, "filter")
				}
			case "mark_read":
				_ = m.db.SetReadByID(ctx, id, true)
				_ = c.Store(imap.UIDSetNum(imap.UID(email.IMAPUID)),
					&imap.StoreFlags{Op: imap.StoreFlagsAdd, Silent: true, Flags: []imap.Flag{imap.FlagSeen}}, nil).Close()
			case "star":
				_ = m.db.SetStarredByID(ctx, id, true)
			case "set_category":
				_ = m.db.SetCategoryByID(ctx, id, action.Value)
			}
			if moved {
				break
			}
		}
		if moved {
			return // message relocated; stop processing further filters
		}
	}
}

// filterMatches evaluates a filter's conditions (all/any) against an email.
func filterMatches(f domain.Filter, email *domain.Email) bool {
	if len(f.Conditions) == 0 {
		return false
	}
	any := f.MatchType == "any"
	for _, cond := range f.Conditions {
		ok := conditionMatches(cond, email)
		if any && ok {
			return true
		}
		if !any && !ok {
			return false
		}
	}
	return !any
}

func conditionMatches(cond domain.FilterCondition, email *domain.Email) bool {
	field := strings.ToLower(fieldValue(cond.Field, email))
	want := strings.ToLower(cond.Value)
	switch cond.Op {
	case "contains":
		return strings.Contains(field, want)
	case "not_contains":
		return !strings.Contains(field, want)
	case "is":
		return field == want
	case "is_not":
		return field != want
	case "starts_with":
		return strings.HasPrefix(field, want)
	case "ends_with":
		return strings.HasSuffix(field, want)
	default:
		return false
	}
}

func fieldValue(field string, email *domain.Email) string {
	switch field {
	case "subject":
		return email.Subject
	case "from":
		return email.SenderName + " " + email.SenderAddr
	case "body":
		return email.BodyPreview
	case "to":
		return recipientsOfType(email.Recipients, "to")
	case "cc":
		return recipientsOfType(email.Recipients, "cc")
	case "recipient":
		return recipientsOfType(email.Recipients, "")
	default:
		return ""
	}
}

func recipientsOfType(rs domain.Recipients, typ string) string {
	var b strings.Builder
	for _, r := range rs {
		if typ == "" || r.Type == typ {
			b.WriteString(r.Name)
			b.WriteByte(' ')
			b.WriteString(r.Addr)
			b.WriteByte(' ')
		}
	}
	return b.String()
}
