package crypto

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncryptEnvironmentRoundTrip(t *testing.T) {
	t.Parallel()

	masterKey := []byte("12345678901234567890123456789012")
	plaintext := []byte("DATABASE_URL=postgres://user:pass@example/db\nAPI_KEY=secret\n")

	encrypted, err := EncryptEnvironment(plaintext, masterKey)
	if err != nil {
		t.Fatalf("EncryptEnvironment() error = %v", err)
	}
	if encrypted["algorithm"] != EnvironmentEncryptionAlgorithm {
		t.Fatalf("algorithm = %#v", encrypted["algorithm"])
	}
	if strings.Contains(asStringForTest(encrypted["ciphertext"]), "DATABASE_URL") ||
		strings.Contains(asStringForTest(encrypted["ciphertext"]), "secret") {
		t.Fatalf("ciphertext should not contain plaintext: %#v", encrypted)
	}

	got, err := DecryptEnvironment(encrypted, masterKey)
	if err != nil {
		t.Fatalf("DecryptEnvironment() error = %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("decrypted = %q, want %q", got, plaintext)
	}
}

func TestDecryptEnvironmentRejectsWrongKey(t *testing.T) {
	t.Parallel()

	masterKey := []byte("12345678901234567890123456789012")
	encrypted, err := EncryptEnvironment([]byte("API_KEY=secret\n"), masterKey)
	if err != nil {
		t.Fatalf("EncryptEnvironment() error = %v", err)
	}

	_, err = DecryptEnvironment(encrypted, []byte("abcdefghijklmnopqrstuvwxzy123456"))
	if err == nil {
		t.Fatal("DecryptEnvironment() error = nil, want failure")
	}
}

func asStringForTest(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}
