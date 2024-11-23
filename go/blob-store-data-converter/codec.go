package blobstore_data_converter

import (
	"code-samples/blob-store-data-converter/blobstore"
	"context"
	"fmt"
	"github.com/google/uuid"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"strings"
)

const (
	MetadataEncodingBlobStorePlain = "blobstore/plain"
)

type BaseCodec struct {
	client *blobstore.Client
}

var _ = converter.PayloadCodec(&BaseCodec{}) // Ensure that BaseCodec implements converter.PayloadCodec

// NewBaseCodec is not aware of where to store the blobs
// Prefer to use NewCtxAwareCodec when possible
func NewBaseCodec(c *blobstore.Client) *BaseCodec {
	return &BaseCodec{
		client: c,
	}
}

// Encode needs to be implemented for codec-server support inputs from the CLI and UI
// In its current implementation, it does not how/where to store the blobs
func (c *BaseCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		result[i] = &commonpb.Payload{Metadata: p.Metadata, Data: p.Data}
		return nil, fmt.Errorf(string(p.Metadata["encoding"]) + " encoding not implemented")
	}

	return result, nil
}

// Decode does not need to be context aware because it can fetch the blobs via the payload path
func (c *BaseCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	return decode(context.Background(), c.client, payloads)
}

func decode(ctx context.Context, client *blobstore.Client, payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		if string(p.Metadata["encoding"]) != MetadataEncodingBlobStorePlain {
			result[i] = p
			continue
		}

		// fetch it from our blob store db
		data, err := client.GetBlob(ctx, string(p.Data))
		if err != nil {
			return payloads, err
		}

		result[i] = &commonpb.Payload{}
		err = result[i].Unmarshal(data)
		if err != nil {
			return payloads, err
		}
	}

	return result, nil
}

type CtxAwareCodec struct {
	ctx        context.Context // todo: it's bad practice to store context the struct, but this is a singleton, so maybe it's ok
	client     *blobstore.Client
	bucket     string
	pathPrefix []string
}

var _ = converter.PayloadCodec(&CtxAwareCodec{}) // Ensure that CtxAwareCodec implements converter.PayloadCodec

// NewCtxAwareCodec is aware of where of the propagated context values from the data converter
func NewCtxAwareCodec(ctx context.Context, c *blobstore.Client, values PropagatedValues) *CtxAwareCodec {
	return &CtxAwareCodec{
		ctx:        ctx,
		client:     c,
		bucket:     "blob://mybucket",
		pathPrefix: values.BlobStorePathSegments,
	}
}

// Encode knows where to store the blobs from values stored in the context
func (c *CtxAwareCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		origBytes, err := p.Marshal()
		if err != nil {
			return payloads, err
		}

		// save the data in our blob store db
		objectName := strings.Join(c.pathPrefix[1:], "-") + "_" + uuid.New().String() // ensures each blob is unique
		path := fmt.Sprintf("%s/%s/%s", c.bucket, c.pathPrefix[0], objectName)
		err = c.client.SaveBlob(c.ctx, path, origBytes)
		if err != nil {
			return payloads, err
		}

		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				"encoding": []byte(MetadataEncodingBlobStorePlain),
			},
			Data: []byte(path),
		}
	}

	return result, nil
}

// Decode does not need to be context aware because it can fetch the blobs via the payload path
func (c *CtxAwareCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	return decode(c.ctx, c.client, payloads)
}
