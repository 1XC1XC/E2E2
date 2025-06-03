package playground

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/binary"
    "testing"

    cipherpkg "E2E2/Cipher"
)

func TestDeriveEAPI(t *testing.T) {
    k3 := bytes.Repeat([]byte{0x01}, 32)
    timestamp := int64(1234567890)

    // Manually compute expected HMAC
    buf := new(bytes.Buffer)
    if err := binary.Write(buf, binary.BigEndian, timestamp); err != nil {
        t.Fatalf("binary.Write failed: %v", err)
    }
    h := hmac.New(sha256.New, k3)
    h.Write(buf.Bytes())
    expected := h.Sum(nil)

    got := cipherpkg.DeriveEAPI(k3, timestamp)
    if !bytes.Equal(got, expected) {
        t.Errorf("DeriveEAPI() = %x, want %x", got, expected)
    }
}
