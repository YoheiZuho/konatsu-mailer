package imapsync

import (
	"context"
	"fmt"
	"sync"

	"github.com/emersion/go-imap/v2/imapclient"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

// Pool keeps one reusable IMAP connection per account for on-demand operations
// (body fetch, flag changes, moves), so API requests don't dial a fresh
// connection every time. It is independent of the sync workers' own IDLE
// connections; a per-account mutex serializes use (IMAP allows one command at a
// time per connection).
type Pool struct {
	mu    sync.Mutex
	conns map[domain.UUID]*pooledConn
}

type pooledConn struct {
	mu     sync.Mutex
	c      *imapclient.Client
	folder string
}

func NewPool() *Pool {
	return &Pool{conns: make(map[domain.UUID]*pooledConn)}
}

func (p *Pool) get(id domain.UUID) *pooledConn {
	p.mu.Lock()
	defer p.mu.Unlock()
	pc := p.conns[id]
	if pc == nil {
		pc = &pooledConn{}
		p.conns[id] = pc
	}
	return pc
}

func (pc *pooledConn) invalidate() {
	if pc.c != nil {
		_ = pc.c.Close()
		pc.c = nil
		pc.folder = ""
	}
}

// do runs fn with the account's connection, ensuring it is connected, logged in
// and (optionally) selected on folder. A failed operation invalidates the
// connection so the next call reconnects.
func (p *Pool) do(a domain.Account, password, folder string, fn func(*imapclient.Client) error) error {
	pc := p.get(a.ID)
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.c == nil {
		c, err := dial(a)
		if err != nil {
			return fmt.Errorf("dial: %w", err)
		}
		if err := c.Login(a.AuthUser, password).Wait(); err != nil {
			_ = c.Close()
			return fmt.Errorf("login: %w", err)
		}
		pc.c = c
		pc.folder = ""
	}
	if folder != "" && pc.folder != folder {
		if _, err := pc.c.Select(folder, nil).Wait(); err != nil {
			pc.invalidate()
			return fmt.Errorf("select: %w", err)
		}
		pc.folder = folder
	}
	if err := fn(pc.c); err != nil {
		pc.invalidate()
		return err
	}
	return nil
}

// Close shuts down all pooled connections.
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, pc := range p.conns {
		pc.mu.Lock()
		pc.invalidate()
		pc.mu.Unlock()
	}
}

// FetchBodies retrieves+parses bodies for UIDs in a folder (reusing the conn).
func (p *Pool) FetchBodies(_ context.Context, a domain.Account, password, folder string, uids []int64) (map[int64]ParsedBody, error) {
	if len(uids) == 0 {
		return map[int64]ParsedBody{}, nil
	}
	var out map[int64]ParsedBody
	err := p.do(a, password, folder, func(c *imapclient.Client) error {
		var e error
		out, e = fetchBodiesOnConn(c, uids)
		return e
	})
	return out, err
}

// SetSeen propagates a read/unread change to IMAP (reusing the conn).
func (p *Pool) SetSeen(_ context.Context, a domain.Account, password, folder string, uid int64, seen bool) error {
	return p.do(a, password, folder, func(c *imapclient.Client) error {
		return setSeenOnConn(c, uid, seen)
	})
}

// MoveMessage moves a UID to destFolder (reusing the conn).
func (p *Pool) MoveMessage(_ context.Context, a domain.Account, password, srcFolder string, uid int64, destFolder string) error {
	return p.do(a, password, srcFolder, func(c *imapclient.Client) error {
		return moveOnConn(c, uid, destFolder)
	})
}
