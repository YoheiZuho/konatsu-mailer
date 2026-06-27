package api

import (
	"testing"
	"time"
)

func TestValidateLLMBaseURL(t *testing.T) {
	cases := []struct {
		raw          string
		allowPrivate bool
		wantErr      bool
	}{
		{"http://8.8.8.8/v1", false, false},          // public IP literal
		{"https://1.1.1.1/v1", false, false},         // public IP literal
		{"ftp://8.8.8.8/", false, true},              // bad scheme
		{"notaurl", false, true},                     // no scheme/host
		{"http://169.254.169.254/latest", true, true}, // metadata link-local, always blocked
		{"http://127.0.0.1:11434/v1", true, false},   // loopback allowed when allowPrivate
		{"http://127.0.0.1:11434/v1", false, true},   // loopback blocked when !allowPrivate
		{"http://10.0.0.5/v1", false, true},          // private blocked when !allowPrivate
		{"http://10.0.0.5/v1", true, false},          // private allowed when allowPrivate
	}
	for _, c := range cases {
		err := validateLLMBaseURL(c.raw, c.allowPrivate)
		if (err != nil) != c.wantErr {
			t.Errorf("validateLLMBaseURL(%q, allowPrivate=%v) err=%v, wantErr=%v", c.raw, c.allowPrivate, err, c.wantErr)
		}
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := newRateLimiter(2, time.Minute)
	if !rl.allow("ip-a") || !rl.allow("ip-a") {
		t.Fatal("first two requests should be allowed")
	}
	if rl.allow("ip-a") {
		t.Fatal("third request should be blocked")
	}
	if !rl.allow("ip-b") {
		t.Fatal("a different key should be allowed")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := newRateLimiter(1, time.Millisecond)
	if !rl.allow("k") {
		t.Fatal("first allowed")
	}
	if rl.allow("k") {
		t.Fatal("second blocked within window")
	}
	time.Sleep(2 * time.Millisecond)
	if !rl.allow("k") {
		t.Fatal("should be allowed after window reset")
	}
}

func TestSanitizeHTML(t *testing.T) {
	in := `<p onclick="x">hi</p><script>alert(1)</script><a href="http://x">l</a>`
	out := sanitizeHTML(in)
	if contains(out, "<script") || contains(out, "onclick") {
		t.Errorf("dangerous content not stripped: %q", out)
	}
	if !contains(out, "hi") || !contains(out, "<a") {
		t.Errorf("safe content removed: %q", out)
	}
	if sanitizeHTML("") != "" {
		t.Error("empty stays empty")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
