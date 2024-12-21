# Pause Workflow via Signal
This sample shows how to pause and resume a single workflow using a signal.

This is build using workflow interceptors.

# Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
    ```
    go run pause-workflow-via-signal/worker/main.go
    ```
3) Run the following command to start the example
    ```
    go run pause-workflow-via-signal/starter/main.go
    ```
4) Pause or Resume the workflow by running
    ```bash
    temporal workflow signal -w pause_workflow_ID --name pause
    temporal workflow signal -w pause_workflow_ID --name resume
    ```

Note, there's an interesting behavior if you send `pause, resume, pause` 
very quickly (or before a worker can pick it up), it will let 1 activity through.
This makes sense because the execution is linear, so it really becomes:
- pause signal sent
- interceptor **queues up** the next activity
- resume signal sent
- interceptor releases that activity
- pause signal sent
- interceptor **queues up** the next activity

It's probably possible to change the behavior so that the interceptor 
reduces over the signals and only uses the last result,
but that feels really weird/wrong.
