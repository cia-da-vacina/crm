package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
)

func newTestCipher(t *testing.T) *Cipher {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	t.Setenv("APP_ENCRYPTION_KEY", hex.EncodeToString(key))
	c, err := NewCipherFromEnv()
	if err != nil {
		t.Fatalf("NewCipherFromEnv failed: %v", err)
	}
	return c
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	c := newTestCipher(t)

	plaintext := "EAAG1234567890abcdefTOKEN"
	ciphertext, err := c.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if string(ciphertext) == plaintext {
		t.Fatal("ciphertext must not equal plaintext")
	}

	decrypted, err := c.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecrypt_TamperedCiphertextFails(t *testing.T) {
	c := newTestCipher(t)

	ciphertext, err := c.Encrypt("secret")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	ciphertext[len(ciphertext)-1] ^= 0xFF

	if _, err := c.Decrypt(ciphertext); err == nil {
		t.Fatal("expected error decrypting tampered ciphertext, got nil")
	}
}

func TestNewCipherFromEnv_MissingKey(t *testing.T) {
	os.Unsetenv("APP_ENCRYPTION_KEY")
	if _, err := NewCipherFromEnv(); err == nil {
		t.Fatal("expected error when APP_ENCRYPTION_KEY is unset")
	}
}
