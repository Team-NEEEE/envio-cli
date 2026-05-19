package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	EnvironmentEncryptionAlgorithm      = "aes-256-gcm"
	EnvironmentValueEncryptionAlgorithm = "aes-256-gcm-per-value"
	environmentEncryptionEncoding       = "base64"
	environmentEncryptionVersion        = 1
	environmentValueEncryptionVersion   = 2
)

func EncryptEnvironment(plaintext []byte, projectMasterKey []byte) (map[string]any, error) {
	aead, err := projectMasterKeyAEAD(projectMasterKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate environment nonce: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return map[string]any{
		"schemaVersion": environmentEncryptionVersion,
		"algorithm":     EnvironmentEncryptionAlgorithm,
		"encoding":      environmentEncryptionEncoding,
		"nonce":         base64.StdEncoding.EncodeToString(nonce),
		"ciphertext":    base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

func EncryptEnvironmentValues(values map[string]string, projectMasterKey []byte) (map[string]any, error) {
	aead, err := projectMasterKeyAEAD(projectMasterKey)
	if err != nil {
		return nil, err
	}

	encryptedValues := make(map[string]any, len(values))
	for _, key := range sortedEnvironmentKeys(values) {
		nonce := make([]byte, aead.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, fmt.Errorf("generate environment value nonce: %w", err)
		}

		ciphertext := aead.Seal(nil, nonce, []byte(values[key]), []byte(key))
		encryptedValues[key] = map[string]any{
			"nonce":      base64.StdEncoding.EncodeToString(nonce),
			"ciphertext": base64.StdEncoding.EncodeToString(ciphertext),
		}
	}

	return map[string]any{
		"schemaVersion": environmentValueEncryptionVersion,
		"algorithm":     EnvironmentValueEncryptionAlgorithm,
		"encoding":      environmentEncryptionEncoding,
		"variables":     encryptedValues,
	}, nil
}

func DecryptEnvironment(encrypted map[string]any, projectMasterKey []byte) ([]byte, error) {
	if len(encrypted) == 0 {
		return nil, errors.New("encrypted environment is empty")
	}

	algorithm, err := encryptedString(encrypted, "algorithm")
	if err != nil {
		return nil, err
	}
	switch algorithm {
	case EnvironmentEncryptionAlgorithm:
		return decryptWholeEnvironment(encrypted, projectMasterKey)
	case EnvironmentValueEncryptionAlgorithm:
		return decryptEnvironmentValues(encrypted, projectMasterKey)
	default:
		return nil, fmt.Errorf("unsupported environment encryption algorithm: %s", algorithm)
	}
}

func decryptWholeEnvironment(encrypted map[string]any, projectMasterKey []byte) ([]byte, error) {
	encoding, err := encryptedString(encrypted, "encoding")
	if err != nil {
		return nil, err
	}
	if encoding != environmentEncryptionEncoding {
		return nil, fmt.Errorf("unsupported environment encoding: %s", encoding)
	}

	nonceText, err := encryptedString(encrypted, "nonce")
	if err != nil {
		return nil, err
	}
	ciphertextText, err := encryptedString(encrypted, "ciphertext")
	if err != nil {
		return nil, err
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceText)
	if err != nil {
		return nil, fmt.Errorf("decode environment nonce: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextText)
	if err != nil {
		return nil, fmt.Errorf("decode environment ciphertext: %w", err)
	}

	aead, err := projectMasterKeyAEAD(projectMasterKey)
	if err != nil {
		return nil, err
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt environment: %w", err)
	}
	return plaintext, nil
}

func decryptEnvironmentValues(encrypted map[string]any, projectMasterKey []byte) ([]byte, error) {
	encoding, err := encryptedString(encrypted, "encoding")
	if err != nil {
		return nil, err
	}
	if encoding != environmentEncryptionEncoding {
		return nil, fmt.Errorf("unsupported environment encoding: %s", encoding)
	}

	variables, err := encryptedObject(encrypted, "variables")
	if err != nil {
		return nil, err
	}

	aead, err := projectMasterKeyAEAD(projectMasterKey)
	if err != nil {
		return nil, err
	}

	values := make(map[string]string, len(variables))
	for _, key := range sortedAnyKeys(variables) {
		entry, err := encryptedObject(variables, key)
		if err != nil {
			return nil, fmt.Errorf("encrypted environment variable %s: %w", key, err)
		}
		nonceText, err := encryptedString(entry, "nonce")
		if err != nil {
			return nil, fmt.Errorf("encrypted environment variable %s: %w", key, err)
		}
		ciphertextText, err := encryptedString(entry, "ciphertext")
		if err != nil {
			return nil, fmt.Errorf("encrypted environment variable %s: %w", key, err)
		}

		nonce, err := base64.StdEncoding.DecodeString(nonceText)
		if err != nil {
			return nil, fmt.Errorf("decode environment value nonce for %s: %w", key, err)
		}
		ciphertext, err := base64.StdEncoding.DecodeString(ciphertextText)
		if err != nil {
			return nil, fmt.Errorf("decode environment value ciphertext for %s: %w", key, err)
		}
		plaintext, err := aead.Open(nil, nonce, ciphertext, []byte(key))
		if err != nil {
			return nil, fmt.Errorf("decrypt environment value for %s: %w", key, err)
		}
		values[key] = string(plaintext)
	}

	return []byte(formatDotenv(values)), nil
}

func projectMasterKeyAEAD(projectMasterKey []byte) (cipher.AEAD, error) {
	if err := ValidateProjectMasterKey(projectMasterKey); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(projectMasterKey)
	if err != nil {
		return nil, fmt.Errorf("create environment cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create environment gcm: %w", err)
	}
	return aead, nil
}

func sortedEnvironmentKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedAnyKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func formatDotenv(values map[string]string) string {
	var builder strings.Builder
	for _, key := range sortedEnvironmentKeys(values) {
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(formatDotenvValue(values[key]))
		builder.WriteByte('\n')
	}
	return builder.String()
}

func formatDotenvValue(value string) string {
	if value == "" {
		return ""
	}
	if strings.Contains(value, " #") || strings.ContainsAny(value, "\"'\\") || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return strconv.Quote(value)
	}
	return value
}

func encryptedObject(encrypted map[string]any, key string) (map[string]any, error) {
	value, ok := encrypted[key]
	if !ok {
		return nil, fmt.Errorf("encrypted environment missing %s", key)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("encrypted environment %s must be an object", key)
	}
	return object, nil
}

func encryptedString(encrypted map[string]any, key string) (string, error) {
	value, ok := encrypted[key]
	if !ok {
		return "", fmt.Errorf("encrypted environment missing %s", key)
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("encrypted environment %s must be a string", key)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("encrypted environment %s is empty", key)
	}
	return text, nil
}
