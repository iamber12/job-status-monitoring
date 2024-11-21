package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// JobStatus represents the current status of a job
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

// StatusResponse is the response returned by the server
type StatusResponse struct {
	Payload StatusResponsePayload `json:"payload"`
}

// Client is used to talk to the status server
type Client struct {
	baseURL     string
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
}

// NewClient creates a new Client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:     baseURL,
		maxAttempts: 10,
		baseDelay:   100 * time.Millisecond,
		maxDelay:    10 * time.Second,
	}
}

func (c *Client) SetMaxAttempts(maxAttempts int) {
	c.maxAttempts = maxAttempts
}

func (c *Client) SetBaseDelay(baseDelay time.Duration) {
	c.baseDelay = baseDelay
}

func (c *Client) SetMaxDelay(maxDelay time.Duration) {
	c.maxDelay = maxDelay
}

// CreateJob sends a POST request to create a job and returns the job ID
func (c *Client) CreateJob(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned error: %d", resp.StatusCode)
	}

	var createResp CreateJobResponse
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse create job response: %w", err)
	}

	if createResp.Payload.JobID == "" {
		return "", fmt.Errorf("job ID is missing in the response")
	}

	return createResp.Payload.JobID, nil
}

func (c *Client) WaitForJob(ctx context.Context, jobID string) (*StatusResponsePayload, error) {
	return c.waitForJob(ctx, jobID, nil)
}

func (c *Client) WaitForJobWithUpdates(ctx context.Context, jobID string, statusUpdate chan<- string) (*StatusResponsePayload, error) {
	return c.waitForJob(ctx, jobID, statusUpdate)
}

// WaitForJob checks the job status and waits until it's completed or fails
func (c *Client) waitForJob(ctx context.Context, jobID string, statusUpdate chan<- string) (*StatusResponsePayload, error) {
	for attempt := 0; attempt < c.maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		status, _ := c.getStatus(jobID)

		switch status.Payload.Result {
		case StatusCompleted:
			return &status.Payload, nil
		case StatusError:
			return &status.Payload, errors.New("job failed")
		case StatusPending:
			if statusUpdate != nil {
				select {
				case statusUpdate <- fmt.Sprintf("Attempt %d: job is pending", attempt+1):
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
		default:
			return nil, errors.New("unexpected error occurred")
		}

		delay := calculateBackoff(attempt, c.baseDelay, c.maxDelay)
		time.Sleep(delay)
	}

	return nil, errors.New("job did not complete after maximum retries")
}

// getStatus makes an HTTP request to check the job status
func (c *Client) getStatus(jobID string) (*StatusResponse, error) {
	url := fmt.Sprintf("%s/status/%s", c.baseURL, jobID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var statusResp StatusResponse
	err = json.NewDecoder(resp.Body).Decode(&statusResp)

	if resp.StatusCode != http.StatusOK {
		return &statusResp, fmt.Errorf("server returned error: %d", resp.StatusCode)
	}

	if err != nil {
		return &statusResp, fmt.Errorf("failed to parse response: %w", err)
	}

	return &statusResp, nil
}

// calculateBackoff calculates an exponential backoff with jitter
func calculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	backoff := baseDelay * (1 << attempt)
	if backoff > maxDelay {
		backoff = maxDelay
	}
	// Add jitter to spread out retries
	jitter := time.Duration(rand.Int63n(int64(backoff) / 2))
	return backoff + jitter
}
