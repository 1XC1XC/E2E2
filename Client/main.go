package main

import (
	"E2E2/Cipher"
	"E2E2/Storage"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	PrivateKey [32]byte
	PublicKey  [32]byte
	K1         [32]byte
	K2         []byte
	K3         []byte
	SessionID  string
	ServerURL  string
	Storage    *Storage.Client
}

func NewClient(serverURL string) (*Client, error) {
	privateKey, publicKey, err := Cipher.CreateKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client keys: %w", err)
	}

	storage := Storage.Init("client_session")

	return &Client{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		ServerURL:  serverURL,
		Storage:    storage,
	}, nil
}

func (c *Client) LoadSession() bool {
	sessionID, err := c.Storage.Get("SessionID")
	if err != nil {
		return false
	}

	k3Hex, err := c.Storage.Get("K3")
	if err != nil {
		return false
	}

	k3, err := hex.DecodeString(k3Hex)
	if err != nil {
		log.Printf("Failed to decode K3: %v", err)
		return false
	}

	c.SessionID = sessionID
	c.K3 = k3
	log.Println("Session loaded successfully.")
	return true
}

func (c *Client) SaveSession() {
	c.Storage.Set("SessionID", c.SessionID)
	c.Storage.Set("K3", hex.EncodeToString(c.K3))
	c.Storage.Save()
}

func (c *Client) ExchangeKeys() error {
	clientPublicKeyHex := hex.EncodeToString(c.PublicKey[:])
	resp, err := http.PostForm(c.ServerURL+"/exchange-keys", url.Values{
		"ClientPublicKey": {clientPublicKeyHex},
	})
	if err != nil {
		return fmt.Errorf("failed to send public key: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ServerPublicKey string `json:"ServerPublicKey"`
		SessionID       string `json:"SessionID"`
		Error           string `json:"Error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse server response: %w", err)
	}

	if result.Error != "" {
		return fmt.Errorf("server returned an error: %s", result.Error)
	}

	serverPublicKey, err := hex.DecodeString(result.ServerPublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode server public key: %w", err)
	}

	c.K1, err = Cipher.CreateSharedKey(c.PrivateKey, *(*[32]byte)(serverPublicKey))
	if err != nil {
		return fmt.Errorf("failed to create K1: %w", err)
	}

	c.SessionID = result.SessionID

	c.K2, err = Cipher.DeriveK2(c.K1[:], c.SessionID)
	if err != nil {
		return fmt.Errorf("failed to derive K2: %w", err)
	}

	c.K3, err = Cipher.DeriveK3(c.K2)
	if err != nil {
		return fmt.Errorf("failed to derive K3: %w", err)
	}

	c.SaveSession()
	return nil
}

func (c *Client) GetCurrentEAPI() string {
	timestamp := time.Now().Unix() / 30 // Change every 30 seconds
	eapi := Cipher.DeriveEAPI(c.K3, timestamp)
	return hex.EncodeToString(eapi)
}

func (c *Client) Send(message string) (string, error) {
	encrypted, err := Cipher.EncryptAES([]byte(message), c.K3)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt message: %w", err)
	}

	eapi := c.GetCurrentEAPI()

	resp, err := http.PostForm(c.ServerURL+"/tunnel", url.Values{
		"Data":      {encrypted},
		"EAPI":      {eapi},
		"SessionID": {c.SessionID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to send encrypted message: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Data  string `json:"Data"`
		Error string `json:"Error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse server response: %w", err)
	}

	if response.Error != "" {
		return "", fmt.Errorf("server returned an error: %s", response.Error)
	}

	decrypted, err := Cipher.DecryptAES(response.Data, c.K3)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt server response: %w", err)
	}

	return decrypted, nil
}

func main() {
	client, err := NewClient("http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	if !client.LoadSession() {
		if err := client.ExchangeKeys(); err != nil {
			log.Fatalf("Failed to exchange keys: %v", err)
		}
	}

	response, err := client.Send("ping")
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	fmt.Printf("Server response: %s\n", response)
}
