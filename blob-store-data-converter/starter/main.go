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
	ctx := context.Background()

	bsClient := blobstore.NewClient()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		DataConverter: bsdc.NewDataConverter(
			converter.GetDefaultDataConverter(),
			bsClient,
		),
		// Use a ContextPropagator so that the KeyID value set in the workflow context is
		// also available in the context for activities.
		ContextPropagators: []workflow.ContextPropagator{
			bsdc.NewContextPropagator(),
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx = context.WithValue(ctx, bsdc.PropagatedValuesKey, bsdc.PropagatedValues{
		TenantId:              "tenant1",
		BlobStorePathSegments: []string{"tenant1", "starter"},
	})

	workflowOptions := client.StartWorkflowOptions{
		ID:        "blobstore_codec_wfID",
		TaskQueue: "blobstore_codec",
	}

	we, err := c.ExecuteWorkflow(
		ctx,
		workflowOptions,
		bsdc.Workflow,
		"Starter: big big blob",
	)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(ctx, &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
