package blobstore_data_converter

import (
	"code-samples/blob-store-data-converter/blobstore"
	"context"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

type DataConverter struct {
	client *blobstore.Client
	parent converter.DataConverter // Until EncodingDataConverter supports workflow.ContextAware we'll store parent here.

	converter.DataConverter // embeds converter.DataConverter
}

// Ensure that DataConverter implements workflow.ContextAware
var _ = workflow.ContextAware(&DataConverter{})

// NewDataConverter returns DataConverter, which embeds converter.DataConverter
func NewDataConverter(parent converter.DataConverter, client *blobstore.Client) *DataConverter {
	next := []converter.PayloadCodec{
		NewCodec(client),
	}

	return &DataConverter{
		client:        client,
		parent:        parent,
		DataConverter: converter.NewCodecDataConverter(parent, next...),
	}
}

// WithWorkflowContext is needed to allow the blobstore be path prefixed by a tenant ID
func (dc *DataConverter) WithWorkflowContext(ctx workflow.Context) converter.DataConverter {
	if val, ok := ctx.Value(TenantKey).(string); ok {
		_ = val
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithWorkflowContext(ctx)
		}

		return NewDataConverter(parent, dc.client)
	}

	return dc
}

// WithContext is called from the starter and used to inject values to the workflow
func (dc *DataConverter) WithContext(ctx context.Context) converter.DataConverter {
	if val, ok := ctx.Value(TenantKey).(string); ok {
		_ = val
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithContext(ctx)
		}

		// use the default converter to encode/decode the context data
		return converter.GetDefaultDataConverter()
	}

	return dc
}
