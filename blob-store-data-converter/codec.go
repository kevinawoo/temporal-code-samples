package blobstore_data_converter

import (
	"blob-store-data-converter/blobstore"
	"context"
	"fmt"
	"github.com/google/uuid"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"strings"
)

const (
	MetadataEncodingBlobStorePlain = "blobstore/plain"

	// gRPC has a 4MB limit.
	// To save some space for other metadata, we should maybe 80% or half that.
	//
	// For this example however, we'll use much smaller limit as a proof of concept.
	payloadSizeLimit = 33
)

// BaseCodec is not aware of where to store the blobs, however fetching blobs can be done without
// knowing where to store the blog because we're passed the fully qualified path in the payload
//
// Prefer the use of CtxAwareCodec when possible.
type BaseCodec struct {
	client *blobstore.Client
}

var _ = converter.PayloadCodec(&BaseCodec{}) // Ensure that BaseCodec implements converter.PayloadCodec

func NewBaseCodec(c *blobstore.Client) *BaseCodec {
	return &BaseCodec{
		client: c,
	}
}

// Encode In its current implementation is just a pass-through to allow the UI/CLI
// to send regular json payloads.
func (c *BaseCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		result[i] = &commonpb.Payload{Metadata: p.Metadata, Data: p.Data}
	}

	return result, nil
}

// Decode does not need to be context aware because it can fetch the blobs via a fully qualified object path
func (c *BaseCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	return decode(context.Background(), c.client, payloads)
}

func decode(ctx context.Context, client *blobstore.Client, payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		if string(p.Metadata["encoding"]) != MetadataEncodingBlobStorePlain {
			result[i] = &commonpb.Payload{Metadata: p.Metadata, Data: p.Data}
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

// CtxAwareCodec knows where to store the blobs from the PropagatedValues
// Note, see readme for details on missing values
type CtxAwareCodec struct {
	ctx        context.Context // todo: it's bad practice to store context the struct, but this is a singleton, so maybe it's ok
	client     *blobstore.Client
	bucket     string
	tenant     string
	pathPrefix []string
}

var _ = converter.PayloadCodec(&CtxAwareCodec{}) // Ensure that CtxAwareCodec implements converter.PayloadCodec

// NewCtxAwareCodec is aware of where of the propagated context values from the data converter
func NewCtxAwareCodec(ctx context.Context, c *blobstore.Client, values PropagatedValues) *CtxAwareCodec {
	return &CtxAwareCodec{
		ctx:        ctx,
		client:     c,
		bucket:     "blob://mybucket",
		tenant:     values.TenantID,
		pathPrefix: values.BlobNamePrefix,
	}
}

// Encode knows where to store the blobs from values stored in the context
func (c *CtxAwareCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		// if the payload is small enough, just send it as is
		fmt.Printf("payload len(%s): %d\n", string(p.Data), len(p.Data))
		if len(p.Data) < payloadSizeLimit {
			result[i] = &commonpb.Payload{Metadata: p.Metadata, Data: p.Data}
			continue
		}

		origBytes, err := p.Marshal()
		if err != nil {
			return payloads, err
		}

		// save the data in our blob store db
		objectName := strings.Join(c.pathPrefix, "_") + "__" + uuid.New().String() // ensures each blob is unique
		path := fmt.Sprintf("%s/%s/%s", c.bucket, c.tenant, objectName)
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
