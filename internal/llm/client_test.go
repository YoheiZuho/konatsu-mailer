package llm

import "testing"

func TestParseResult_Plain(t *testing.T) {
	r, err := parseResult(`{"summary":"要約","priority":4,"labels":["仕事"],"is_spam":false}`)
	if err != nil {
		t.Fatal(err)
	}
	if r.Summary != "要約" || r.Priority != 4 || len(r.Labels) != 1 || r.IsSpam {
		t.Fatalf("unexpected: %+v", r)
	}
}

func TestParseResult_Wrapped(t *testing.T) {
	// Model wraps JSON in prose / code fences.
	content := "```json\n{\"summary\":\"x\",\"priority\":2,\"labels\":[],\"is_spam\":true}\n```"
	r, err := parseResult(content)
	if err != nil {
		t.Fatal(err)
	}
	if r.Priority != 2 || !r.IsSpam {
		t.Fatalf("unexpected: %+v", r)
	}
}

func TestParseResult_Invalid(t *testing.T) {
	if _, err := parseResult("not json at all"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestValidate_ClampAndWhitelist(t *testing.T) {
	in := ClassifyResult{Summary: "s", Priority: 9, Labels: []string{"仕事", "存在しない"}}
	out := validate(in, []string{"仕事", "重要"})
	if out.Priority != 5 {
		t.Errorf("priority not clamped: %d", out.Priority)
	}
	if len(out.Labels) != 1 || out.Labels[0] != "仕事" {
		t.Errorf("labels not whitelisted: %v", out.Labels)
	}

	low := validate(ClassifyResult{Priority: 0}, nil)
	if low.Priority != 1 {
		t.Errorf("low priority not clamped to 1: %d", low.Priority)
	}
}

func TestValidate_SummaryTruncate(t *testing.T) {
	long := make([]rune, 300)
	for i := range long {
		long[i] = 'あ'
	}
	out := validate(ClassifyResult{Summary: string(long), Priority: 3}, nil)
	if len([]rune(out.Summary)) != 200 {
		t.Errorf("summary not truncated: %d", len([]rune(out.Summary)))
	}
}

func TestTruncate(t *testing.T) {
	if truncate("hello", 10) != "hello" {
		t.Error("short string changed")
	}
	if got := truncate("あいうえお", 3); got != "あいう" {
		t.Errorf("truncate = %q", got)
	}
}
