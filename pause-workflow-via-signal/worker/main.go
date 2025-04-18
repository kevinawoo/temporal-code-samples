package main

import (
	pause "github.com/kevinawoo/temporal-code-samples/pause-workflow-via-signal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"log"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "pause-workflow", worker.Options{
		Interceptors: []interceptor.WorkerInterceptor{
			pause.NewPauseInterceptor(pause.Workflow),
		},
	})

	w.RegisterWorkflow(pause.Workflow)
	w.RegisterActivity(pause.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
