package blobstore_data_converter

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// Workflow is a standard workflow definition.
// Note that the Workflow and Activity don't need to care that
// their inputs/results are being encrypted/decrypted.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Blob Store workflow started", "name", name)

	info := map[string]string{
		"name": name,
	}

	ctxVal := ctx.Value(BlobStorePathPrefixKey)
	logger.Info("workflow ctx value:", ctxVal)

	var result string
	err := workflow.ExecuteActivity(ctx, Activity, info).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	result = "Worker: " + result
	logger.Info("Blob Store workflow completed.", "result", result)

	return result, nil
}

func Activity(ctx context.Context, info map[string]string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Blob Store Activity", "info", info)

	val := ctx.Value(BlobStorePathPrefixKey)
	logger.Info("activity ctx value:", val)

	name, ok := info["name"]
	if !ok {
		name = "someone"
	}

	return "Activity: Hello " + name + "!", nil
}
