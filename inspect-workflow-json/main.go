package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

type WorkflowHistory struct {
	Events []struct {
		EventId                                 string    `json:"eventId"`
		EventTime                               time.Time `json:"eventTime"`
		EventType                               string    `json:"eventType"`
		Version                                 string    `json:"version"`
		TaskId                                  string    `json:"taskId"`
		WorkflowExecutionStartedEventAttributes struct {
			WorkflowType struct {
				Name string `json:"name"`
			} `json:"workflowType"`
			TaskQueue struct {
				Name string `json:"name"`
				Kind string `json:"kind"`
			} `json:"taskQueue"`
			Input struct {
				Payloads []struct {
					Metadata struct {
						Encoding string `json:"encoding"`
					} `json:"metadata"`
					Data struct {
						Payload struct {
							Event           string `json:"event"`
							Method          string `json:"method"`
							Path            string `json:"path"`
							Service         string `json:"service"`
							StoredPayloadID string `json:"storedPayloadID"`
						} `json:"payload,omitempty"`
						ApiLogID           int    `json:"apiLogID,omitempty"`
						SubscriberClientID int    `json:"subscriberClientID,omitempty"`
						ClientID           int    `json:"clientID,omitempty"`
						WebhookInstanceID  string `json:"webhookInstanceID,omitempty"`
						Details            string `json:"details,omitempty"`
						Targets            []struct {
							Id             string      `json:"id"`
							CreatedOn      time.Time   `json:"createdOn"`
							DeletedOn      interface{} `json:"deletedOn"`
							ClientId       int         `json:"clientId"`
							TargetURL      string      `json:"targetURL"`
							PayloadFormat  string      `json:"payloadFormat"`
							EnvironmentId  int         `json:"environmentId"`
							AuthToken      string      `json:"authToken"`
							BasicAuthToken string      `json:"basicAuthToken"`
						} `json:"targets,omitempty"`
						ActiveFeatures []int `json:"activeFeatures,omitempty"`
					} `json:"data"`
				} `json:"payloads"`
			} `json:"input"`
			WorkflowExecutionTimeout string `json:"workflowExecutionTimeout"`
			WorkflowRunTimeout       string `json:"workflowRunTimeout"`
			WorkflowTaskTimeout      string `json:"workflowTaskTimeout"`
			OriginalExecutionRunId   string `json:"originalExecutionRunId"`
			Identity                 string `json:"identity"`
			FirstExecutionRunId      string `json:"firstExecutionRunId"`
			Attempt                  int    `json:"attempt"`
			FirstWorkflowTaskBackoff string `json:"firstWorkflowTaskBackoff"`
			Header                   struct {
			} `json:"header"`
			WorkflowId string `json:"workflowId"`
		} `json:"workflowExecutionStartedEventAttributes,omitempty"`
		WorkflowTaskScheduledEventAttributes struct {
			TaskQueue struct {
				Name       string `json:"name"`
				Kind       string `json:"kind"`
				NormalName string `json:"normalName,omitempty"`
			} `json:"taskQueue"`
			StartToCloseTimeout string `json:"startToCloseTimeout"`
			Attempt             int    `json:"attempt"`
		} `json:"workflowTaskScheduledEventAttributes,omitempty"`
		WorkflowTaskStartedEventAttributes struct {
			ScheduledEventId string `json:"scheduledEventId"`
			Identity         string `json:"identity"`
			RequestId        string `json:"requestId"`
			HistorySizeBytes string `json:"historySizeBytes"`
			WorkerVersion    struct {
				BuildId string `json:"buildId"`
			} `json:"workerVersion"`
		} `json:"workflowTaskStartedEventAttributes,omitempty"`
		WorkflowTaskCompletedEventAttributes struct {
			ScheduledEventId string `json:"scheduledEventId"`
			StartedEventId   string `json:"startedEventId"`
			Identity         string `json:"identity"`
			WorkerVersion    struct {
				BuildId string `json:"buildId"`
			} `json:"workerVersion"`
			SdkMetadata struct {
				LangUsedFlags []int `json:"langUsedFlags,omitempty"`
			} `json:"sdkMetadata"`
			MeteringMetadata struct {
			} `json:"meteringMetadata"`
		} `json:"workflowTaskCompletedEventAttributes,omitempty"`
		ActivityTaskScheduledEventAttributes struct {
			ActivityId   string `json:"activityId"`
			ActivityType struct {
				Name string `json:"name"`
			} `json:"activityType"`
			TaskQueue struct {
				Name string `json:"name"`
				Kind string `json:"kind"`
			} `json:"taskQueue"`
			Header struct {
			} `json:"header"`
			Input struct {
				Payloads []struct {
					Metadata struct {
						Encoding string `json:"encoding"`
					} `json:"metadata"`
					Data interface{} `json:"data"`
				} `json:"payloads"`
			} `json:"input"`
			ScheduleToCloseTimeout       string `json:"scheduleToCloseTimeout"`
			ScheduleToStartTimeout       string `json:"scheduleToStartTimeout"`
			StartToCloseTimeout          string `json:"startToCloseTimeout"`
			HeartbeatTimeout             string `json:"heartbeatTimeout"`
			WorkflowTaskCompletedEventId string `json:"workflowTaskCompletedEventId"`
			RetryPolicy                  struct {
				InitialInterval    string `json:"initialInterval"`
				BackoffCoefficient int    `json:"backoffCoefficient"`
				MaximumInterval    string `json:"maximumInterval"`
				MaximumAttempts    int    `json:"maximumAttempts"`
			} `json:"retryPolicy"`
			UseWorkflowBuildId bool `json:"useWorkflowBuildId"`
		} `json:"activityTaskScheduledEventAttributes,omitempty"`
		ActivityTaskStartedEventAttributes struct {
			ScheduledEventId string `json:"scheduledEventId"`
			Identity         string `json:"identity"`
			RequestId        string `json:"requestId"`
			Attempt          int    `json:"attempt"`
			WorkerVersion    struct {
				BuildId string `json:"buildId"`
			} `json:"workerVersion"`
		} `json:"activityTaskStartedEventAttributes,omitempty"`
		ActivityTaskCompletedEventAttributes struct {
			Result struct {
				Payloads []struct {
					Metadata struct {
						Encoding string `json:"encoding"`
					} `json:"metadata"`
					Data struct {
						Targets []struct {
							Id             string      `json:"id"`
							CreatedOn      time.Time   `json:"createdOn"`
							DeletedOn      interface{} `json:"deletedOn"`
							ClientId       int         `json:"clientId"`
							TargetURL      string      `json:"targetURL"`
							PayloadFormat  string      `json:"payloadFormat"`
							EnvironmentId  int         `json:"environmentId"`
							AuthToken      string      `json:"authToken"`
							BasicAuthToken string      `json:"basicAuthToken"`
						} `json:"Targets"`
						Errors            interface{} `json:"Errors"`
						TargetAndInstance []struct {
							Target struct {
								Id             string      `json:"id"`
								CreatedOn      time.Time   `json:"createdOn"`
								DeletedOn      interface{} `json:"deletedOn"`
								ClientId       int         `json:"clientId"`
								TargetURL      string      `json:"targetURL"`
								PayloadFormat  string      `json:"payloadFormat"`
								EnvironmentId  int         `json:"environmentId"`
								AuthToken      string      `json:"authToken"`
								BasicAuthToken string      `json:"basicAuthToken"`
							} `json:"Target"`
							Instance struct {
								Id                 string      `json:"id"`
								CreatedOn          time.Time   `json:"createdOn"`
								TemporalWorkflowID string      `json:"temporalWorkflowID"`
								WebhookInstanceID  string      `json:"webhookInstanceID"`
								WebhookTargetID    string      `json:"webhookTargetID"`
								CompletedOn        interface{} `json:"completedOn"`
								IsSuccess          bool        `json:"isSuccess"`
							} `json:"Instance"`
						} `json:"TargetAndInstance"`
					} `json:"data"`
				} `json:"payloads"`
			} `json:"result,omitempty"`
			ScheduledEventId string `json:"scheduledEventId"`
			StartedEventId   string `json:"startedEventId"`
			Identity         string `json:"identity"`
		} `json:"activityTaskCompletedEventAttributes,omitempty"`
		WorkflowExecutionCompletedEventAttributes struct {
			WorkflowTaskCompletedEventId string `json:"workflowTaskCompletedEventId"`
		} `json:"workflowExecutionCompletedEventAttributes,omitempty"`
	} `json:"events"`
}

//go:embed example_events.json
var wfHistoryJSON string

func main() {
	wfHistory := WorkflowHistory{}
	err := json.Unmarshal([]byte(wfHistoryJSON), &wfHistory)
	if err != nil {
		panic(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
	_, _ = fmt.Fprintf(w, "EventId\t")
	_, _ = fmt.Fprintf(w, "EventType\t")
	_, _ = fmt.Fprintf(w, "EventTime\t")
	_, _ = fmt.Fprintf(w, "Unix\t")
	_, _ = fmt.Fprintf(w, "Delta\n")

	prevEventTime := wfHistory.Events[0].EventTime
	for _, event := range wfHistory.Events {
		delta := event.EventTime.Sub(prevEventTime)
		_, _ = fmt.Fprintf(w, "%s\t", event.EventId)
		_, _ = fmt.Fprintf(w, "%s\t", event.EventType)
		_, _ = fmt.Fprintf(w, "%s\t", event.EventTime.Format(time.RFC3339))
		_, _ = fmt.Fprintf(w, "%d\t", event.EventTime.UnixMilli())
		_, _ = fmt.Fprintf(w, "%s\n", delta)
		prevEventTime = event.EventTime
	}
	_ = w.Flush()
}
