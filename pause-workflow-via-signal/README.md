### Steps to run this sample:

1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
    ```
    go run pause-workflow-via-signal/worker/main.go
    ```
3) Run the following command to start the example
    ```
    go run pause-workflow-via-signal/starter/main.go
    ```
4) Toggle the workflow by running toggle example
    ```
    go run pause-workflow-via-signal/toggle/main.go
    ```
