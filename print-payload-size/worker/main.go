package main

import (
	"fmt"
	clientLogger "go.temporal.io/sdk/log"
	"google.golang.org/grpc"
	"log"
	"print_payload_size"
	"print_payload_size/grpc_stats"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		Logger: &simpleLogger{},
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

	w := worker.New(c, "hello-world", worker.Options{})

	w.RegisterWorkflow(print_payload_size.Workflow)
	w.RegisterActivity(print_payload_size.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

// simpleLogger	drops keyvals because they contain the large payloads which we don't want printed out
type simpleLogger struct{}

var _ = (*clientLogger.Logger)(nil)

func (simpleLogger) Debug(msg string, keyvals ...interface{}) {
	fmt.Printf("DEBUG: " + msg + "\n")
}

func (simpleLogger) Info(msg string, keyvals ...interface{}) {
	fmt.Printf("INFO: " + msg + "\n")
}

func (simpleLogger) Warn(msg string, keyvals ...interface{}) {
	fmt.Printf("WARN: " + msg + "\n")
}

func (simpleLogger) Error(msg string, keyvals ...interface{}) {
	fmt.Printf("ERROR: " + msg + "\n")
}
