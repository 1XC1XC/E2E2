package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	clientpkg "E2E2/Client"
	"github.com/gin-gonic/gin"
)

func TestSessionPersistenceAcrossRestarts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(wd)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Start first server
	srv1, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	r1 := gin.Default()
	r1.POST("/exchange-keys", srv1.HandleKeyExchange)
	r1.POST("/tunnel", srv1.HandleTunnel(func(s string) string {
		if s == "ping" {
			return "pong"
		}
		return s
	}))

	ts1 := httptest.NewServer(r1)

	client, err := clientpkg.NewClient(ts1.URL)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.ExchangeKeys(); err != nil {
		t.Fatalf("ExchangeKeys: %v", err)
	}

	resp, err := client.Send("ping")
	if err != nil || resp != "pong" {
		t.Fatalf("first ping failed: %v %s", err, resp)
	}

	ts1.Close()

	// Path to the session file for later checks
	sessFile := filepath.Join(tempDir, "server_sessions.json")

	// Restart server
	srv2, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer2: %v", err)
	}

	r2 := gin.Default()
	r2.POST("/exchange-keys", srv2.HandleKeyExchange)
	r2.POST("/tunnel", srv2.HandleTunnel(func(s string) string {
		if s == "ping" {
			return "pong"
		}
		return s
	}))

	ts2 := httptest.NewServer(r2)
	defer ts2.Close()

	client.ServerURL = ts2.URL

	resp, err = client.Send("ping")
	if err != nil || resp != "pong" {
		t.Fatalf("ping after restart failed: %v %s", err, resp)
	}

	srv2.SessionMu.RLock()
	_, exists := srv2.Sessions[client.SessionID]
	srv2.SessionMu.RUnlock()
	if !exists {
		t.Fatalf("session not loaded on restart")
	}

	_, err = os.Stat(sessFile)
	if err != nil {
		t.Fatalf("session file missing after restart: %v", err)
	}
}
