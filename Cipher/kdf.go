package Cipher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"golang.org/x/crypto/hkdf"
	"io"
)

func DeriveKey(secret, salt, info []byte, keyLength int) ([]byte, error) {
	hkdf := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, keyLength)
	_, err := io.ReadFull(hkdf, key)
	return key, err
}

func DeriveK2(k1 []byte, apiKey string) ([]byte, error) {
	return DeriveKey(k1, []byte(apiKey), []byte("K2"), 32)
}

func DeriveK3(k2 []byte) ([]byte, error) {
	return DeriveKey(k2, nil, []byte("K3"), 32)
}

func DeriveEAPI(k3 []byte, timestamp int64) []byte {
	h := hmac.New(sha256.New, k3)
	// Writing to an in-memory hash should never fail but we capture the
	// error for completeness.
	if err := binary.Write(h, binary.BigEndian, timestamp); err != nil {
		// panic is acceptable here since failure indicates a bug in the
		// underlying implementation and cannot be recovered from.
		panic(err)
	}
	return h.Sum(nil)
}
