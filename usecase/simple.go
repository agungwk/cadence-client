package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	HelloWorldApplicationName = "helloWorldGroup"
)

var version string

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
	err = workflow.ExecuteActivity(ctx, HelloworldActivity, name).Get(ctx, &helloworldResult)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	// activityInfo := activity.GetInfo(context.Background())
	// logger.Info("activity token %v", zap.ByteString("activityToken", activityInfo.TaskToken))

	err = workflow.ExecuteActivity(ctx, HelloworldGuys, name, time.Now()).Get(ctx, &helloworldResult)
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

	HelloNoContext(context.Background(), "Unkown", "Home")

	if wfVersion == workflow.DefaultVersion {
		logger.Info("Do nothing")
	} else if wfVersion == 1 {
		logger.Info("Moving to new location")
		var helloMovingResult string
		err = workflow.ExecuteActivity(ctx, HelloMoving, name, "Work").Get(ctx, &helloMovingResult)
		if err != nil {
			logger.Error("Activity failed.", zap.Error(err))
			return err
		}
	} else if wfVersion == 2 {
		logger.Info("Moving to somewhere else")
		var helloMovingResult string
		err = workflow.ExecuteActivity(ctx, HelloMoving, name, "Far far away.....").Get(ctx, &helloMovingResult)
		if err != nil {
			logger.Error("Activity failed.", zap.Error(err))
			return err
		}
	}

	logger.Info("Workflow completed.", zap.String("Result", helloworldResult))

	return nil
}

func HelloworldActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("helloworld activity started")
	activityInfo := activity.GetInfo(ctx)
	logger.Info("activity token %v", zap.ByteString("activityToken", activityInfo.TaskToken))
	return "Hello " + name + "!", nil
}

func HelloworldGuys(ctx context.Context, name string, curtime time.Time) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("helloworldGuys activity started")
	return "Hello " + name + " at " + curtime.Format("2006-01-02 15:04:05"), nil
}

func HelloMoving(ctx context.Context, name, newLocation string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Moving ......")
	return "Hello " + name + ", moving to " + newLocation, nil
}

func HelloNoContext(ctx context.Context, name, newLocation string) (string, error) {
	return "Hello " + name + ", moving to " + newLocation, nil
}

func sleep(ctx workflow.Context, logger *zap.Logger, seconds int) {
	logger.Info(fmt.Sprintf("Sleeping for %v second", seconds))
	err := workflow.Sleep(ctx, time.Second*time.Duration(seconds))
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
	}
	logger.Info("WAKE UP!!!")
}
