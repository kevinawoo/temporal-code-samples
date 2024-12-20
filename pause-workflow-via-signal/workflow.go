package pause_workflow_via_signal

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

// Workflow allows pausing of Activities via a signal
func Workflow(ctx workflow.Context, name string) (string, error) {
	// normally we'd use a logger so that replay doesn't spam the console
	// but in this case, we'll want to see how interceptors work during replay
	logger := workflow.GetLogger(ctx)
	_ = logger

	fmt.Println("workflow started")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 0; ; i++ {
		fmt.Println("try sending a signal to pause a workflow!")
		fmt.Println("scheduling activity", i)
		err := workflow.ExecuteActivity(ctx, Activity, i).Get(ctx, nil)
		if err != nil {
			fmt.Println("activity failed.", "Error", err)
			return "", err
		}
		fmt.Println("activity completed", i, "\n")
	}

	return "done", nil
}

func Activity(ctx context.Context, i int) (string, error) {
	fmt.Printf("running activity %d for 5s", i)
	for i := 0; i < 5; i++ {
		fmt.Printf(".")
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("\n")
	return "", nil
}
