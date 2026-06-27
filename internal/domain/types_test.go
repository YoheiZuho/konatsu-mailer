package domain

import "testing"

func TestRecipients_ValueScanRoundtrip(t *testing.T) {
	in := Recipients{{Name: "山田", Addr: "y@x", Type: "to"}}
	v, err := in.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out Recipients
	if err := out.Scan(v.([]byte)); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Addr != "y@x" || out[0].Type != "to" {
		t.Fatalf("roundtrip mismatch: %+v", out)
	}
}

func TestSyncState_ValueScanRoundtrip(t *testing.T) {
	in := SyncState{"INBOX": {UIDValidity: 42, LastUID: 7}}
	v, err := in.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out SyncState
	if err := out.Scan(v.([]byte)); err != nil {
		t.Fatal(err)
	}
	if out["INBOX"].UIDValidity != 42 || out["INBOX"].LastUID != 7 {
		t.Fatalf("roundtrip mismatch: %+v", out)
	}
}

func TestScan_WrongType(t *testing.T) {
	var r Recipients
	if err := r.Scan("not bytes"); err == nil {
		t.Error("expected error scanning non-[]byte")
	}
}
