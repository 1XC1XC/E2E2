package Storage

import (
	"os"
)

func Init(Name string) *Client {
	client := &Client{
		File: Name + ".json",
		Data: make(map[interface{}]interface{}),
	}

	if _, err := os.Stat(client.File); os.IsNotExist(err) {
		client.Save()
	} else {
		client.Load()
	}

	return client
}
