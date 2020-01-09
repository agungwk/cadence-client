package main

import (
	"flag"
	"time"

	"github.com/agungwk/cadence-client/usecase"
	"github.com/pborman/uuid"
	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

var version string

const (
	signalName = "sampleSignal"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper, applicationName string) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}
	h.StartWorkers(h.Config.DomainName, applicationName, workerOptions)
}

func startWorkflow(h *common.SampleHelper, workflow interface{}, applicationName string, params ...interface{}) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              applicationName + uuid.New(),
		TaskList:                        applicationName,
		ExecutionStartToCloseTimeout:    time.Minute * 5,
		DecisionTaskStartToCloseTimeout: time.Minute * 5,
		// Needed when using custom search attributes
		// SearchAttributes: map[string]interface{}{
		// 	"helloChangesVersion": 2,
		// },
	}
	h.StartWorkflow(workflowOptions, workflow, params...)
}

func main() {
	var mode, workflowID, applicationGroup, workflow string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.StringVar(&version, "v", "version", "Version of workflow run.")
	flag.StringVar(&workflowID, "id", "workflow ID", "Workflow ID")
	flag.StringVar(&applicationGroup, "a", "applicationGroup", "Application Group")
	flag.StringVar(&workflow, "w", "workflow", "Workflow")
	flag.Parse()

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		//startWorkers(&h, ApplicationName)
		switch workflow {
		case "simple":
			startWorkers(&h, usecase.HelloWorldApplicationName)
		case "orchestration":
			startWorkers(&h, usecase.CreateOrderGroupApplicationName)
		}

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "trigger":

		switch workflow {
		case "simple":
			startWorkflow(&h, usecase.Workflow, usecase.HelloWorldApplicationName, "Your name")
		case "orchestration":
			startWorkflow(&h, usecase.OrchestrationWorkflow, usecase.CreateOrderGroupApplicationName, "PKG-123", "VOUCHER")
		}
	case "signal":
		sendSignal(&h, workflowID)

	case "terminate":
		terminateWorkflow(&h, workflowID)

	case "query":
		query(&h)
	}
}

func createWorker() common.SampleHelper {
	return common.SampleHelper{}
}

/**
 * This is the hello world workflow sample.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "helloWorldGroup"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	workflow.Register(usecase.Workflow)
	activity.Register(usecase.HelloworldActivity)
	activity.Register(usecase.HelloworldGuys)
	activity.Register(usecase.HelloMoving)
	activity.Register(usecase.HelloNoContext)

	workflow.Register(usecase.OrchestrationWorkflow)
	activity.Register(usecase.ValidateProductPackage)
	activity.Register(usecase.ValidateVoucher)
	activity.Register(usecase.CreateOrder)
	activity.Register(usecase.CreatePayment)
	activity.Register(usecase.RollbackOrder)
}

func sendSignal(h *common.SampleHelper, workflowID string) {
	h.SignalWorkflow(workflowID, signalName, "buzz")
}

func terminateWorkflow(h *common.SampleHelper, workflowID string) {
	h.CancelWorkflow(workflowID)
}

func query(h *common.SampleHelper) {
	h.QueryWorkflow("helloworld_293eec42-d15b-4138-b3de-4dd94b60ff86", "8521f93e-0aea-43bd-8e14-efa844957173", "__open_sessions")
}
