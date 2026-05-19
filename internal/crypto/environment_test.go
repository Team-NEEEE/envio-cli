package crypto

import (
	"bytes"
	"encoding/json"
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

func TestEncryptEnvironmentValuesLeavesKeysPlainAndEncryptsValues(t *testing.T) {
	t.Parallel()

	masterKey := []byte("12345678901234567890123456789012")
	encrypted, err := EncryptEnvironmentValues(map[string]string{
		"DATABASE_URL": "postgres://user:pass@example/db",
		"API_KEY":      "secret",
	}, masterKey)
	if err != nil {
		t.Fatalf("EncryptEnvironmentValues() error = %v", err)
	}
	if encrypted["algorithm"] != EnvironmentValueEncryptionAlgorithm {
		t.Fatalf("algorithm = %#v", encrypted["algorithm"])
	}
	variables, ok := encrypted["variables"].(map[string]any)
	if !ok {
		t.Fatalf("variables = %#v", encrypted["variables"])
	}
	if _, ok := variables["API_KEY"]; !ok {
		t.Fatalf("variables should include plaintext key API_KEY: %#v", variables)
	}

	rawJSON, err := json.Marshal(encrypted)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if strings.Contains(string(rawJSON), "secret") ||
		strings.Contains(string(rawJSON), "postgres://user:pass@example/db") {
		t.Fatalf("encrypted environment leaked plaintext values: %s", string(rawJSON))
	}

	got, err := DecryptEnvironment(encrypted, masterKey)
	if err != nil {
		t.Fatalf("DecryptEnvironment() error = %v", err)
	}
	want := []byte("API_KEY=secret\nDATABASE_URL=postgres://user:pass@example/db\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("decrypted = %q, want %q", got, want)
	}
}

func TestDecryptEnvironmentValuesBindsCiphertextToKey(t *testing.T) {
	t.Parallel()

	masterKey := []byte("12345678901234567890123456789012")
	encrypted, err := EncryptEnvironmentValues(map[string]string{"API_KEY": "secret"}, masterKey)
	if err != nil {
		t.Fatalf("EncryptEnvironmentValues() error = %v", err)
	}
	variables := encrypted["variables"].(map[string]any)
	variables["OTHER_KEY"] = variables["API_KEY"]
	delete(variables, "API_KEY")

	_, err = DecryptEnvironment(encrypted, masterKey)
	if err == nil {
		t.Fatal("DecryptEnvironment() error = nil, want failure")
	}
}

func TestDecryptEnvironmentValuesQuotesDotenvValuesWhenNeeded(t *testing.T) {
	t.Parallel()

	masterKey := []byte("12345678901234567890123456789012")
	encrypted, err := EncryptEnvironmentValues(map[string]string{
		"EMPTY": "",
		"TEXT":  "hello world # not-comment",
	}, masterKey)
	if err != nil {
		t.Fatalf("EncryptEnvironmentValues() error = %v", err)
	}

	got, err := DecryptEnvironment(encrypted, masterKey)
	if err != nil {
		t.Fatalf("DecryptEnvironment() error = %v", err)
	}
	want := []byte("EMPTY=\nTEXT=\"hello world # not-comment\"\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("decrypted = %q, want %q", got, want)
	}
}

func asStringForTest(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}
