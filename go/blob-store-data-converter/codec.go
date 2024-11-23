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

type Codec struct {
	client *blobstore.Client
}

var _ = converter.PayloadCodec(&Codec{})

func NewBaseCodec(c *blobstore.Client) *Codec {
	return &Codec{
		client: c,
	}
}

type BlobStoreCodec struct {
	ctx        context.Context // todo: remove this hack
	client     *blobstore.Client
	pathPrefix []string
}

var _ = converter.PayloadCodec(&BlobStoreCodec{})

func NewBlobStoreCodec(ctx context.Context, c *blobstore.Client, pathPrefix []string) *BlobStoreCodec {
	return &BlobStoreCodec{
		ctx:        ctx,
		client:     c,
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
func (c *BlobStoreCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	bucket := "blob://mybucket"

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
		path := fmt.Sprintf("%s/%s/%s", bucket, c.pathPrefix[0], objectName)
		err = c.client.SaveBlob(c.ctx, path, origBytes)
		if err != nil {
			return payloads, err
		}

		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				"encoding": []byte("blobstore/plain"),
			},
			Data: []byte(path),
		}
	}

	return result, nil
}

func (c *BlobStoreCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
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

func (c *Codec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	panic("unimplemented")
}

// Decode gets called from the starter on workflow completion
func (c *Codec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		if string(p.Metadata["encoding"]) != "blobstore/plain" {
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
