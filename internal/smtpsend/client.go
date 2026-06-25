// Package smtpsend builds and submits outgoing mail over SMTP, supporting both
// implicit TLS (SMTPS, typically port 465) and STARTTLS (typically port 587),
// as well as an unencrypted fallback.
package smtpsend

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"strings"
	"time"
)

// Message is a single outgoing email.
type Message struct {
	FromName string
	FromAddr string
	To       []string
	Cc       []string
	Bcc      []string
	Subject  string
	Text     string
	HTML     string // optional; when set, a multipart/alternative is built

	InReplyTo  string // Message-ID being replied to (optional)
	References string // References header (optional)
}

// Params describes the SMTP server connection plus the message to send.
type Params struct {
	Host        string
	Port        int
	UseStartTLS bool
	AuthUser    string
	Password    string
	Message     Message
}

// Send connects to the SMTP server with the appropriate transport security and
// submits the message.
//
// Transport selection:
//   - port 465        → implicit TLS (SMTPS)
//   - UseStartTLS     → plaintext connection upgraded with STARTTLS
//   - otherwise       → plaintext (only sensible for trusted/local relays)
func Send(ctx context.Context, p Params) error {
	if len(p.Message.To) == 0 {
		return fmt.Errorf("no recipients")
	}
	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
	tlsConfig := &tls.Config{ServerName: p.Host, MinVersion: tls.VersionTLS12}

	var client *smtp.Client
	var err error

	switch {
	case p.Port == 465: // implicit TLS (SMTPS)
		dialer := &tls.Dialer{Config: tlsConfig}
		c, derr := dialer.DialContext(ctx, "tcp", addr)
		if derr != nil {
			return fmt.Errorf("tls dial: %w", derr)
		}
		client, err = smtp.NewClient(c, p.Host)
	default:
		client, err = smtp.Dial(addr)
		if err == nil && p.UseStartTLS {
			if ok, _ := client.Extension("STARTTLS"); ok {
				err = client.StartTLS(tlsConfig)
			} else {
				err = fmt.Errorf("server does not support STARTTLS")
			}
		}
	}
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	if p.Password != "" || p.AuthUser != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", p.AuthUser, p.Password, p.Host)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("auth: %w", err)
			}
		}
	}

	from := p.Message.FromAddr
	if from == "" {
		from = p.AuthUser
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	for _, rcpt := range allRecipients(p.Message) {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("RCPT TO %s: %w", rcpt, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err := w.Write(buildMIME(from, p.Message)); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close body: %w", err)
	}
	return client.Quit()
}

func allRecipients(m Message) []string {
	out := make([]string, 0, len(m.To)+len(m.Cc)+len(m.Bcc))
	out = append(out, m.To...)
	out = append(out, m.Cc...)
	out = append(out, m.Bcc...)
	return out
}

// buildMIME assembles an RFC 5322 message. Bodies are base64-encoded so UTF-8
// (e.g. Japanese) content survives transport intact.
func buildMIME(from string, m Message) []byte {
	var b bytes.Buffer
	fromHeader := from
	if m.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", m.FromName), from)
	}
	writeHeader(&b, "From", fromHeader)
	writeHeader(&b, "To", strings.Join(m.To, ", "))
	if len(m.Cc) > 0 {
		writeHeader(&b, "Cc", strings.Join(m.Cc, ", "))
	}
	writeHeader(&b, "Subject", mime.QEncoding.Encode("utf-8", m.Subject))
	writeHeader(&b, "Date", time.Now().Format(time.RFC1123Z))
	writeHeader(&b, "MIME-Version", "1.0")
	if m.InReplyTo != "" {
		writeHeader(&b, "In-Reply-To", m.InReplyTo)
		refs := m.References
		if refs == "" {
			refs = m.InReplyTo
		}
		writeHeader(&b, "References", refs)
	}

	if m.HTML != "" {
		boundary := "konatsu-alt-boundary-7f3c1a"
		writeHeader(&b, "Content-Type", fmt.Sprintf("multipart/alternative; boundary=%q", boundary))
		b.WriteString("\r\n")
		writePart(&b, boundary, "text/plain; charset=utf-8", m.Text)
		writePart(&b, boundary, "text/html; charset=utf-8", m.HTML)
		fmt.Fprintf(&b, "--%s--\r\n", boundary)
	} else {
		writeHeader(&b, "Content-Type", "text/plain; charset=utf-8")
		writeHeader(&b, "Content-Transfer-Encoding", "base64")
		b.WriteString("\r\n")
		b.WriteString(base64Wrap(m.Text))
	}
	return b.Bytes()
}

func writeHeader(b *bytes.Buffer, key, value string) {
	fmt.Fprintf(b, "%s: %s\r\n", key, value)
}

func writePart(b *bytes.Buffer, boundary, contentType, body string) {
	fmt.Fprintf(b, "--%s\r\n", boundary)
	fmt.Fprintf(b, "Content-Type: %s\r\n", contentType)
	b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	b.WriteString(base64Wrap(body))
}

// base64Wrap encodes s and wraps it at 76 columns per RFC 2045.
func base64Wrap(s string) string {
	enc := base64.StdEncoding.EncodeToString([]byte(s))
	var b strings.Builder
	for len(enc) > 76 {
		b.WriteString(enc[:76])
		b.WriteString("\r\n")
		enc = enc[76:]
	}
	b.WriteString(enc)
	b.WriteString("\r\n")
	return b.String()
}
