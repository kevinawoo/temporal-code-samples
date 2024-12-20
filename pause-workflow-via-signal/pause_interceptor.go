package pause_workflow_via_signal

import (
	"fmt"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	SearchAttributeName = "PauseField"
	PauseSignalName     = "pause"
	ResumeSignalName    = "resume"
)

var (
	PauseSearchAttrKey = temporal.NewSearchAttributeKeyBool(SearchAttributeName)
)

func NewPauseInterceptor() *PauseInterceptor {
	return &PauseInterceptor{}
}

type PauseInterceptor struct {
	interceptor.WorkerInterceptorBase

	Paused bool
}

// InterceptWorkflow is between Starter > TemporalServer > * > start executing Workflow
// This handles the case when the workflow is started in a paused state
func (pi *PauseInterceptor) InterceptWorkflow(ctx workflow.Context, next interceptor.WorkflowInboundInterceptor) interceptor.WorkflowInboundInterceptor {
	attr := workflow.GetTypedSearchAttributes(ctx)
	pi.Paused, _ = attr.GetBool(PauseSearchAttrKey)
	fmt.Println("InterceptWorkflow, paused?", pi.Paused)

	return &wfInbound{
		pi:                             pi,
		WorkflowInboundInterceptorBase: interceptor.WorkflowInboundInterceptorBase{Next: next},
	}
}

func (pi *PauseInterceptor) handlePause(ctx workflow.Context) {
	err := workflow.UpsertTypedSearchAttributes(ctx, PauseSearchAttrKey.ValueSet(pi.Paused))
	if err != nil {
		panic(fmt.Errorf("failed to pause workflow: %w", err))
	}

	if pi.Paused {
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(workflow.GetSignalChannel(ctx, ResumeSignalName), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			fmt.Printf("resume signal received")
		})
		fmt.Println("waiting for a resume signal")
		selector.Select(ctx)
	}
}

type wfInbound struct {
	interceptor.WorkflowInboundInterceptorBase
	pi *PauseInterceptor
}

var _ interceptor.WorkflowInboundInterceptor = (*wfInbound)(nil) // ensure interface is implemented

func (i *wfInbound) Init(outbound interceptor.WorkflowOutboundInterceptor) error {
	poi := &wfOutbound{
		pi:                              i.pi,
		WorkflowOutboundInterceptorBase: interceptor.WorkflowOutboundInterceptorBase{Next: outbound},
	}

	return i.Next.Init(poi)
}

// HandleSignal is between Temporal Server >...*...> execute WorkflowSignalHandler
// Signals will be intercepted here to control the workflow's pause state
func (i *wfInbound) HandleSignal(ctx workflow.Context, in *interceptor.HandleSignalInput) error {
	fmt.Println("HandleSignal interceptor got signal: ", in.SignalName)

	switch in.SignalName {
	case PauseSignalName:
		i.pi.Paused = true
	case ResumeSignalName:
		i.pi.Paused = false
	}

	return i.Next.HandleSignal(ctx, in)
}

type wfOutbound struct {
	interceptor.WorkflowOutboundInterceptorBase
	pi *PauseInterceptor
}

var _ interceptor.WorkflowOutboundInterceptor = (*wfOutbound)(nil) // ensure interface is implemented

// ExecuteActivity interceptor is between: WorkflowCode.ExecuteActivity >...*...> RunningAnActivity
// We're still in the workflow context, so we can add any workflow logic here.
//
// You'll also want to consider which other sdk calls you may want intercept to pause.
func (i *wfOutbound) ExecuteActivity(ctx workflow.Context, activityType string, args ...any) workflow.Future {
	fmt.Println("ExecuteActivity interceptor, paused?", i.pi.Paused)
	i.pi.handlePause(ctx)
	return i.Next.ExecuteActivity(ctx, activityType, args...)
}

func (i *wfOutbound) ExecuteLocalActivity(ctx workflow.Context, activityType string, args ...interface{}) workflow.Future {
	fmt.Println("ExecuteLocalActivity interceptor, paused?", i.pi.Paused)
	i.pi.handlePause(ctx)
	return i.Next.ExecuteLocalActivity(ctx, activityType, args...)
}

func (i *wfOutbound) ExecuteChildWorkflow(ctx workflow.Context, childWorkflowType string, args ...interface{}) workflow.ChildWorkflowFuture {
	fmt.Println("ExecuteChildWorkflow interceptor, paused?", i.pi.Paused)
	i.pi.handlePause(ctx)
	return i.Next.ExecuteChildWorkflow(ctx, childWorkflowType, args...)
}

func (i *wfOutbound) ExecuteNexusOperation(ctx workflow.Context, input interceptor.ExecuteNexusOperationInput) workflow.NexusOperationFuture {
	fmt.Println("ExecuteNexusOperation interceptor, paused?", i.pi.Paused)
	i.pi.handlePause(ctx)
	return i.Next.ExecuteNexusOperation(ctx, input)
}
