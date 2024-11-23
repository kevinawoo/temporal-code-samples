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

var _ = converter.PayloadCodec(&BaseCodec{})

func NewBaseCodec(c *blobstore.Client) *BaseCodec {
	return &BaseCodec{
		client: c,
	}
}

type ScopedCodec struct {
	ctx        context.Context // todo: remove this hack
	client     *blobstore.Client
	bucket     string
	pathPrefix []string
}

var _ = converter.PayloadCodec(&ScopedCodec{})

func NewScopedCodec(ctx context.Context, c *blobstore.Client, pathPrefix []string) *ScopedCodec {
	return &ScopedCodec{
		ctx:        ctx,
		client:     c,
		bucket:     "blob://mybucket",
		pathPrefix: pathPrefix,
	}
}

// Encode
//
//	payloads looks like:
//	[
//		{metadata: {encoding: "json/plain", tenantId: "..."}, data: ...},
//		{metadata: {encoding: "json/plain", tenantId: "..."}, data: ...},
//	]
func (c *ScopedCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		origBytes, err := p.Marshal()
		if err != nil {
			return payloads, err
		}

		// save the data in our blob store db
		// path = blob://<tenantId>/<workflowId>-<runId>-<eventId>
		objectName := strings.Join(c.pathPrefix[1:], "-")
		if objectName == "" {
			objectName = "unknown-" + uuid.New().String()
		}
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

func (c *ScopedCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		if string(p.Metadata["encoding"]) != "blobstore/plain" {
			continue
		}

		// fetch it from our blob store db
		data, err := c.client.GetBlob(c.ctx, string(p.Data))
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

func (c *BaseCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	panic("unimplemented")
}

// Decode gets called from the starter on workflow completion
func (c *BaseCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		if string(p.Metadata["encoding"]) != MetadataEncodingBlobStorePlain {
			result[i] = p
			continue
		}

		// fetch it from our blob store db
		data, err := c.client.GetBlob(context.TODO(), string(p.Data))
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
