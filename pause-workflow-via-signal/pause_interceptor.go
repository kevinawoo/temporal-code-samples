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
}

var _ interceptor.WorkerInterceptor = (*PauseInterceptor)(nil) // ensure interface is implemented

// InterceptWorkflow
//
//	Starter > Temporal Server > * > 1st line of the Workflow
//
// This handles the case when the workflow is started in a paused state
func (pi *PauseInterceptor) InterceptWorkflow(ctx workflow.Context, next interceptor.WorkflowInboundInterceptor) interceptor.WorkflowInboundInterceptor {
	// create an instance of the inbound interceptor for this workflow
	// note: this is unique for each workflow execution
	return &wfInbound{
		WorkflowInboundInterceptorBase: interceptor.WorkflowInboundInterceptorBase{Next: next},
		wfOutbound:                     nil, // set in wfInbound.Init
	}
}

// wfInbound is strict 1:1 to workflow execution
type wfInbound struct {
	interceptor.WorkflowInboundInterceptorBase
	wfOutbound *wfOutbound
}

var _ interceptor.WorkflowInboundInterceptor = (*wfInbound)(nil) // ensure interface is implemented

func (i *wfInbound) Init(outbound interceptor.WorkflowOutboundInterceptor) error {
	ob := &wfOutbound{
		WorkflowOutboundInterceptorBase: interceptor.WorkflowOutboundInterceptorBase{Next: outbound},
		paused:                          false,
	}
	i.wfOutbound = ob

	return i.Next.Init(ob)
}

// HandleSignal is between Temporal Server > * > execute WorkflowSignalHandler
// Signals will be intercepted here to be used to control the pause state of the workflow
func (i *wfInbound) HandleSignal(ctx workflow.Context, in *interceptor.HandleSignalInput) error {
	fmt.Println("HandleSignal interceptor got signal: ", in.SignalName)

	switch in.SignalName {
	case PauseSignalName:
		i.wfOutbound.paused = true
	case ResumeSignalName:
		i.wfOutbound.paused = false
	}

	return i.Next.HandleSignal(ctx, in)
}

// You can image each call to a WorkflowOutboundInterceptor method is like
// prepending code before executing the actual SDK call.
type wfOutbound struct {
	interceptor.WorkflowOutboundInterceptorBase
	paused bool
}

var _ interceptor.WorkflowOutboundInterceptor = (*wfOutbound)(nil) // ensure interface is implemented

func (o *wfOutbound) handlePause(ctx workflow.Context) {
	// If we want to set the SearchAttribute, we can only do this when we're "inside a workflow".
	// Inbound interceptors execute before the worker loads the workflow state,
	// thus we'll hijack outbound calls to prepend the Upsert.
	err := workflow.UpsertTypedSearchAttributes(ctx, PauseSearchAttrKey.ValueSet(o.paused))
	if err != nil {
		panic(fmt.Errorf("failed to pause workflow: %w", err))
	}

	workflow.GetSignalChannel(ctx, PauseSignalName) // register the pause signal name for a nicer experience in the UI

	if o.paused {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(workflow.GetSignalChannel(ctx, ResumeSignalName), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			fmt.Printf("resume signal received")
		})
		fmt.Println("waiting for a resume signal")
		selector.Select(ctx)
	}
	return
}

// ExecuteActivity interceptor is between: WorkflowCode.ExecuteActivity >...*...> RunningAnActivity
// We're still in the workflow context, so we can add any workflow logic here.
//
// You'll also want to consider which other sdk functions need the pause functionality.
// In this case, Activities-like things are good enough.
func (o *wfOutbound) ExecuteActivity(ctx workflow.Context, activityType string, args ...any) workflow.Future {
	fmt.Println("ExecuteActivity interceptor, paused?", o.paused)
	o.handlePause(ctx)
	return o.Next.ExecuteActivity(ctx, activityType, args...)
}

func (o *wfOutbound) ExecuteLocalActivity(ctx workflow.Context, activityType string, args ...interface{}) workflow.Future {
	fmt.Println("ExecuteLocalActivity interceptor, paused?", o.paused)
	o.handlePause(ctx)
	return o.Next.ExecuteLocalActivity(ctx, activityType, args...)
}

func (o *wfOutbound) ExecuteChildWorkflow(ctx workflow.Context, childWorkflowType string, args ...interface{}) workflow.ChildWorkflowFuture {
	fmt.Println("ExecuteChildWorkflow interceptor, paused?", o.paused)
	o.handlePause(ctx)
	return o.Next.ExecuteChildWorkflow(ctx, childWorkflowType, args...)
}

func (o *wfOutbound) ExecuteNexusOperation(ctx workflow.Context, input interceptor.ExecuteNexusOperationInput) workflow.NexusOperationFuture {
	fmt.Println("ExecuteNexusOperation interceptor, paused?", o.paused)
	o.handlePause(ctx)
	return o.Next.ExecuteNexusOperation(ctx, input)
}
