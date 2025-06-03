package Storage

import (
	"fmt"
	"os"
)

type Client struct {
	Data map[interface{}]interface{}
	File string
}

func (c *Client) Exists(key interface{}) bool {
	_, exists := c.Data[key]
	return exists
}

func (c *Client) GetCast(key interface{}) (interface{}, error) {
	if c.Data == nil {
		return "", os.ErrInvalid
	}

	value, exists := c.Data[key]
	if !exists {
		return "", os.ErrNotExist
	}

	return value, nil
}

func (c *Client) Get(key interface{}) (string, error) {
	value, err := c.GetCast(key)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", value), nil
}

func (c *Client) Set(key, value interface{}) error {
	if c.Data == nil {
		return os.ErrInvalid
	}
	c.Data[key] = value
	return nil
}

func (c *Client) Save() error {
	data, err := JSON.Encode(c.Data)
	if err != nil {
		return err
	}
	return os.WriteFile(c.File, []byte(data), 0644)
}

func (c *Client) Load() error {
	Bytes, err := os.ReadFile(c.File)
	if err != nil {
		return err
	}

	Data, err := JSON.Decode(string(Bytes))
	if err != nil {
		return err
	}

	c.Data = Data
	return nil
}
