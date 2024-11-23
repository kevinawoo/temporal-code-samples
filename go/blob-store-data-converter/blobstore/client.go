package blobstore

import (
	"context"
	"errors"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (b *Client) SaveBlob(ctx context.Context, key string, data []byte) error {
	db[key] = data
	return nil
}

var NotFoundError = errors.New("blob not found")

func (b *Client) GetBlob(ctx context.Context, key string) ([]byte, error) {
	if data, ok := db[key]; ok {
		return data, nil
	}

	return nil, NotFoundError
}
