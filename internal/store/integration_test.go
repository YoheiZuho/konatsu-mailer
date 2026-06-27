//go:build integration

// Integration tests run against a real PostgreSQL. They are skipped unless
// TEST_DATABASE_URL is set. Run with: make test-integration
package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

func testDB(t *testing.T) *DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	abs, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(url, abs); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db, err := New(url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	return db
}

func seedUserAccount(t *testing.T, db *DB, ctx context.Context) (domain.UUID, domain.UUID, func()) {
	t.Helper()
	email := "it-" + time.Now().Format("20060102150405.000000000") + "@example.com"
	var uid domain.UUID
	if err := db.Pool.QueryRow(ctx,
		`INSERT INTO users(email, password_hash) VALUES($1,'x') RETURNING id`, email).Scan(&uid); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var aid domain.UUID
	if err := db.Pool.QueryRow(ctx,
		`INSERT INTO accounts(user_id, email, imap_host, smtp_host, auth_user, password_encrypted)
		 VALUES($1,$2,'imap.example.com','smtp.example.com','u',$3) RETURNING id`,
		uid, email, []byte{0, 1, 2}).Scan(&aid); err != nil {
		t.Fatalf("insert account: %v", err)
	}
	cleanup := func() { _, _ = db.Pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, uid) }
	return uid, aid, cleanup
}

func TestStoreEmailFlow_Integration(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()
	uid, aid, cleanup := seedUserAccount(t, db, ctx)
	defer cleanup()

	tid, err := db.UpsertThread(ctx, aid, "thread-key", "件名 invoice", time.Now())
	if err != nil {
		t.Fatalf("UpsertThread: %v", err)
	}
	email := &domain.Email{
		AccountID: aid, ThreadID: &tid, Folder: "INBOX", IMAPUID: 1,
		Subject: "Hello invoice 請求", SenderName: "A", SenderAddr: "a@x.com",
		BodyPreview: "body preview text", Category: "primary", DateSent: time.Now(),
		Recipients: domain.Recipients{{Addr: "me@x.com", Type: "to"}},
	}
	id, inserted, err := db.UpsertEmail(ctx, email)
	if err != nil || !inserted {
		t.Fatalf("UpsertEmail: id=%v inserted=%v err=%v", id, inserted, err)
	}
	// Re-upsert is idempotent (not inserted again).
	if _, ins2, _ := db.UpsertEmail(ctx, email); ins2 {
		t.Fatal("re-upsert should not insert")
	}

	rec, err := db.GetEmailForUser(ctx, string(id), string(uid))
	if err != nil || rec.Subject != email.Subject {
		t.Fatalf("GetEmailForUser: %+v err=%v", rec, err)
	}

	// Labels.
	lid, _, err := db.GetOrCreateLabel(ctx, aid, "仕事", true)
	if err != nil {
		t.Fatalf("GetOrCreateLabel: %v", err)
	}
	if err := db.LinkEmailLabel(ctx, id, lid, "ai"); err != nil {
		t.Fatalf("LinkEmailLabel: %v", err)
	}
	labels, _ := db.ListLabelsForUser(ctx, string(uid))
	if len(labels) == 0 {
		t.Fatal("expected at least one label")
	}

	// Unread counts then mark-read.
	counts, _, _, err := db.UnreadCounts(ctx, string(uid))
	if err != nil || counts["INBOX"] != 1 {
		t.Fatalf("UnreadCounts before: %v err=%v", counts, err)
	}
	if err := db.SetReadForUser(ctx, string(id), string(uid), true); err != nil {
		t.Fatal(err)
	}
	counts2, _, _, _ := db.UnreadCounts(ctx, string(uid))
	if counts2["INBOX"] != 0 {
		t.Fatalf("UnreadCounts after read: %v", counts2)
	}

	// Body cache roundtrip.
	if err := db.SaveEmailBody(ctx, id, CachedBody{Text: "full body", HTML: "<p>x</p>"}); err != nil {
		t.Fatal(err)
	}
	cb, ok, _ := db.GetEmailBody(ctx, id)
	if !ok || cb.Text != "full body" {
		t.Fatalf("GetEmailBody: %+v ok=%v", cb, ok)
	}
}

func TestStoreFiltersAndLLM_Integration(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()
	uid, _, cleanup := seedUserAccount(t, db, ctx)
	defer cleanup()

	f := domain.Filter{
		Name: "請求書", Enabled: true, MatchType: "all",
		Conditions: []domain.FilterCondition{{Field: "subject", Op: "contains", Value: "請求"}},
		Actions:    []domain.FilterAction{{Type: "add_label", Value: "請求書"}},
	}
	created, err := db.CreateFilter(ctx, string(uid), &f)
	if err != nil || created.Name != "請求書" {
		t.Fatalf("CreateFilter: %+v err=%v", created, err)
	}
	if list, _ := db.ListFilters(ctx, string(uid)); len(list) != 1 {
		t.Fatalf("ListFilters len = %d", len(list))
	}
	if enabled, _ := db.EnabledFiltersForUser(ctx, uid); len(enabled) != 1 {
		t.Fatalf("EnabledFiltersForUser len = %d", len(enabled))
	}

	lc := domain.LLMConfig{Name: "OpenAI", BaseURL: "https://api.openai.com/v1", Model: "gpt-4o-mini", IsDefault: true, IsActive: true}
	if _, err := db.CreateLLMConfig(ctx, string(uid), &lc); err != nil {
		t.Fatalf("CreateLLMConfig: %v", err)
	}
	def, err := db.DefaultLLMConfig(ctx, uid)
	if err != nil || def.Name != "OpenAI" {
		t.Fatalf("DefaultLLMConfig: %+v err=%v", def, err)
	}

	if err := db.SavePushSubscription(ctx, string(uid), "https://push/ep", "p256", "auth", "ua"); err != nil {
		t.Fatal(err)
	}
	subs, _ := db.ListPushSubscriptions(ctx, uid)
	if len(subs) != 1 {
		t.Fatalf("ListPushSubscriptions len = %d", len(subs))
	}
}
