package grpc_stats

import (
	"context"
	"fmt"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/grpc/stats"
	"path/filepath"
	"reflect"
)

type Handler struct{}

// implements [stats.Handler](https://pkg.go.dev/google.golang.org/grpc/stats#Handler) interface.
var _ stats.Handler = (*Handler)(nil)

func New() *Handler {
	return &Handler{}
}

type connStatCtxKey struct{}

// TagConn can attach some information to the given context.
// The context used in HandleConn for this connection will be derived from the context returned.
// In the gRPC client:
// The context used in HandleRPC for RPCs on this connection will be the user's context and NOT derived from the context returned here.
// In the gRPC server:
// The context used in HandleRPC for RPCs on this connection will be derived from the context returned here.
func (st *Handler) TagConn(ctx context.Context, stat *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn processes the Conn stats.
func (st *Handler) HandleConn(ctx context.Context, stat stats.ConnStats) {
}

type rpcStatCtxKey struct{}

// TagRPC can attach some information to the given context.
// The context used for the rest lifetime of the RPC will be derived from the returned context.
func (st *Handler) TagRPC(ctx context.Context, stat *stats.RPCTagInfo) context.Context {
	return context.WithValue(ctx, rpcStatCtxKey{}, stat)
}

// HandleRPC processes the RPC stats. Note: All stat fields are read-only.
func (st *Handler) HandleRPC(ctx context.Context, stat stats.RPCStats) {
	var methodName string
	if s, ok := ctx.Value(rpcStatCtxKey{}).(*stats.RPCTagInfo); ok {
		methodName = filepath.Base(s.FullMethodName)
	}

	fmt.Println("handleRPC:", methodName, reflect.TypeOf(stat))

	switch methodName {
	case "StartWorkflowExecution":
		switch stat := stat.(type) {
		//case *stats.OutHeader:
		//	// There's usually not an issue with OutHeader unless there's a proxy adding headers
		//	fmt.Printf("%s %d out headers: %+v\n", methodName, stat.Header.Len(), stat.Header)
		case *stats.OutPayload:
			payload := stat.Payload.(*workflowservice.StartWorkflowExecutionRequest)
			_ = payload
			fmt.Println(methodName, "payload size", ByteCountSI(int64(stat.Length)))
		}

	// see helloworld Workflow for more details on RespondWorkflowTaskCompleted
	case "RespondWorkflowTaskCompleted":
		switch stat := stat.(type) {
		case *stats.OutPayload:
			// reasons the payload might exceed 4MB limit:
			//	- the payload itself might be too large
			//	- the SDK batching together Workflow Event commands, especially if the Workflow creates a bunch of async commands
			payload := stat.Payload.(*workflowservice.RespondWorkflowTaskCompletedRequest)
			fmt.Println(methodName, "payload size", ByteCountSI(int64(stat.Length)), "len(commands)", len(payload.Commands))
		}

	case "RespondActivityTaskCompleted":
		switch stat := stat.(type) {
		case *stats.OutPayload:
			payload := stat.Payload.(*workflowservice.RespondActivityTaskCompletedRequest)
			_ = payload
			fmt.Println(methodName, "payload size", ByteCountSI(int64(stat.Length)))
		}
	}
}

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
