package main

import (
	"context"
	"fmt"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflowservice/v1"
	"log"
	"sync"

	"go.temporal.io/sdk/client"
)

func main() {
	// Temporal Cloud connection settings via API Key
	//namespace := "woo-apikey.a2dd6"
	//apiKey := os.Getenv("TEMPORAL_API_KEY")
	//
	//clientOptions := client.Options{
	//	HostPort:    "us-east-1.aws.api.temporal.io:7233",
	//	Namespace:   namespace,
	//	Credentials: client.NewAPIKeyStaticCredentials(apiKey),
	//	ConnectionOptions: client.ConnectionOptions{
	//		TLS: &tls.Config{},
	//		DialOptions: []grpc.DialOption{
	//			grpc.WithUnaryInterceptor(
	//				func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	//					return invoker(metadata.AppendToOutgoingContext(ctx, "temporal-namespace", namespace),
	//						method,
	//						req,
	//						reply,
	//						cc,
	//						opts...,
	//					)
	//				})}},
	//}

	// local connection settings
	clientOptions := client.Options{}
	namespace := "default"

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	d := NewSearch(namespace, c, `StartTime>="2025-05-08T07:00:00.000Z"`)

	err = d.search(ctx)
	if err != nil {
		log.Fatalf("unable to download workflows: %v\n", err)
	}
}

type search struct {
	namespace    string
	outputDir    string
	client       client.Client
	query        string
	pageSize     int32
	numOfWorkers int
}

func NewSearch(namespace string, client client.Client, query string) *search {
	return &search{
		namespace:    namespace,
		client:       client,
		query:        query,
		pageSize:     1000,
		numOfWorkers: 3,
	}
}

func (d *search) search(ctx context.Context) error {
	jobs := make(chan *common.WorkflowExecution)

	workersDone := d.fetchers(ctx, jobs)

	go func() {
		var nextPageToken []byte
		for {
			wfs, err := d.client.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
				Namespace:     d.namespace,
				Query:         d.query,
				NextPageToken: nextPageToken,
				PageSize:      d.pageSize,
			})
			if err != nil {
				log.Printf("unable to list workflows: %v\n", err)
			}

			found := 0
			if wfs != nil {
				found = len(wfs.Executions)
			}
			log.Printf("found %d workflows\n", found)

			if wfs == nil {
				break
			}

			for _, w := range wfs.Executions {
				jobs <- w.Execution
			}

			if wfs.NextPageToken != nil {
				nextPageToken = wfs.NextPageToken
				continue
			}

			break
		}
		close(jobs)
	}()

	workersDone.Wait()

	return nil
}

func (d *search) fetchers(ctx context.Context, workflowC chan *common.WorkflowExecution) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	for id := 0; id < d.numOfWorkers; id++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()

			for wf := range workflowC {
				log.Printf("worker%d: downloading workflow: %s, %s\n", workerId, wf.WorkflowId, wf.RunId)

				iter := d.client.GetWorkflowHistory(ctx, wf.WorkflowId, wf.RunId, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

				var events []*history.HistoryEvent
				for iter.HasNext() {
					event, err := iter.Next()
					if err != nil {
						log.Printf("worker%d: failed getting event: %v\n", workerId, err)
						break
					}
					events = append(events, event)
				}
				hist := &history.History{Events: events}
				d.inspectEventHistory(hist, events, wf)
			}
		}(id)
	}

	return wg
}

func (d *search) inspectEventHistory(hist *history.History, events []*history.HistoryEvent, wf *common.WorkflowExecution) {
	for _, event := range hist.GetEvents() {
		switch event.GetEventType() {
		case enums.EVENT_TYPE_ACTIVITY_TASK_STARTED:
			scheduleId := event.GetActivityTaskStartedEventAttributes().GetScheduledEventId()
			// events are 0 indexed, where as eventIds are 1 based
			name := events[scheduleId-1].GetActivityTaskScheduledEventAttributes().GetActivityType().GetName()
			attempts := event.GetActivityTaskStartedEventAttributes().GetAttempt()

			if name == "Activity" && attempts > 0 {
				fmt.Println("match:", wf.WorkflowId, wf.RunId)
			}
		}
	}
}
