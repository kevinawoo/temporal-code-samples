# Blobstore DataConverter
This sample demonstrates how to use the DataConverter to store large payloads greater than a certain size 
in a blobstore and passes the object path around in the Temporal Event History.

The payload size limit is set in [codec.go: `payloadSizeLimit`](./codec.go#L20).

It relies on the use of context propagation to pass blobstore config metadata, like object path prefixes.

In this example, we prefix all object paths with a tenantID. There is an edge case when supplying input via UI/CLI, see warning below.

> [!WARNING]
> As of `Temporal UI v2.33.1` (`Temporal v1.25.2`) and `Temporal CLI v1.1.1`, they **do not** have the ability to send context headers.
> This means that Workflow Start, Signal, Queries, etc. from the UI/CLI will pass payloads to the codec-server but the 
> worker will need handle a missing context propagation header.
> 
> In this sample when the header is missing, we use a default  `PropagatedValues{TenantID: "unknownTenant"}`, 
> see [propagator.go: `missingHeaderContextPropagationKeyError`](./propagator.go#L66).
> 
> This allows this sample to still work with the UI/CLI. This might be fine depending on your requirements. 

> [!CAUTION]
> Bug: when using the UI/CLI with a codec-server, if the codec-server is not running
> the UI/CLI will fail to encode the payload and an empty payload will be sent to the Temporal Server.


### Steps to run this sample:
1. Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2. Run the following command to start the worker
    ```
    go run worker/main.go
    ```
3. Run the following command to start the example
    ```
    go run starter/main.go
    ```
4. Open the Temporal Web UI and observe the workflow events, the Workflow Input, and Activity Input will be visible,
   but the Activity Result, and Workflow Result will be an object path.
5. Run the following command to start the remote codec server
    ```
    go run ./codec-server
    ```
6. Open the Temporal Web UI and observe the workflow execution, all payloads will now be human-readable.
7. You can use the Temporal CLI as well
    ```
    # payloads will be object paths
    temporal workflow show --env local -w WORKFLOW_ID

    # payloads will be visible
    temporal --codec-endpoint 'http://localhost:8081/' workflow show -w WORKFLOW_ID
    ``````

Note: Please see the [codec-server](../codec-server/) sample for a more complete example of a codec server which provides UI decoding and oauth.
