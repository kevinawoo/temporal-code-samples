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

	// Notice that if we call these ExecuteActivity calls are asynchronous, the SDK will batch them together to the Server
	// In this case, the gRPC interceptor will print out
	//		RespondWorkflowTaskCompleted payload size 3.0 MB len(commands) 3
	//
	// The `.Get()` command is a synchronous, which the SDK will wait for a result
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
