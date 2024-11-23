package blobstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (b *Client) SaveBlob(ctx context.Context, key string, data []byte) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/blob-store-data-converter/blobs/%s", dir, strings.ReplaceAll(key, "/", "_"))
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

var NotFoundError = errors.New("blob not found")

func (b *Client) GetBlob(ctx context.Context, key string) ([]byte, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/blob-store-data-converter/blobs/%s", dir, strings.ReplaceAll(key, "/", "_"))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Join(NotFoundError, err)
	}

	return data, nil
}
