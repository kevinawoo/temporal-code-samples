package blobstore_data_converter

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// Workflow is a standard workflow definition.
// Note that the Workflow and Activity doesn't need to care that
// their inputs/results are being stored in a blog store and not on the workflow history.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("workflow started", "name", name)

	ctxVal, ok := ctx.Value(PropagatedValuesKey).(PropagatedValues)
	if !ok {
		msg := "failed to find our propagated values in the context"
		logger.Error(msg)
		return "", fmt.Errorf(msg)
	}
	fmt.Printf("workflow ctx value: %+v\n", ctxVal)

	info := map[string]string{
		"name": name,
	}

	wfInfo := workflow.GetInfo(ctx)
	ctxVal.BlobStorePathSegments = []string{ctxVal.TenantId, wfInfo.WorkflowType.Name, wfInfo.WorkflowExecution.ID}
	ctx = workflow.WithValue(ctx, PropagatedValuesKey, ctxVal)

	var result string
	err := workflow.ExecuteActivity(ctx, Activity, info).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}
	err = workflow.ExecuteActivity(ctx, Activity, info).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	result = "Workflow: " + result
	fmt.Println("workflow completed.", "result", result)

	return result, nil
}

func Activity(ctx context.Context, info map[string]string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "info", info)

	val := ctx.Value(PropagatedValuesKey)
	fmt.Printf("Activity ctx value: %+v\n", val)

	name, ok := info["name"]
	if !ok {
		name = "someone"
	}

	return "Activity: " + name + "!", nil
}
