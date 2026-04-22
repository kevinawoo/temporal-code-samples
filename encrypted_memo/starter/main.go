package main

import (
	"context"
	"log"

	"encrypted_memo"
	"encrypted_memo/codec"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		DataConverter: codec.DefaultEncryptionCodec,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "encryption_workflowID",
		TaskQueue: "encryption",
	}

	ctx := context.Background()

	// The workflow input "My Secret Friend" will be encrypted by the DataConverter before being sent to Temporal
	we, err := c.ExecuteWorkflow(
		ctx,
		workflowOptions,
		encryption.Workflow,
		"My Secret Friend",
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

	resp, err := c.DescribeWorkflowExecution(context.Background(), we.GetID(), we.GetRunID())
	if err != nil {
		log.Fatalln("Unable to describe workflow", err)
	}

	// read memo
	memo, err := decodeMemo(resp.GetWorkflowExecutionInfo().GetMemo())
	if err != nil {
		log.Fatalln("Unable to decode workflow memo", err)
	}

	log.Println("Workflow memo:", memo)
}

func decodeMemo(memo *commonpb.Memo) (map[string]interface{}, error) {
	fields := memo.GetFields()
	if len(fields) == 0 {
		return map[string]interface{}{}, nil
	}

	decoded := make(map[string]interface{}, len(fields))
	for key, payload := range fields {
		var value interface{}
		if err := codec.DefaultEncryptionCodec.FromPayload(payload, &value); err != nil {
			return nil, err
		}
		decoded[key] = value
	}

	return decoded, nil
}
