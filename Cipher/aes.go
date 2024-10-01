package Cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

func EncryptAES(PlainText, Key []byte) (string, error) {
	Block, err := aes.NewCipher(Key)
	if err != nil {
		return "", fmt.Errorf("error creating AES cipher: %w", err)
	}

	AES_GCM, err := cipher.NewGCM(Block)
	if err != nil {
		return "", fmt.Errorf("error creating GCM: %w", err)
	}

	Nonce := make([]byte, AES_GCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, Nonce); err != nil {
		return "", fmt.Errorf("error generating nonce: %w", err)
	}

	CipherText := AES_GCM.Seal(Nonce, Nonce, PlainText, nil)

	return hex.EncodeToString(CipherText), nil
}

func DecryptAES(HexCipherText string, Key []byte) (string, error) {

	CipherText, err := hex.DecodeString(HexCipherText)
	if err != nil {
		return "", fmt.Errorf("error decoding ciphertext: %w", err)
	}

	Block, err := aes.NewCipher(Key)
	if err != nil {
		return "", fmt.Errorf("error creating AES cipher: %w", err)
	}

	AES_GCM, err := cipher.NewGCM(Block)
	if err != nil {
		return "", fmt.Errorf("error creating GCM: %w", err)
	}

	NonceSize := AES_GCM.NonceSize()

	if len(CipherText) < NonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	Nonce, CipherText := CipherText[:NonceSize], CipherText[NonceSize:]

	PlainText, err := AES_GCM.Open(nil, Nonce, CipherText, nil)
	if err != nil {
		return "", fmt.Errorf("error decrypting message: %w", err)
	}

	return string(PlainText), nil
}
