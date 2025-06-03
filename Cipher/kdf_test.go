package Cipher

import "testing"

func TestDeriveK2K3(t *testing.T) {
    k1 := make([]byte, 32)
    for i := 0; i < 32; i++ {
        k1[i] = byte(i + 1)
    }

    k2, err := DeriveK2(k1, "test-session")
    if err != nil {
        t.Fatalf("DeriveK2 error: %v", err)
    }
    if len(k2) != 32 {
        t.Fatalf("expected 32 bytes for k2 got %d", len(k2))
    }

    k3, err := DeriveK3(k2)
    if err != nil {
        t.Fatalf("DeriveK3 error: %v", err)
    }
    if len(k3) != 32 {
        t.Fatalf("expected 32 bytes for k3 got %d", len(k3))
    }
}
