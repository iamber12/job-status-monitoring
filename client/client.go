package client

import (
	"context"
	"time"
)

type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusCompleted JobStatus = "completed"
	StatusError     JobStatus = "error"
)

type CreateJobPayload struct {
	JobID string `json:"jobID"`
}

type StatusResponsePayload struct {
	Result JobStatus `json:"status"`
}

type CreateJobResponse struct {
	Payload CreateJobPayload `json:"payload"`
}

type StatusResponse struct {
	Payload StatusResponsePayload `json:"payload"`
}

// Client interface defines the job-related API.
type Client interface {
	CreateJob(ctx context.Context) (string, error)
	WaitForJob(ctx context.Context, jobID string) (*StatusResponsePayload, error)
	WaitForJobWithUpdates(ctx context.Context, jobID string, statusUpdate chan<- string) (*StatusResponsePayload, error)
	SetMaxAttempts(maxAttempts int) error
	SetBaseDelay(baseDelay time.Duration) error
	SetMaxDelay(maxDelay time.Duration) error
	SetTimeout(timeout time.Duration) error
}
