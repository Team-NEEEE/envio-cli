package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"math/big"
)

// GenerateRSAKeyPairPEM generates a 2048-bit RSA key pair and returns both keys as PEM strings.
func GenerateRSAKeyPairPEM() (privatePEM string, publicPEM string, err error) {
	privateKey, err := generateRSAKey()
	if err != nil {
		return "", "", err
	}

	publicPEM, err = encodeRSAPublicKeyPEM(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	return encodeRSAPrivateKeyPEM(privateKey), publicPEM, nil
}

// GenerateRSAKeyPairForLogin returns a private PEM and an SSH authorized_keys public key.
func GenerateRSAKeyPairForLogin() (privatePEM string, publicKey string, err error) {
	privateKey, err := generateRSAKey()
	if err != nil {
		return "", "", err
	}

	publicKey, err = encodeRSAPublicKeySSH(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	return encodeRSAPrivateKeyPEM(privateKey), publicKey, nil
}

func generateRSAKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("RSA private key 생성 실패: %w", err)
	}
	return privateKey, nil
}

func encodeRSAPrivateKeyPEM(privateKey *rsa.PrivateKey) string {
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateDER,
	}
	return string(pem.EncodeToMemory(privateBlock))
}

func encodeRSAPublicKeyPEM(publicKey *rsa.PublicKey) (string, error) {
	publicDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("RSA public key 직렬화 실패: %w", err)
	}

	publicBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	}

	return string(pem.EncodeToMemory(publicBlock)), nil
}

func encodeRSAPublicKeySSH(publicKey *rsa.PublicKey) (string, error) {
	if publicKey == nil {
		return "", fmt.Errorf("RSA public key is nil")
	}

	blob := make([]byte, 0, 3+4+3+4+256)
	blob = appendSSHString(blob, []byte("ssh-rsa"))
	blob = appendSSHString(blob, marshalSSHMPInt(big.NewInt(int64(publicKey.E))))
	blob = appendSSHString(blob, marshalSSHMPInt(publicKey.N))

	return "ssh-rsa " + base64.StdEncoding.EncodeToString(blob), nil
}

func appendSSHString(dst []byte, value []byte) []byte {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	dst = append(dst, length[:]...)
	return append(dst, value...)
}

func marshalSSHMPInt(value *big.Int) []byte {
	if value == nil || value.Sign() == 0 {
		return nil
	}

	encoded := value.Bytes()
	if encoded[0]&0x80 != 0 {
		encoded = append([]byte{0}, encoded...)
	}
	return encoded
}
