package blobstore_data_converter

import (
	"context"

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

// TenantKey is the key used to store the value in the Context object
var TenantKey = contextKey{}

// propagationKey is the key used by the propagator to pass values through the
// Temporal server headers
const propagationKey = "encryption"

// NewContextPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewContextPropagator() workflow.ContextPropagator {
	return &propagator{}
}

// Inject injects values from context into headers for propagation
// This is called inside the starter
func (s *propagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(TenantKey)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
// This is called inside the worker
func (s *propagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(TenantKey)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// Extract extracts values from headers and puts them into context
// This is called inside the worker
func (s *propagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if value, ok := reader.Get(propagationKey); ok {
		var cryptContext string
		if err := converter.GetDefaultDataConverter().FromPayload(value, &cryptContext); err != nil {
			return ctx, nil
		}
		ctx = context.WithValue(ctx, TenantKey, cryptContext)
	}

	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
// This is called inside the worker
func (s *propagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if value, ok := reader.Get(propagationKey); ok {
		var cryptContext string
		if err := converter.GetDefaultDataConverter().FromPayload(value, &cryptContext); err != nil {
			return ctx, nil
		}
		ctx = workflow.WithValue(ctx, TenantKey, cryptContext)
	}

	return ctx, nil
}
