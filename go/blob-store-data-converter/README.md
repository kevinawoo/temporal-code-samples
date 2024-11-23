This sample how to use the DataConverter to store payloads in a blobstore
and saves the reference in the Temporal Event History.

It relies on the use of context propagation to pass the blobstore config metadata.

> [!CAUTION]
> As of Temporal v1.25.2, the UI and CLI does not have the ability to set headers. This means that
> Workflow Start, Signal, Queries, etc from the UI/CLI will not be able to encode the payloads. 
>
> The UI/CLI is still able to decode payloads using the codec server.

Potential workarounds:
- Do not use the UI or CLI to start, signal, or query workflows
- Remove the DataConverter and have the Codec rely on specific payloads instead of using context propagation. 
  This not a great solution as Workflow Input, Activity Input, Activity Result, Workflow Output, etc. will need
  follow the same format and always rely on specific payloads.
- Don't use the DataConverter and Context Propagation as it apply across all event boundaries. Instead, be selective between
  event boundaries and store/retrieve from the blobstore inline, and only passing by reference.

Potential improvements:
- Have the Codec only offload payloads that are above a certain size to the blobstore.


### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the remote codec server
```
go run ./codec-server
```
3) Run the following command to start the worker
```
go run worker/main.go
```
4) Run the following command to start the example
```
go run starter/main.go
```
5) Run the following command and see the payloads cannot be decoded
```
tctl workflow show --wid blobstore_codec_workflow
```
6) Run the following command and see the decoded payloads
```
tctl --codec_endpoint 'http://localhost:8081/' workflow show --wid blobstore_codec_workflow
```

Note: Please see the [codec-server](../codec-server/) sample for a more complete example of a codec server which provides UI decoding and oauth.
