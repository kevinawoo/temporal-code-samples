package main

import (
	"context"
	"fmt"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/temporalproto"
	"go.temporal.io/api/workflowservice/v1"
	"log"
	"os"
	"sync"

	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	d := NewDownloader("default", "out", c, `WorkflowType = 'Workflow'`)

	err = d.downloadWorkflows(ctx)
	if err != nil {
		log.Fatalf("unable to download workflows: %v\n", err)
	}
}

type downloader struct {
	namespace    string
	outputDir    string
	jsonIndent   string
	client       client.Client
	query        string
	pageSize     int32
	numOfWorkers int
}

func NewDownloader(namespace, outputDir string, client client.Client, query string) *downloader {
	return &downloader{
		namespace:    namespace,
		outputDir:    outputDir,
		jsonIndent:   "",
		client:       client,
		query:        query,
		pageSize:     1000,
		numOfWorkers: 3,
	}
}

func (d *downloader) downloadWorkflows(ctx context.Context) error {
	jobs := make(chan *common.WorkflowExecution)

	workersDone := d.startDownloadWorkers(ctx, jobs)

	err := os.MkdirAll(d.outputDir, 0755) // ensure dir exists
	if err != nil {
		return fmt.Errorf("unable to create output directory: %w", err)
	}

	go func() {
		var nextPageToken []byte
		for {
			wfs, err := d.client.ScanWorkflow(ctx, &workflowservice.ScanWorkflowExecutionsRequest{
				Namespace:     d.namespace,
				Query:         d.query,
				NextPageToken: nextPageToken,
				PageSize:      d.pageSize,
			})
			if err != nil {
				log.Printf("unable to scan workflows: %v\n", err)
			}

			log.Printf("found %d workflows\n", len(wfs.Executions))

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

func (d *downloader) startDownloadWorkers(ctx context.Context, workflowC chan *common.WorkflowExecution) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	for id := 0; id < d.numOfWorkers; id++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()

			for w := range workflowC {
				log.Printf("worker%d: downloading workflow: %s, %s\n", workerId, w.WorkflowId, w.RunId)

				iter := d.client.GetWorkflowHistory(ctx, w.WorkflowId, w.RunId, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

				events := []*history.HistoryEvent{}
				for iter.HasNext() {
					event, err := iter.Next()
					if err != nil {
						log.Printf("worker%d: failed getting event: %v\n", workerId, err)
						break
					}
					events = append(events, event)
				}
				hist := &history.History{Events: events}

				m := &temporalproto.CustomJSONMarshalOptions{
					Indent: d.jsonIndent,
				}
				b, err := m.Marshal(hist)
				if err != nil {
					log.Printf("worker%d: failed encoding history: %v\n", workerId, err)
					continue
				}

				filename := fmt.Sprintf("%s/%s-%s.json", d.outputDir, w.WorkflowId, w.RunId)
				log.Printf("worker%d: writing to file: %s\n", workerId, filename)
				err = os.WriteFile(filename, b, 0644)
				if err != nil {
					log.Printf("worker%d: failed writing file: %v\n", workerId, err)
					continue
				}
			}
		}(id)
	}

	return wg
}
