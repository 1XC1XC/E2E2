package main

import (
	"log"

	serverpkg "E2E2/Server"
	"github.com/gin-gonic/gin"
)

func main() {
	server, err := serverpkg.NewServer()
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
