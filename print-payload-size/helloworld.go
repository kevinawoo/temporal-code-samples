package print_payload_size

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

// Workflow is a Hello World workflow definition.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	var result string
	name = RandStringBytes(1_000_000 * 1)

	// Notice that if we call these calls are async, the SDK will batch up commands together
	// see the log message: "RespondWorkflowTaskCompleted payload ..."
	//
	// To break them apart, you'll need to issue a synchronous command like .Get()
	workflow.ExecuteActivity(ctx, Activity, name)
	workflow.ExecuteActivity(ctx, Activity, name)
	err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("HelloWorld workflow completed.", "result", result)

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	return RandStringBytes(1_000_000 * 2), nil
}
