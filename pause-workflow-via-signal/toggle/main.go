package main

import (
	"context"
	pause "github.com/kevinawoo/temporal-code-samples/pause-workflow-via-signal"
	"go.temporal.io/sdk/client"
	"log"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	d, err := c.DescribeWorkflowExecution(context.Background(), "hello_world_workflowID", "")
	if err != nil {
		log.Fatalln("Unable to describe workflow", err)
	}

	var paused bool
	data := d.WorkflowExecutionInfo.SearchAttributes.GetIndexedFields()[pause.SearchAttributeName].GetData()
	if string(data) == "true" {
		paused = true
	}
	log.Println("workflow pause state:", paused)

	signal := pause.PauseSignalName
	if paused {
		signal = pause.ResumeSignalName
	}
	err = c.SignalWorkflow(context.Background(), "hello_world_workflowID", "", signal, nil)
	if err != nil {
		log.Fatalln("unable to signal workflow", err)
	}

	log.Println("workflow now:", !paused)
}
