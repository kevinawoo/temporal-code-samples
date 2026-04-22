package main

import (
	"log"

	"encrypted_memo"
	"encrypted_memo/codec"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		DataConverter: codec.DefaultEncryptionCodec,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "encryption", worker.Options{})

	w.RegisterWorkflow(encryption.Workflow)
	w.RegisterActivity(encryption.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
