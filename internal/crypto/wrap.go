package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const projectMasterKeySize = 32

func GenerateProjectMasterKey() ([]byte, error) {
	key := make([]byte, projectMasterKeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("project master key generation failed: %w", err)
	}
	return key, nil
}

func ValidateProjectMasterKey(projectMasterKey []byte) error {
	if len(projectMasterKey) != projectMasterKeySize {
		return fmt.Errorf("project master key must be %d bytes", projectMasterKeySize)
	}
	return nil
}

func WrapProjectMasterKey(publicKeyValue string, projectMasterKey []byte) (string, error) {
	if err := ValidateProjectMasterKey(projectMasterKey); err != nil {
		return "", err
	}

	publicKey, err := parseRSAPublicKey(publicKeyValue)
	if err != nil {
		return "", err
	}

	wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, projectMasterKey, nil)
	if err != nil {
		return "", fmt.Errorf("wrap project master key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(wrapped), nil
}

func UnwrapProjectMasterKey(privateKeyPEM string, wrappedMasterKey string) ([]byte, error) {
	privateKey, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimSpace(wrappedMasterKey))
	if err != nil {
		return nil, fmt.Errorf("decode wrapped project master key: %w", err)
	}

	projectMasterKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("unwrap project master key: %w", err)
	}
	if err := ValidateProjectMasterKey(projectMasterKey); err != nil {
		return nil, err
	}
	return projectMasterKey, nil
}

func parseRSAPrivateKey(raw string) (*rsa.PrivateKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("private key is empty")
	}

	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, errors.New("private key PEM block is missing")
	}

	if privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return privateKey, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse RSA private key: %w", err)
	}
	privateKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}
	return privateKey, nil
}

func parseRSAPublicKey(raw string) (*rsa.PublicKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("public key is empty")
	}

	block, _ := pem.Decode([]byte(raw))
	if block != nil {
		return parsePEMRSAPublicKey(block)
	}

	if publicKey, err := parseSSHRSAAuthorizedKey(raw); err == nil {
		return publicKey, nil
	}
	return nil, errors.New("public key must be an RSA public key in PEM or ssh-rsa format")
}

func parsePEMRSAPublicKey(block *pem.Block) (*rsa.PublicKey, error) {
	if block == nil {
		return nil, errors.New("public key PEM block is missing")
	}

	if parsed, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		publicKey, ok := parsed.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("public key is not RSA")
		}
		return publicKey, nil
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse RSA public key: %w", err)
	}
	return publicKey, nil
}

func parseSSHRSAAuthorizedKey(raw string) (*rsa.PublicKey, error) {
	fields := strings.Fields(raw)
	if len(fields) < 2 || fields[0] != "ssh-rsa" {
		return nil, errors.New("not an ssh-rsa public key")
	}

	blob, err := base64.StdEncoding.DecodeString(fields[1])
	if err != nil {
		return nil, fmt.Errorf("decode ssh-rsa public key: %w", err)
	}

	algorithm, rest, err := readSSHString(blob)
	if err != nil {
		return nil, err
	}
	if string(algorithm) != "ssh-rsa" {
		return nil, errors.New("ssh public key algorithm is not rsa")
	}

	exponentBytes, rest, err := readSSHString(rest)
	if err != nil {
		return nil, err
	}
	modulusBytes, rest, err := readSSHString(rest)
	if err != nil {
		return nil, err
	}
	if len(rest) != 0 {
		return nil, errors.New("ssh-rsa public key has trailing data")
	}

	exponent := new(big.Int).SetBytes(exponentBytes)
	if !exponent.IsInt64() || exponent.Sign() <= 0 {
		return nil, errors.New("ssh-rsa public key exponent is invalid")
	}
	modulus := new(big.Int).SetBytes(modulusBytes)
	if modulus.Sign() <= 0 {
		return nil, errors.New("ssh-rsa public key modulus is invalid")
	}

	return &rsa.PublicKey{N: modulus, E: int(exponent.Int64())}, nil
}

func readSSHString(raw []byte) ([]byte, []byte, error) {
	if len(raw) < 4 {
		return nil, nil, errors.New("ssh public key string is truncated")
	}
	length := int(binary.BigEndian.Uint32(raw[:4]))
	raw = raw[4:]
	if length < 0 || length > len(raw) {
		return nil, nil, errors.New("ssh public key string length is invalid")
	}
	return raw[:length], raw[length:], nil
}
