package imapsync

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-message/mail"

	// Registers message.CharsetReader so non-UTF-8 bodies decode correctly.
	_ "github.com/emersion/go-message/charset"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

// AttachmentInfo describes one attachment part.
type AttachmentInfo struct {
	Filename string
	Size     int
}

// ParsedBody is a decoded message body extracted from MIME.
type ParsedBody struct {
	Text        string
	HTML        string
	Attachments []AttachmentInfo
}

// FetchBodies opens a short-lived IMAP connection and retrieves+parses the full
// bodies for the given UIDs in a folder. Used for on-demand detail rendering so
// full bodies are never persisted (design §2.1 / §7.2).
func FetchBodies(ctx context.Context, a domain.Account, password, folder string, uids []int64) (map[int64]ParsedBody, error) {
	if len(uids) == 0 {
		return map[int64]ParsedBody{}, nil
	}

	c, err := dial(a)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	defer c.Close()
	if err := c.Login(a.AuthUser, password).Wait(); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	if _, err := c.Select(folder, nil).Wait(); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	uidList := make([]imap.UID, len(uids))
	for i, u := range uids {
		uidList[i] = imap.UID(u)
	}
	section := &imap.FetchItemBodySection{Peek: true}
	opts := &imap.FetchOptions{UID: true, BodySection: []*imap.FetchItemBodySection{section}}

	msgs, err := c.Fetch(imap.UIDSetNum(uidList...), opts).Collect()
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	out := make(map[int64]ParsedBody, len(msgs))
	for _, msg := range msgs {
		raw := msg.FindBodySection(section)
		if raw == nil {
			continue
		}
		out[int64(msg.UID)] = parseMIME(raw)
	}
	return out, nil
}

// parseMIME extracts plain text, HTML, and attachment metadata from a raw
// RFC 822 message. On parse failure it falls back to treating the bytes as text.
func parseMIME(raw []byte) ParsedBody {
	var body ParsedBody
	mr, err := mail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		body.Text = string(raw)
		return body
	}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		switch h := part.Header.(type) {
		case *mail.InlineHeader:
			contentType, _, _ := h.ContentType()
			data, _ := io.ReadAll(part.Body)
			if contentType == "text/html" {
				body.HTML += string(data)
			} else {
				body.Text += string(data)
			}
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			data, _ := io.ReadAll(part.Body)
			body.Attachments = append(body.Attachments, AttachmentInfo{
				Filename: filename,
				Size:     len(data),
			})
		}
	}
	return body
}
