package imapsync

import (
	"testing"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

func TestCategorize(t *testing.T) {
	cases := []struct {
		header string
		sender string
		want   string
	}{
		{"", "boss@example.com", "primary"},
		{"List-Id: <news.example.com>", "news@example.com", "newsletters"},
		{"List-Unsubscribe: <mailto:x>", "promo@shop.com", "promotions"},
		{"Precedence: bulk", "x@y.com", "promotions"},
		{"", "notifications@github.com", "social"},
		{"List-Id: x", "noreply@facebook.com", "social"}, // social wins over list-id
	}
	for _, c := range cases {
		if got := categorize(c.header, c.sender); got != c.want {
			t.Errorf("categorize(%q, %q) = %q, want %q", c.header, c.sender, got, c.want)
		}
	}
}

func mkEmail() *domain.Email {
	return &domain.Email{
		Subject:    "請求書のご案内",
		SenderName: "経理部",
		SenderAddr: "billing@example.co.jp",
		BodyPreview: "今月分の請求書を送付します",
		Recipients: domain.Recipients{
			{Name: "自分", Addr: "me@example.com", Type: "to"},
			{Name: "上司", Addr: "boss@example.com", Type: "cc"},
		},
	}
}

func TestConditionMatches(t *testing.T) {
	e := mkEmail()
	cases := []struct {
		cond domain.FilterCondition
		want bool
	}{
		{domain.FilterCondition{Field: "subject", Op: "contains", Value: "請求書"}, true},
		{domain.FilterCondition{Field: "subject", Op: "not_contains", Value: "領収書"}, true},
		{domain.FilterCondition{Field: "from", Op: "contains", Value: "billing@"}, true},
		{domain.FilterCondition{Field: "from", Op: "is", Value: "x"}, false},
		{domain.FilterCondition{Field: "to", Op: "contains", Value: "me@example.com"}, true},
		{domain.FilterCondition{Field: "cc", Op: "contains", Value: "boss@example.com"}, true},
		{domain.FilterCondition{Field: "cc", Op: "contains", Value: "me@example.com"}, false},
		{domain.FilterCondition{Field: "subject", Op: "starts_with", Value: "請求"}, true},
		{domain.FilterCondition{Field: "subject", Op: "ends_with", Value: "案内"}, true},
		{domain.FilterCondition{Field: "body", Op: "contains", Value: "送付"}, true},
		{domain.FilterCondition{Field: "unknown", Op: "contains", Value: "x"}, false},
	}
	for _, c := range cases {
		if got := conditionMatches(c.cond, e); got != c.want {
			t.Errorf("conditionMatches(%+v) = %v, want %v", c.cond, got, c.want)
		}
	}
}

func TestFilterMatches_AllAny(t *testing.T) {
	e := mkEmail()
	all := domain.Filter{MatchType: "all", Conditions: []domain.FilterCondition{
		{Field: "subject", Op: "contains", Value: "請求書"},
		{Field: "from", Op: "contains", Value: "billing"},
	}}
	if !filterMatches(all, e) {
		t.Error("all-match should pass")
	}
	allFail := all
	allFail.Conditions = append([]domain.FilterCondition{{Field: "subject", Op: "contains", Value: "ZZZ"}}, all.Conditions...)
	if filterMatches(allFail, e) {
		t.Error("all-match with one failing condition should fail")
	}
	any := domain.Filter{MatchType: "any", Conditions: []domain.FilterCondition{
		{Field: "subject", Op: "contains", Value: "ZZZ"},
		{Field: "from", Op: "contains", Value: "billing"},
	}}
	if !filterMatches(any, e) {
		t.Error("any-match should pass when one matches")
	}
	if filterMatches(domain.Filter{MatchType: "all"}, e) {
		t.Error("filter with no conditions should not match")
	}
}
