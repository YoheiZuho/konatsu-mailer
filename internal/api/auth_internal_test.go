package api

import (
	"testing"
	"time"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

func TestIssueAndParseToken(t *testing.T) {
	secret := "test-secret"
	tok, err := issueToken(secret, "user-123", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if got := userIDFromToken(tok, secret); got != "user-123" {
		t.Errorf("userIDFromToken = %q, want user-123", got)
	}
	if got := userIDFromToken(tok, "wrong-secret"); got != "" {
		t.Errorf("wrong secret should yield empty, got %q", got)
	}
	if got := userIDFromToken("garbage", secret); got != "" {
		t.Errorf("garbage token should yield empty, got %q", got)
	}
}

func TestTokenFromProtocolHeader(t *testing.T) {
	cases := map[string]string{
		"bearer, abc.def.ghi": "abc.def.ghi",
		"bearer,abc":          "abc",
		"bearer":              "",
		"":                    "",
	}
	for in, want := range cases {
		if got := tokenFromProtocolHeader(in); got != want {
			t.Errorf("tokenFromProtocolHeader(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDeref(t *testing.T) {
	if deref(nil) != "" {
		t.Error("nil deref should be empty")
	}
	s := "x"
	if deref(&s) != "x" {
		t.Error("deref mismatch")
	}
}

func TestRecipientsByType(t *testing.T) {
	rs := domain.Recipients{
		{Name: "A", Addr: "a@x", Type: "to"},
		{Name: "B", Addr: "b@x", Type: "cc"},
		{Name: "C", Addr: "c@x", Type: "to"},
	}
	to := recipientsByType(rs, "to")
	if len(to) != 2 {
		t.Fatalf("want 2 to-recipients, got %d", len(to))
	}
	if to[0]["addr"] != "a@x" {
		t.Errorf("unexpected addr: %v", to[0])
	}
	if len(recipientsByType(rs, "bcc")) != 0 {
		t.Error("want 0 bcc")
	}
}

func TestFilterReqToFilter_Defaults(t *testing.T) {
	f := filterReq{Name: "x", MatchType: "weird"}.toFilter()
	if f.MatchType != "all" {
		t.Errorf("invalid match_type should default to all, got %q", f.MatchType)
	}
	if !f.Enabled {
		t.Error("enabled should default true")
	}
	if f.Conditions == nil || f.Actions == nil {
		t.Error("conditions/actions should be non-nil slices")
	}
}
