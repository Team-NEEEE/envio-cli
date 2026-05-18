package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	EnvironmentEncryptionAlgorithm = "aes-256-gcm"
	environmentEncryptionEncoding  = "base64"
	environmentEncryptionVersion   = 1
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

func DecryptEnvironment(encrypted map[string]any, projectMasterKey []byte) ([]byte, error) {
	if len(encrypted) == 0 {
		return nil, errors.New("encrypted environment is empty")
	}

	algorithm, err := encryptedString(encrypted, "algorithm")
	if err != nil {
		return nil, err
	}
	if algorithm != EnvironmentEncryptionAlgorithm {
		return nil, fmt.Errorf("unsupported environment encryption algorithm: %s", algorithm)
	}

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
