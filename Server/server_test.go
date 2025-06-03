package server

import (
	"net/http/httptest"
	"os"
	"testing"

	clientpkg "E2E2/Client"
	"github.com/gin-gonic/gin"
)

func startTestServer(t *testing.T) *httptest.Server {
	srv, err := NewServer()
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	router := gin.Default()
	router.POST("/exchange-keys", srv.HandleKeyExchange)
	router.POST("/tunnel", srv.HandleTunnel(func(msg string) string {
		if msg == "ping" {
			return "pong"
		}
		return "unknown message"
	}))
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	return ts
}

func TestServerClientCommunication(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(cwd) })

	ts := startTestServer(t)

	c, err := clientpkg.NewClient(ts.URL)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if err := c.ExchangeKeys(); err != nil {
		t.Fatalf("ExchangeKeys: %v", err)
	}

	resp, err := c.Send("ping")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if resp != "pong" {
		t.Fatalf("expected pong got %s", resp)
	}
}
