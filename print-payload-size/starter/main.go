package main

import (
	"context"
	"fmt"
	enumspb "go.temporal.io/api/enums/v1"
	"google.golang.org/grpc"
	"log"
	"print_payload_size"
	"print_payload_size/grpc_stats"

	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		ConnectionOptions: client.ConnectionOptions{
			DialOptions: []grpc.DialOption{
				grpc.WithStatsHandler(grpc_stats.New()),
			},
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:                    "hello_world_workflowID",
		TaskQueue:             "hello-world",
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	payload := print_payload_size.RandStringBytes(1_000_000 * 1)
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, print_payload_size.Workflow, payload)
	_, _ = we, err
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	fmt.Println("starter waiting to close")
	select {}
}
