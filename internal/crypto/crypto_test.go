package crypto

import (
	"bytes"
	"testing"
)

func key32() []byte { return []byte("0123456789abcdef0123456789abcdef") }

func TestNewAES256GCM_KeyLength(t *testing.T) {
	if _, err := NewAES256GCM([]byte("short")); err == nil {
		t.Fatal("expected error for short key")
	}
	if _, err := NewAES256GCM(key32()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	c, err := NewAES256GCM(key32())
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("秘密のパスワード123")
	ct, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(ct, plain) {
		t.Fatal("ciphertext should differ from plaintext")
	}
	got, err := c.Decrypt(ct)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("roundtrip mismatch: %q != %q", got, plain)
	}
}

func TestDecrypt_TooShort(t *testing.T) {
	c, _ := NewAES256GCM(key32())
	if _, err := c.Decrypt([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected error for short ciphertext")
	}
}

func TestEncrypt_UniqueNonce(t *testing.T) {
	c, _ := NewAES256GCM(key32())
	a, _ := c.Encrypt([]byte("x"))
	b, _ := c.Encrypt([]byte("x"))
	if bytes.Equal(a, b) {
		t.Fatal("two encryptions of same plaintext should differ (random nonce)")
	}
}
