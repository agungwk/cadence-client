package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/agungwk/cadence-client/usecase"
	"github.com/pborman/uuid"
	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
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
			startWorkers(&h, ApplicationName)
		case "orchestration":
			startWorkers(&h, usecase.CreateOrderGroupApplicationName)
		}

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "trigger":

		switch workflow {
		case "simple":
			startWorkflow(&h, Workflow, "helloWorldGroup", "Your name")
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
	workflow.Register(Workflow)
	activity.Register(helloworldActivity)
	activity.Register(helloworldGuys)
	activity.Register(helloMoving)
	activity.Register(helloNoContext)

	workflow.Register(usecase.OrchestrationWorkflow)
	activity.Register(usecase.ValidateProductPackage)
	activity.Register(usecase.ValidateVoucher)
	activity.Register(usecase.CreateOrder)
	activity.Register(usecase.CreatePayment)
	activity.Register(usecase.RollbackOrder)
}

// Workflow workflow decider
func Workflow(ctx workflow.Context, name string) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 5,
		StartToCloseTimeout:    time.Minute * 5,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// register state handler
	currentState := "initial"

	// var signalVal string
	// signalChan := workflow.GetSignalChannel(ctx, signalName)
	// s := workflow.NewSelector(ctx)
	// s.AddReceive(signalChan, func(c workflow.Channel, more bool) {
	// 	c.Receive(ctx, &signalVal)
	// 	workflow.GetLogger(ctx).Info("Received signal!", zap.String("signal", signalName), zap.String("value", signalVal))
	// })
	// s.Select(ctx)

	// if len(signalVal) > 0 && signalVal != "buzz" {
	// 	return errors.New("signalVal")
	// }

	err := workflow.SetQueryHandler(ctx, "current_state", func() (string, error) {
		return currentState, nil
	})
	if err != nil {
		currentState = "failed to register query handler"
		return err
	}

	logger := workflow.GetLogger(ctx)
	logger.Info("helloworld workflow started")
	var helloworldResult string
	err = workflow.ExecuteActivity(ctx, helloworldActivity, name).Get(ctx, &helloworldResult)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	// activityInfo := activity.GetInfo(context.Background())
	// logger.Info("activity token %v", zap.ByteString("activityToken", activityInfo.TaskToken))

	err = workflow.ExecuteActivity(ctx, helloworldGuys, name, time.Now()).Get(ctx, &helloworldResult)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	currentState = "sleep"
	sleep(ctx, logger, 30)
	currentState = "wake up"

	strconv.Atoi(version)
	wfVersion := workflow.GetVersion(ctx, "helloChanges", workflow.DefaultVersion, 2)
	logger.Info(fmt.Sprintf("Current version is %v", wfVersion))
	// workflow.UpsertSearchAttributes(ctx, map[string]interface{}{"helloChangesVersion": wfVersion})

	sleep(ctx, logger, 30)

	helloNoContext(context.Background(), "Unkown", "Home")

	if wfVersion == workflow.DefaultVersion {
		logger.Info("Do nothing")
	} else if wfVersion == 1 {
		logger.Info("Moving to new location")
		var helloMovingResult string
		err = workflow.ExecuteActivity(ctx, helloMoving, name, "Work").Get(ctx, &helloMovingResult)
		if err != nil {
			logger.Error("Activity failed.", zap.Error(err))
			return err
		}
	} else if wfVersion == 2 {
		logger.Info("Moving to somewhere else")
		var helloMovingResult string
		err = workflow.ExecuteActivity(ctx, helloMoving, name, "Far far away.....").Get(ctx, &helloMovingResult)
		if err != nil {
			logger.Error("Activity failed.", zap.Error(err))
			return err
		}
	}

	logger.Info("Workflow completed.", zap.String("Result", helloworldResult))

	return nil
}

func helloworldActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("helloworld activity started")
	activityInfo := activity.GetInfo(ctx)
	logger.Info("activity token %v", zap.ByteString("activityToken", activityInfo.TaskToken))
	return "Hello " + name + "!", nil
}

func helloworldGuys(ctx context.Context, name string, curtime time.Time) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("helloworldGuys activity started")
	return "Hello " + name + " at " + curtime.Format("2006-01-02 15:04:05"), nil
}

func helloMoving(ctx context.Context, name, newLocation string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Moving ......")
	return "Hello " + name + ", moving to " + newLocation, nil
}

func helloNoContext(ctx context.Context, name, newLocation string) (string, error) {
	return "Hello " + name + ", moving to " + newLocation, nil
}

func query(h *common.SampleHelper) {
	h.QueryWorkflow("helloworld_293eec42-d15b-4138-b3de-4dd94b60ff86", "8521f93e-0aea-43bd-8e14-efa844957173", "__open_sessions")
}

func sleep(ctx workflow.Context, logger *zap.Logger, seconds int) {
	logger.Info(fmt.Sprintf("Sleeping for %v second", seconds))
	err := workflow.Sleep(ctx, time.Second*time.Duration(seconds))
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
	}
	logger.Info("WAKE UP!!!")
}

func sendSignal(h *common.SampleHelper, workflowID string) {
	h.SignalWorkflow(workflowID, signalName, "buzz")
}

func terminateWorkflow(h *common.SampleHelper, workflowID string) {
	h.CancelWorkflow(workflowID)
}
