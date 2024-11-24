package blobstore_data_converter

import (
	"blob-store-data-converter/blobstore"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
)

func Test_DataConverter(t *testing.T) {
	defaultDc := converter.GetDefaultDataConverter()

	ctx := context.Background()
	ctx = context.WithValue(ctx, PropagatedValuesKey, PropagatedValues{
		TenantID:       "t1",
		BlobNamePrefix: []string{"t1", "starter"},
	})

	blobDc := NewDataConverter(
		converter.GetDefaultDataConverter(),
		blobstore.NewTestClient(),
	)
	blobDcCtx := blobDc.WithContext(ctx)

	defaultPayloads, err := defaultDc.ToPayloads("Testing")
	require.NoError(t, err)

	offloadedPayloads, err := blobDcCtx.ToPayloads("Testing")
	require.NoError(t, err)

	require.NotEqual(t, defaultPayloads.Payloads[0].GetData(), offloadedPayloads.Payloads[0].GetData())

	var result string
	err = blobDc.FromPayloads(offloadedPayloads, &result)
	require.NoError(t, err)

	require.Equal(t, "Testing", result)
}
