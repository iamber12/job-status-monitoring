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

type client struct {
	baseURL     string
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
	timeout     time.Duration
}

func NewClient(baseURL string) Client {
	return &client{
		baseURL:     baseURL,
		maxAttempts: 10,
		baseDelay:   100 * time.Millisecond,
		maxDelay:    10 * time.Second,
		timeout:     2 * time.Minute,
	}
}

func (c *client) CreateJob(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
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
		return "", errors.New("job ID is missing in the response")
	}

	return createResp.Payload.JobID, nil
}

func (c *client) WaitForJob(ctx context.Context, jobID string) (*StatusResponsePayload, error) {
	return c.waitForJob(ctx, jobID, nil)
}

func (c *client) WaitForJobWithUpdates(ctx context.Context, jobID string, statusUpdate chan<- string) (*StatusResponsePayload, error) {
	return c.waitForJob(ctx, jobID, statusUpdate)
}

func (c *client) waitForJob(ctx context.Context, jobID string, statusUpdate chan<- string) (*StatusResponsePayload, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	for attempt := 0; attempt < c.maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, errors.New("operation timed out: the job did not complete within the expected time frame")
			}
			return nil, ctx.Err()
		default:
		}

		status, err := c.getStatus(ctx, jobID)
		if status == nil && err != nil {
			return nil, err
		}

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
					if errors.Is(ctx.Err(), context.DeadlineExceeded) {
						return nil, errors.New("operation timed out: the job did not complete within the expected time frame")
					}
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

func (c *client) getStatus(ctx context.Context, jobID string) (*StatusResponse, error) {
	url := fmt.Sprintf("%s/status/%s", c.baseURL, jobID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var statusResp StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &statusResp, fmt.Errorf("server returned error: %d", resp.StatusCode)
	}

	return &statusResp, nil
}

func calculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	backoff := baseDelay * (1 << attempt)
	if backoff > maxDelay {
		backoff = maxDelay
	}

	jitter := time.Duration(rand.Int63n(int64(backoff) / 2))
	return backoff + jitter
}

func (c *client) SetMaxAttempts(maxAttempts int) error {
	if maxAttempts < 1 {
		return errors.New("max attempts must be greater than zero")
	}
	c.maxAttempts = maxAttempts
	return nil
}

func (c *client) SetBaseDelay(baseDelay time.Duration) error {
	if baseDelay <= 0 {
		return fmt.Errorf("baseDelay must be greater than 0, got: %v", baseDelay)
	}
	c.baseDelay = baseDelay
	return nil
}

func (c *client) SetMaxDelay(maxDelay time.Duration) error {
	if maxDelay <= 0 {
		return fmt.Errorf("maxDelay must be greater than 0, got: %v", maxDelay)
	}
	if maxDelay < c.baseDelay {
		return fmt.Errorf("maxDelay (%v) must be greater than or equal to baseDelay (%v)", maxDelay, c.baseDelay)
	}
	c.maxDelay = maxDelay
	return nil
}

func (c *client) SetTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0, got: %v", timeout)
	}
	c.timeout = timeout
	return nil
}
