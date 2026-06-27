package smtpsend

import (
	"strings"
	"testing"
)

func TestAllRecipients(t *testing.T) {
	m := Message{To: []string{"a@x"}, Cc: []string{"b@x"}, Bcc: []string{"c@x"}}
	got := allRecipients(m)
	if len(got) != 3 {
		t.Fatalf("want 3 recipients, got %v", got)
	}
}

func TestBase64Wrap(t *testing.T) {
	out := base64Wrap(strings.Repeat("A", 100))
	for _, line := range strings.Split(strings.TrimRight(out, "\r\n"), "\r\n") {
		if len(line) > 76 {
			t.Fatalf("line exceeds 76 cols: %d", len(line))
		}
	}
	if !strings.HasSuffix(out, "\r\n") {
		t.Error("should end with CRLF")
	}
}

func TestBuildMIME_PlainHeaders(t *testing.T) {
	msg := Message{To: []string{"to@x.com"}, Cc: []string{"cc@x.com"}, Subject: "件名テスト", Text: "本文"}
	out := string(buildMIME("from@x.com", msg))
	for _, want := range []string{"From: from@x.com", "To: to@x.com", "Cc: cc@x.com", "MIME-Version: 1.0", "text/plain"} {
		if !strings.Contains(out, want) {
			t.Errorf("MIME missing %q\n---\n%s", want, out)
		}
	}
	if !strings.Contains(out, "=?utf-8?") { // subject MIME-encoded
		t.Error("subject should be RFC2047 encoded")
	}
}

func TestBuildMIME_Multipart(t *testing.T) {
	msg := Message{To: []string{"to@x.com"}, Subject: "s", Text: "plain", HTML: "<b>html</b>"}
	out := string(buildMIME("from@x.com", msg))
	if !strings.Contains(out, "multipart/alternative") {
		t.Error("should be multipart/alternative when HTML present")
	}
	if !strings.Contains(out, "text/html") || !strings.Contains(out, "text/plain") {
		t.Error("should contain both parts")
	}
}

func TestBuildMIME_InReplyTo(t *testing.T) {
	msg := Message{To: []string{"to@x"}, Subject: "Re: s", Text: "t", InReplyTo: "<orig@x>"}
	out := string(buildMIME("from@x", msg))
	if !strings.Contains(out, "In-Reply-To: <orig@x>") || !strings.Contains(out, "References: <orig@x>") {
		t.Error("reply headers missing")
	}
}
