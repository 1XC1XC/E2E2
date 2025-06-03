package main

import (
	"fmt"
	"log"

	clientpkg "E2E2/Client"
)

func main() {
	client, err := clientpkg.NewClient("http://localhost:8080")
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
