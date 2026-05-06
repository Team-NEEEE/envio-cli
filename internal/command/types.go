package command

import (
	"context"
	"fmt"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusWarning Status = "warning"
	StatusError   Status = "error"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Step struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Status    Status `json:"status"`
	Detail    string `json:"detail,omitempty"`
	DetailKey string `json:"-"`
}

type StepUpdate struct {
	ID        string
	Status    Status
	Detail    string
	DetailKey string
}

type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

type SummaryItem struct {
	Label     string `json:"label"`
	Value     string `json:"value"`
	Sensitive bool   `json:"-"`
}

type Result struct {
	Title    string        `json:"title"`
	Summary  []SummaryItem `json:"summary"`
	Warnings []Warning     `json:"warnings,omitempty"`
}

type AppError struct {
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Hint     string   `json:"hint,omitempty"`
	ExitCode int      `json:"exitCode"`
	Severity Severity `json:"severity"`
}

func NewAppError(code, message, hint string, exitCode int, severity Severity) *AppError {
	if exitCode == 0 {
		exitCode = 1
	}
	if severity == "" {
		severity = SeverityError
	}
	return &AppError{
		Code:     code,
		Message:  message,
		Hint:     hint,
		ExitCode: exitCode,
		Severity: severity,
	}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) CodeOrDefault() string {
	if e == nil || e.Code == "" {
		return "UNKNOWN_ERROR"
	}
	return e.Code
}

type Reporter interface {
	UpdateStep(StepUpdate)
}

type Command interface {
	Name() string
	Steps() []Step
	Run(ctx context.Context, reporter Reporter) (Result, *AppError)
}
