package analysis

import (
	"reflect"
	"testing"
)

func TestSkipCategory(t *testing.T) {
	prefs := map[string]any{
		"ai_filters": map[string]any{"promotions": true, "social": false},
	}
	if !skipCategory("promotions", prefs) {
		t.Error("promotions should be skipped")
	}
	if skipCategory("social", prefs) {
		t.Error("social should not be skipped (false)")
	}
	if skipCategory("primary", prefs) {
		t.Error("primary never skipped")
	}
	if skipCategory("newsletters", prefs) {
		t.Error("unset category defaults to not-skip")
	}
	if skipCategory("promotions", map[string]any{}) {
		t.Error("no ai_filters → not skipped")
	}
}

func TestMergeCandidates(t *testing.T) {
	got := mergeCandidates([]string{"仕事", "カスタム"})
	// account labels first, then defaults, de-duplicated
	if got[0] != "仕事" || got[1] != "カスタム" {
		t.Errorf("account labels should come first: %v", got[:2])
	}
	seen := map[string]int{}
	for _, l := range got {
		seen[l]++
		if seen[l] > 1 {
			t.Errorf("duplicate candidate: %q", l)
		}
	}
	// "仕事" is also a default; should not be duplicated.
}

func TestIntersects(t *testing.T) {
	if !intersects([]string{"a", "b"}, []string{"b", "c"}) {
		t.Error("should intersect on b")
	}
	if intersects([]string{"a"}, []string{"x", "y"}) {
		t.Error("should not intersect")
	}
}

func TestStringSlice(t *testing.T) {
	if got := stringSlice([]any{"a", "b", 3}); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("stringSlice []any = %v", got)
	}
	if got := stringSlice([]string{"x"}); !reflect.DeepEqual(got, []string{"x"}) {
		t.Errorf("stringSlice []string = %v", got)
	}
	if got := stringSlice(42); got != nil {
		t.Errorf("stringSlice(non-slice) = %v, want nil", got)
	}
}
