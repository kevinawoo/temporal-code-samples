package blobstore_data_converter

import (
	"code-samples/blob-store-data-converter/blobstore"
	"context"
	"fmt"
	commonpb "go.temporal.io/api/common/v1"
)

type Codec struct {
	client *blobstore.Client
}

func NewCodec(c *blobstore.Client) *Codec {
	return &Codec{
		client: c,
	}
}

// Encode
//
//	payloads looks like:
//	[
//		{metadata: {encoding: "plainText", tenantId: "..."}, data: ...},
//		{metadata: {encoding: "plainText", tenantId: "..."}, data: ...},
//	]
func (b *Codec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	ctx := context.Background()

	pathPrefix := "blob://"

	result := make([]*commonpb.Payload, len(payloads))
	for _, p := range payloads {
		// metadata = {encoding: "plainText", tenantId: "...", workflowId: "..."}

		origBytes, err := p.Marshal()
		if err != nil {
			return payloads, err
		}

		// path = blob://<tenantId>/<workflowId>-<runId>-<eventId>
		// save the data in our blob store db
		path := fmt.Sprintf("%s/%s/%s-%s-%s", pathPrefix, string(p.Metadata["tenantId"]), string(p.Metadata["workflowId"]), string(p.Metadata["runId"]), string(p.Metadata["eventId"]))
		err = b.client.SaveBlob(ctx, path, origBytes)
		if err != nil {
			return payloads, err
		}

		result = append(result, &commonpb.Payload{
			Metadata: map[string][]byte{
				"encoding": []byte("blobStore"),
			},
			Data: []byte(path),
		})
	}

	return result, nil
}

func (b *Codec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	ctx := context.Background()

	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		if string(p.Metadata["encoding"]) != "blobStore" {
			return payloads, fmt.Errorf("unhandled encoding: %s", p.Metadata["encoding"])
		}

		// fetch it from our blob store db
		data, err := b.client.GetBlob(ctx, string(p.Data))
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
