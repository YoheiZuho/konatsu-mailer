package imapsync

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// setSeenOnConn adds/removes the \Seen flag on a UID using a folder-selected
// connection.
func setSeenOnConn(c *imapclient.Client, uid int64, seen bool) error {
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

// moveOnConn moves a UID to destFolder using a folder-selected connection.
func moveOnConn(c *imapclient.Client, uid int64, destFolder string) error {
	if _, err := c.Move(imap.UIDSetNum(imap.UID(uid)), destFolder).Wait(); err != nil {
		return fmt.Errorf("move: %w", err)
	}
	return nil
}
