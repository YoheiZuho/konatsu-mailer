package config

import "testing"

func TestEnvBool(t *testing.T) {
	cases := []struct {
		val      string
		set      bool
		fallback bool
		want     bool
	}{
		{set: false, fallback: true, want: true},
		{set: false, fallback: false, want: false},
		{val: "true", set: true, fallback: false, want: true},
		{val: "false", set: true, fallback: true, want: false},
		{val: "1", set: true, fallback: false, want: true},
		{val: "0", set: true, fallback: true, want: false},
		{val: "notabool", set: true, fallback: true, want: true}, // parse error → fallback
	}
	const k = "KONATSU_TEST_BOOL"
	for _, c := range cases {
		t.Setenv(k, "")
		if c.set {
			t.Setenv(k, c.val)
		}
		if got := envBool(k, c.fallback); got != c.want {
			t.Errorf("envBool(%q set=%v fb=%v) = %v, want %v", c.val, c.set, c.fallback, got, c.want)
		}
	}
}

func TestEnv_Fallback(t *testing.T) {
	t.Setenv("KONATSU_TEST_STR", "")
	if got := env("KONATSU_TEST_STR", "def"); got != "def" {
		t.Errorf("env fallback = %q, want def", got)
	}
	t.Setenv("KONATSU_TEST_STR", "val")
	if got := env("KONATSU_TEST_STR", "def"); got != "val" {
		t.Errorf("env value = %q, want val", got)
	}
}
