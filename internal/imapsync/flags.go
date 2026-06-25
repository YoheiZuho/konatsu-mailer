package imapsync

import (
	"context"
	"fmt"

	"github.com/emersion/go-imap/v2"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

// SetSeen propagates a read/unread change to the IMAP server by adding or
// removing the \Seen flag on a message UID. Used to keep app and server state
// in sync. Opens a short-lived connection (best-effort).
func SetSeen(ctx context.Context, a domain.Account, password, folder string, uid int64, seen bool) error {
	c, err := dial(a)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer c.Close()
	if err := c.Login(a.AuthUser, password).Wait(); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	if _, err := c.Select(folder, nil).Wait(); err != nil {
		return fmt.Errorf("select: %w", err)
	}

	op := imap.StoreFlagsDel
	if seen {
		op = imap.StoreFlagsAdd
	}
	cmd := c.Store(imap.UIDSetNum(imap.UID(uid)), &imap.StoreFlags{
		Op:     op,
		Silent: true,
		Flags:  []imap.Flag{imap.FlagSeen},
	}, nil)
	if err := cmd.Close(); err != nil {
		return fmt.Errorf("store: %w", err)
	}
	return nil
}
