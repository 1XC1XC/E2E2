package Cipher

import "testing"

func TestEncryptDecryptAES(t *testing.T) {
    key := make([]byte, 32)
    for i := 0; i < len(key); i++ {
        key[i] = byte(i)
    }
    plain := []byte("hello world")

    enc, err := EncryptAES(plain, key)
    if err != nil {
        t.Fatalf("EncryptAES error: %v", err)
    }

    dec, err := DecryptAES(enc, key)
    if err != nil {
        t.Fatalf("DecryptAES error: %v", err)
    }

    if dec != string(plain) {
        t.Fatalf("expected %q got %q", string(plain), dec)
    }
}
