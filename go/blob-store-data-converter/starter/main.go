package main

import (
	bsdc "code-samples/blob-store-data-converter"
	"code-samples/blob-store-data-converter/blobstore"
	"context"
	"go.temporal.io/sdk/workflow"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

func main() {
	bsClient := blobstore.NewClient()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		DataConverter: bsdc.NewDataConverter(
			converter.GetDefaultDataConverter(),
			bsClient,
		),
		// Use a ContextPropagator so that the KeyID value set in the workflow context is
		// also availble in the context for activities.
		ContextPropagators: []workflow.ContextPropagator{
			bsdc.NewContextPropagator(),
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "blobstore_codec_workflow",
		TaskQueue: "blobstore_codec",
	}

	ctx := context.Background()
	// If you are using a ContextPropagator and varying keys per workflow you need to set
	// the KeyID to use for this workflow in the context:
	ctx = context.WithValue(ctx, bsdc.BlobStorePathPrefixKey, []string{"tenant111"})

	we, err := c.ExecuteWorkflow(
		ctx,
		workflowOptions,
		bsdc.Workflow,
		"Starter: Big Blob Bob",
	)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
