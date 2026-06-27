package imapsync

import (
	"strings"
	"testing"

	"github.com/emersion/go-imap/v2"
	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

func TestNormalizeSubject(t *testing.T) {
	cases := map[string]string{
		"Re: Hello":        "hello",
		"RE: FWD: Report":  "report",
		"Fwd:  spaced   x": "spaced x",
		"":                 "",
		"Plain":            "plain",
	}
	for in, want := range cases {
		if got := normalizeSubject(in); got != want {
			t.Errorf("normalizeSubject(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestThreadKey(t *testing.T) {
	withSubject := &domain.Email{Subject: "Re: 週次報告"}
	if got := threadKey(withSubject); got != "週次報告" {
		t.Errorf("threadKey subject = %q", got)
	}
	noSubject := &domain.Email{Subject: "", MessageID: "<abc@x>"}
	if got := threadKey(noSubject); got != "<abc@x>" {
		t.Errorf("threadKey messageID = %q", got)
	}
	bare := &domain.Email{IMAPUID: 42}
	if got := threadKey(bare); got != "uid:42" {
		t.Errorf("threadKey uid = %q", got)
	}
}

func TestCleanUTF8(t *testing.T) {
	if got := cleanUTF8("ok日本語"); got != "ok日本語" {
		t.Errorf("valid utf8 changed: %q", got)
	}
	got := cleanUTF8("abc\xe3\x81") // truncated multibyte
	if !strings.HasPrefix(got, "abc") || strings.ContainsRune(got, '�') {
		t.Errorf("cleanUTF8 should drop invalid bytes, got %q", got)
	}
}

func TestPreviewText(t *testing.T) {
	if got := previewText(nil); got != "" {
		t.Errorf("empty preview = %q", got)
	}
	got := previewText([]byte("  hello\n\nworld  "))
	if got != "hello world" {
		t.Errorf("previewText collapse = %q", got)
	}
	long := previewText([]byte(strings.Repeat("あ", 600)))
	if len([]rune(long)) != 500 {
		t.Errorf("previewText should cap at 500 runes, got %d", len([]rune(long)))
	}
}

func TestHasFlag(t *testing.T) {
	flags := []imap.Flag{imap.FlagSeen, imap.FlagFlagged}
	if !hasFlag(flags, imap.FlagSeen) {
		t.Error("should have Seen")
	}
	if hasFlag(flags, imap.FlagDraft) {
		t.Error("should not have Draft")
	}
}
