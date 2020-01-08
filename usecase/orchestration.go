package usecase

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	letters                         = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CreateOrderGroupApplicationName = "CreateOrderGroup"
)

// Register workflows
func init() {
	activity.Register(OrchestrationWorkflow)
}

// OrchestrationWorkflow simulate saga orchestration workflow
func OrchestrationWorkflow(ctx workflow.Context, packageSerial, voucherSerial string) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 5,
		StartToCloseTimeout:    time.Minute * 5,
		HeartbeatTimeout:       time.Second * 20,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	var currentState string
	err := workflow.SetQueryHandler(ctx, "current_state", func() (string, error) {
		return currentState, nil
	})
	if err != nil {
		currentState = "failed to register query handler"
		return err
	}

	// Launch activity validate product package
	currentState = "VALIDATE_PRODUCT_PACKAGE"
	SimulateProcessDelay(ctx)
	err = workflow.ExecuteActivity(ctx, ValidateProductPackage, packageSerial).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	// Launch activity validate voucher
	currentState = "VALIDATE_VOUCHER"
	SimulateProcessDelay(ctx)
	err = workflow.ExecuteActivity(ctx, ValidateVoucher, voucherSerial).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	// Launch activity create order
	currentState = "CREATE_ORDER"
	SimulateProcessDelay(ctx)
	var orderSerial string
	err = workflow.ExecuteActivity(ctx, CreateOrder, packageSerial, voucherSerial).Get(ctx, &orderSerial)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	// Launch activity create payment
	currentState = "CREATE_PAYMENT"
	SimulateProcessDelay(ctx)
	err = workflow.ExecuteActivity(ctx, CreatePayment, orderSerial).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		err = workflow.ExecuteActivity(ctx, RollbackOrder, orderSerial).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	currentState = "ORDER_CREATED"

	return nil
}

// RandStringBytes generate random string with n characters
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// RandBool generate random bool
func RandBool() bool {
	return rand.Intn(10) != 0
}

// RandInt generate random int in range
func RandInt(min, max int) int {
	return rand.Intn(max-min) + min
}

// SimulateProcessDelay simulate processing somthing
func SimulateProcessDelay(ctx workflow.Context) error {
	err := workflow.Sleep(ctx, time.Millisecond*time.Duration(RandInt(700, 2000)))
	return err
}

// ValidateProductPackage simulate validate product API
func ValidateProductPackage(ctx context.Context, packageSerial string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Validating package serial...")
	return nil
}

// ValidateVoucher simulate validate voucher API
func ValidateVoucher(ctx context.Context, voucherSerial string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Validating voucher ...")
	return nil
}

// CreateOrder simulate create order API
func CreateOrder(ctx context.Context, packageSerial, voucherSerial string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating order ...")
	orderSerial := "INV-" + RandStringBytes(8)
	logger.Sugar().Infof("Order created, order serial %s", orderSerial)
	return orderSerial, nil
}

// CreatePayment simulate create payment API
func CreatePayment(ctx context.Context, orderSerial string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating payment ...")
	paymentSerial := "GW-" + RandStringBytes(8)

	// simulate error occur
	simulatedError := RandBool()
	if simulatedError {
		return "", errors.New("Something bad happen")
	}
	return paymentSerial, nil
}

// RollbackOrder simulate rollback order
func RollbackOrder(ctx context.Context, orderSerial string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Compensation order ...")
	logger.Info("Rollback order serial {orderSerial}", zap.String("orderSerial", orderSerial))
	return nil
}
