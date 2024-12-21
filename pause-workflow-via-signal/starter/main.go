package main

import (
	"context"
	pause "github.com/kevinawoo/temporal-code-samples/pause-workflow-via-signal"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"log"
	"time"

	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:                    "pause_workflow_ID",
		TaskQueue:             "pause-workflow",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		WorkflowTaskTimeout:   time.Hour,

		// SearchAttributes are only used to make it easier to search for paused workflows.
		// We can rely solely on Signals to pause/resume a workflow
		TypedSearchAttributes: temporal.NewSearchAttributes(),
	}

	// if we want to start the workflow in paused state, then SignalWithStartWorkflow:
	//we, err := c.SignalWithStartWorkflow(
	//	context.Background(),
	//	"pause_workflow_ID",
	//	pause.PauseSignalName,
	//	nil,
	//	workflowOptions,
	//	pause.Workflow,
	//)

	// execute workflow normally in a running state
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, pause.Workflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
