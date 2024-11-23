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

var _ = workflow.ContextAware(&DataConverter{}) // Ensure that DataConverter implements workflow.ContextAware

// NewDataConverter returns DataConverter, which embeds converter.DataConverter
func NewDataConverter(parent converter.DataConverter, client *blobstore.Client) *DataConverter {
	next := []converter.PayloadCodec{
		NewBaseCodec(client),
	}

	return &DataConverter{
		client:        client,
		parent:        parent,
		DataConverter: converter.NewCodecDataConverter(parent, next...),
	}
}

// WithWorkflowContext will create a CtxAwareCodec used to store payloads in blob storage
// This is called from within the workflow
func (dc *DataConverter) WithWorkflowContext(ctx workflow.Context) converter.DataConverter {
	if vals, ok := ctx.Value(PropagatedValuesKey).(PropagatedValues); ok {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithWorkflowContext(ctx)
		}

		return converter.NewCodecDataConverter(dc.parent, NewCtxAwareCodec(context.Background(), dc.client, vals))
	}

	return dc
}

// WithContext will create a CtxAwareCodec used to store payloads in blob storage
// This is called from the starter when executing and during activities starting and completing
func (dc *DataConverter) WithContext(ctx context.Context) converter.DataConverter {
	if vals, ok := ctx.Value(PropagatedValuesKey).(PropagatedValues); ok {
		parent := dc.parent
		if parentWithContext, ok := parent.(workflow.ContextAware); ok {
			parent = parentWithContext.WithContext(ctx)
		}

		return converter.NewCodecDataConverter(dc.parent, NewCtxAwareCodec(ctx, dc.client, vals))
	}

	return dc
}
