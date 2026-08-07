package subagenttoken

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

var ErrInvalidToken = errors.New("invalid token")

var fixedKey []byte

func init() {
	decoded, err := hex.DecodeString("4b1e9f2a7c3d8e05f61a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7")
	if err != nil {
		panic(fmt.Sprintf("subagenttoken: failed to decode fixed key: %v", err))
	}
	fixedKey = decoded
}

func newAEAD() (cipher.AEAD, error) {
	block, err := aes.NewCipher(fixedKey)
	if err != nil {
		return nil, fmt.Errorf("subagenttoken: failed to create AES cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("subagenttoken: failed to create GCM: %w", err)
	}
	return aead, nil
}

func SubagentTokenGenerate(logicalName string) (string, error) {
	aead, err := newAEAD()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("subagenttoken: failed to generate nonce: %w", err)
	}

	sealed := aead.Seal(nil, nonce, []byte(logicalName), nil)

	combined := append(nonce, sealed...)
	return base64.RawURLEncoding.EncodeToString(combined), nil
}

func SubagentTokenValidate(token string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", ErrInvalidToken
	}

	aead, err := newAEAD()
	if err != nil {
		return "", err
	}

	if len(data) < aead.NonceSize() {
		return "", ErrInvalidToken
	}

	nonce := data[:aead.NonceSize()]
	sealed := data[aead.NonceSize():]

	plaintext, err := aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", ErrInvalidToken
	}

	return string(plaintext), nil
}
