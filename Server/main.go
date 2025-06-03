package main

import (
        "E2E2/Cipher"
        "E2E2/Storage"
        "encoding/hex"
        "fmt"
        "log"
        "net/http"
        "sync"
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
        SessionMu  sync.RWMutex
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
        s.SessionMu.Lock()
        session, exists := s.Sessions[sessionID]
        if !exists {
                s.SessionMu.Unlock()
                return false
        }

        currentTime := time.Now()
        if currentTime.After(session.ExpiresAt) {
                delete(s.Sessions, sessionID)
                s.SessionMu.Unlock()
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
        s.SessionMu.Unlock()
        return true
}

func (s *Server) HandleKeyExchange(c *gin.Context) {
	clientPublicKeyHex := c.PostForm("ClientPublicKey")
	clientPublicKeyBytes, err := hex.DecodeString(clientPublicKeyHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ClientPublicKey"})
		return
	}

	if len(clientPublicKeyBytes) != 32 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid ClientPublicKey length"})
		return
	}

	var clientPublicKey [32]byte
	copy(clientPublicKey[:], clientPublicKeyBytes)

	k1, err := Cipher.CreateSharedKey(s.PrivateKey, clientPublicKey)
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

        s.SessionMu.Lock()
        s.Sessions[sessionID] = &ServerSession{
                K3:        k3,
                LastUsed:  time.Now(),
                ExpiresAt: time.Now().Add(24 * time.Hour), // Session expires after 24 hours
        }
        s.SessionMu.Unlock()

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

                s.SessionMu.RLock()
                session := s.Sessions[sessionID]
                s.SessionMu.RUnlock()
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
