package Cipher

import (
	"crypto/rand"
	"golang.org/x/crypto/curve25519"
)

func CreateKeys() ([32]byte, [32]byte, error) {
	PrivateKey, err := CreatePrivateKey()
	if err != nil {
		return [32]byte{}, [32]byte{}, err
	}

	PublicKey, err := CreatePublicKey(PrivateKey)
	if err != nil {
		return [32]byte{}, [32]byte{}, err
	}

	return PrivateKey, PublicKey, nil
}

func CreatePrivateKey() ([32]byte, error) {
	var ClientPrivateKey [32]byte
	_, err := rand.Read(ClientPrivateKey[:])
	if err != nil {
		return [32]byte{}, err
	}

	return ClientPrivateKey, nil
}

func CreatePublicKey(ClientPrivateKey [32]byte) ([32]byte, error) {
	var ClientPublicKey [32]byte
	curve25519.ScalarBaseMult(&ClientPublicKey, &ClientPrivateKey)

	return ClientPublicKey, nil
}

func CreateSharedKey(PrivateKey, PublicKey [32]byte) ([32]byte, error) {
	var SharedKey [32]byte
	curve25519.ScalarMult(&SharedKey, &PrivateKey, &PublicKey)
	return SharedKey, nil
}
