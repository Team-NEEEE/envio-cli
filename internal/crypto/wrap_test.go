package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
)

func TestGenerateProjectMasterKeyReturns32Bytes(t *testing.T) {
	t.Parallel()

	key, err := GenerateProjectMasterKey()
	if err != nil {
		t.Fatalf("GenerateProjectMasterKey() error = %v", err)
	}
	if len(key) != projectMasterKeySize {
		t.Fatalf("len(key) = %d, want %d", len(key), projectMasterKeySize)
	}
}

func TestWrapProjectMasterKeyWithPEMPublicKey(t *testing.T) {
	t.Parallel()

	privatePEM, publicPEM, err := GenerateRSAKeyPairPEM()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairPEM() error = %v", err)
	}
	privateKey := mustParseRSAPrivateKey(t, privatePEM)
	projectMasterKey := []byte("12345678901234567890123456789012")

	wrapped, err := WrapProjectMasterKey(publicPEM, projectMasterKey)
	if err != nil {
		t.Fatalf("WrapProjectMasterKey() error = %v", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(wrapped)
	if err != nil {
		t.Fatalf("wrapped key is not base64: %v", err)
	}
	got, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		t.Fatalf("DecryptOAEP() error = %v", err)
	}
	if string(got) != string(projectMasterKey) {
		t.Fatalf("decrypted master key = %q", string(got))
	}
}

func TestWrapProjectMasterKeyWithSSHPublicKey(t *testing.T) {
	t.Parallel()

	privatePEM, publicKey, err := GenerateRSAKeyPairForLogin()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairForLogin() error = %v", err)
	}
	privateKey := mustParseRSAPrivateKey(t, privatePEM)
	projectMasterKey := []byte("12345678901234567890123456789012")

	wrapped, err := WrapProjectMasterKey(publicKey, projectMasterKey)
	if err != nil {
		t.Fatalf("WrapProjectMasterKey() error = %v", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(wrapped)
	if err != nil {
		t.Fatalf("wrapped key is not base64: %v", err)
	}
	got, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		t.Fatalf("DecryptOAEP() error = %v", err)
	}
	if string(got) != string(projectMasterKey) {
		t.Fatalf("decrypted master key = %q", string(got))
	}
}

func TestUnwrapProjectMasterKey(t *testing.T) {
	t.Parallel()

	privatePEM, publicKey, err := GenerateRSAKeyPairForLogin()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairForLogin() error = %v", err)
	}
	projectMasterKey := []byte("12345678901234567890123456789012")
	wrapped, err := WrapProjectMasterKey(publicKey, projectMasterKey)
	if err != nil {
		t.Fatalf("WrapProjectMasterKey() error = %v", err)
	}

	got, err := UnwrapProjectMasterKey(privatePEM, wrapped)
	if err != nil {
		t.Fatalf("UnwrapProjectMasterKey() error = %v", err)
	}
	if string(got) != string(projectMasterKey) {
		t.Fatalf("UnwrapProjectMasterKey() = %q", string(got))
	}
}

func TestWrapProjectMasterKeyRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	privatePEM, publicPEM, err := GenerateRSAKeyPairPEM()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPairPEM() error = %v", err)
	}
	_ = privatePEM

	ecPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey() error = %v", err)
	}
	ecDER, err := x509.MarshalPKIXPublicKey(&ecPrivateKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey() error = %v", err)
	}
	ecPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: ecDER}))

	tests := []struct {
		name      string
		publicKey string
		masterKey []byte
	}{
		{
			name:      "empty public key",
			publicKey: "",
			masterKey: []byte("12345678901234567890123456789012"),
		},
		{
			name:      "invalid PEM",
			publicKey: "-----BEGIN PUBLIC KEY-----\nnot-base64\n-----END PUBLIC KEY-----",
			masterKey: []byte("12345678901234567890123456789012"),
		},
		{
			name:      "non RSA public key",
			publicKey: ecPEM,
			masterKey: []byte("12345678901234567890123456789012"),
		},
		{
			name:      "wrong master key size",
			publicKey: publicPEM,
			masterKey: []byte("short"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := WrapProjectMasterKey(tt.publicKey, tt.masterKey); err == nil {
				t.Fatal("WrapProjectMasterKey() error = nil, want error")
			}
		})
	}
}

func mustParseRSAPrivateKey(t *testing.T, privatePEM string) *rsa.PrivateKey {
	t.Helper()

	block, _ := pem.Decode([]byte(privatePEM))
	if block == nil {
		t.Fatal("private key PEM block missing")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("ParsePKCS1PrivateKey() error = %v", err)
	}
	return key
}
