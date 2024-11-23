package blobstore_data_converter

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

type (
	// contextKey is an unexported type used as key for items stored in the
	// Context object
	contextKey struct{}

	// propagator implements the custom context propagator
	propagator struct{}
)

// PropagatedValuesKey is the key used to store the value in the Context object
var PropagatedValuesKey = contextKey{}

type PropagatedValues struct {
	BlobStorePathSegments []string `json:"bspsegs"`
}

func NewPropagatedValues(pathSegments []string) PropagatedValues {
	return PropagatedValues{
		BlobStorePathSegments: pathSegments,
	}
}

// propagationKey is the key used by the propagator to pass values through the
// Temporal server headers
const propagationKey = "context-propagation"

// NewContextPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewContextPropagator() workflow.ContextPropagator {
	return &propagator{}
}

// Inject injects values from context into headers for propagation
func (s *propagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(PropagatedValuesKey)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *propagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	vals := ctx.Value(PropagatedValuesKey).(PropagatedValues)

	payload, err := converter.GetDefaultDataConverter().ToPayload(vals)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// Extract extracts values from headers and puts them into context
func (s *propagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if value, ok := reader.Get(propagationKey); ok {
		var data PropagatedValues
		if err := converter.GetDefaultDataConverter().FromPayload(value, &data); err != nil {
			return ctx, fmt.Errorf("failed to extract value from header: %w", err)
		}
		ctx = context.WithValue(ctx, PropagatedValuesKey, data)
	}

	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *propagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if value, ok := reader.Get(propagationKey); ok {
		var data PropagatedValues
		if err := converter.GetDefaultDataConverter().FromPayload(value, &data); err != nil {
			return ctx, fmt.Errorf("failed to extract value from header: %w", err)
		}
		ctx = workflow.WithValue(ctx, PropagatedValuesKey, data)
	}

	return ctx, nil
}
