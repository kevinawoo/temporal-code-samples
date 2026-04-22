package encryption

import (
	"context"
	"fmt"
	"strings"
	"time"

	"encrypted_memo/codec"

	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

// Workflow is a standard workflow definition.
// Note that the Workflow and Activity don't need to care that
// their inputs/results are being encrypted/decrypted.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Encrypted Payloads workflow started", "name", name)

	memo := map[string]interface{}{
		"Key1": 2,
		"Key2": true,
		"Key3": "seattle",
	}
	err := workflow.UpsertMemo(ctx, memo)
	if err != nil {
		return "", err
	}
	memo = map[string]interface{}{} // clear it

	info := map[string]string{
		"name": name,
	}

	var result string
	err = workflow.ExecuteActivity(ctx, Activity, info).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("Encrypted Payloads workflow completed.", "result", result)

	wfInfo := workflow.GetInfo(ctx)
	fmt.Println("workflow.go:51")
	err = printMemo(wfInfo.Memo, logger)
	if err != nil {
		return "", err
	}
	fmt.Println(wfInfo.Memo)

	return result, nil
}

func Activity(ctx context.Context, info map[string]string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "info", info)

	name, ok := info["name"]
	if !ok {
		name = "someone"
	}

	return "Hello " + name + "!", nil
}

func printMemo(memo *common.Memo, logger log.Logger) error {
	if memo == nil || len(memo.GetFields()) == 0 {
		logger.Info("Current memo is empty.")
		return nil
	}

	var builder strings.Builder
	//workflowcheck:ignore Only iterates for logging reasons
	for k, v := range memo.GetFields() {
		var currentVal interface{}
		err := codec.DefaultEncryptionCodec.FromPayload(v, &currentVal)
		if err != nil {
			logger.Error(fmt.Sprintf("Get memo for key %s failed.", k), "Error", err)
			return err
		}
		builder.WriteString(fmt.Sprintf("%s=%v\n", k, currentVal))
	}
	logger.Info(fmt.Sprintf("Current memo values:\n%s", builder.String()))
	return nil
}
