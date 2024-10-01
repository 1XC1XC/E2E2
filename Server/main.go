package main

import (
	"E2E2/Cipher"
	"E2E2/Storage"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ServerSession struct {
	K3        []byte
	LastUsed  time.Time
	ExpiresAt time.Time
}

type Server struct {
	PrivateKey [32]byte
	PublicKey  [32]byte
	Storage    *Storage.Client
	Sessions   map[string]*ServerSession
}

func NewServer() (*Server, error) {
	privateKey, publicKey, err := Cipher.CreateKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate server keys: %w", err)
	}

	storage := Storage.Init("server_sessions")

	return &Server{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Storage:    storage,
		Sessions:   make(map[string]*ServerSession),
	}, nil
}

func (s *Server) ValidateEAPI(sessionID string, receivedEAPI string) bool {
	session, exists := s.Sessions[sessionID]
	if !exists {
		return false
	}

	currentTime := time.Now()
	if currentTime.After(session.ExpiresAt) {
		delete(s.Sessions, sessionID)
		return false
	}

	timestamp := currentTime.Unix() / 30
	expectedEAPI := hex.EncodeToString(Cipher.DeriveEAPI(session.K3, timestamp))

	if receivedEAPI != expectedEAPI {
		// Check the previous interval as well, in case of slight time differences
		previousEAPI := hex.EncodeToString(Cipher.DeriveEAPI(session.K3, timestamp-1))
		if receivedEAPI != previousEAPI {
			return false
		}
	}

	session.LastUsed = currentTime
	return true
}

func (s *Server) HandleKeyExchange(c *gin.Context) {
	clientPublicKeyHex := c.PostForm("ClientPublicKey")
	clientPublicKey, err := hex.DecodeString(clientPublicKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ClientPublicKey"})
		return
	}

	k1, err := Cipher.CreateSharedKey(s.PrivateKey, *(*[32]byte)(clientPublicKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to create K1"})
		return
	}

	sessionID := uuid.New().String()

	k2, err := Cipher.DeriveK2(k1[:], sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to derive K2"})
		return
	}

	k3, err := Cipher.DeriveK3(k2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to derive K3"})
		return
	}

	s.Sessions[sessionID] = &ServerSession{
		K3:        k3,
		LastUsed:  time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // Session expires after 24 hours
	}

	c.JSON(http.StatusOK, gin.H{
		"ServerPublicKey": hex.EncodeToString(s.PublicKey[:]),
		"SessionID":       sessionID,
	})
}

func (s *Server) HandleTunnel(callback func(string) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.PostForm("SessionID")
		eapi := c.PostForm("EAPI")

		if !s.ValidateEAPI(sessionID, eapi) {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "Invalid or expired EAPI"})
			return
		}

		session := s.Sessions[sessionID]
		encryptedData := c.PostForm("Data")
		decrypted, err := Cipher.DecryptAES(encryptedData, session.K3)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Failed to decrypt"})
			return
		}

		response := callback(decrypted)

		encrypted, err := Cipher.EncryptAES([]byte(response), session.K3)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to encrypt response"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"Data": encrypted})
	}
}

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	router := gin.Default()
	router.POST("/exchange-keys", server.HandleKeyExchange)

	router.POST("/tunnel", server.HandleTunnel(func(response string) string {
		if response == "ping" {
			return "pong"
		}
		return "unknown message"
	}))

	log.Println("Server starting on :8080")
	log.Fatal(router.Run(":8080"))
}
