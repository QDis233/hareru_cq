package hareru_cq

import (
	"fmt"
)

// ActionFailErr occurred when the action failed
type ActionFailErr struct {
	Message string
}

func (e *ActionFailErr) Error() string {
	return fmt.Sprintf("Action失败: %s", e.Message)
}

// NotAvailableErr occurred when the method is not available
type NotAvailableErr struct {
	Message string
}

func (e *NotAvailableErr) Error() string {
	return fmt.Sprintf("Method not available: %s", e.Message)
}

// AlreadyInitializedErr occurred when the instance is already initialized
type AlreadyInitializedErr struct {
	Message string
}

func (e *AlreadyInitializedErr) Error() string {
	return fmt.Sprintf("Already initialized: %s", e.Message)
}

type AlreadyRunningErr struct {
	Message string
}

func (e *AlreadyRunningErr) Error() string {
	return fmt.Sprintf("AlreadyRunningErr: %s", e.Message)
}
