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
		ID:                    "hello_world_workflowID",
		TaskQueue:             "hello-world",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		WorkflowTaskTimeout:   time.Hour,
		TypedSearchAttributes: temporal.NewSearchAttributes(
			pause.PauseSearchAttrKey.ValueSet(true), // start paused
		),
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, pause.Workflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
