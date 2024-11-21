package integration

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
	client "video-translation-status/client"
	"video-translation-status/server/cmd/serve"
)

const (
	COMPLETED       = "completed"
	ERROR           = "error"
	MAX_RETRY_ERROR = "job did not complete after maximum retries"
)

func printStatusUpdates(statusUpdate <-chan string) {
	for update := range statusUpdate {
		fmt.Println("Status update:", update)
	}
}

/*
*****Causing issues in Windows because the server is spun up as a child process,
which does not terminate when serverCmd.Process.Kill() is executed.****

	func setupTestServer(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		serverCmd := exec.CommandContext(ctx, "go", "run", "../../server/cmd/main.go", "serve")
		serverCmd.Stdout = log.Writer()
		serverCmd.Stderr = log.Writer()

		fmt.Println("Starting server...")
		if err := serverCmd.Start(); err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}

		// Ensure the server is terminated at the end of the test
		defer func() {
			cancel() // Cancel context to stop server if it's listening to the context.
			if err := serverCmd.Process.Kill(); err != nil {
				fmt.Printf("Failed to kill server process: %v\n", err)
			} else {
				fmt.Println("Server process killed successfully")
			}
			serverCmd.Wait()
		}()
	}
*/
func setupTestServer(t *testing.T) (*httptest.Server, client.Client) {
	t.Helper()

	router := serve.SetupRouter()
	ts := httptest.NewServer(router)

	clientObj := client.NewClient(ts.URL)

	return ts, clientObj
}

func TestIntegration(t *testing.T) {
	t.Run("Valid Response Test", func(t *testing.T) {
		fmt.Println("**** Starting Valid Response Test ****")
		ts, clientObj := setupTestServer(t)
		defer ts.Close()

		ctx := context.Background()
		fmt.Println("Creating a new job...")
		jobID, err := clientObj.CreateJob(ctx)
		if err != nil {
			t.Fatalf("Failed to create job: %v", err)
			return
		}
		fmt.Printf("Job created successfully! Job ID: %s\n", jobID)

		statusUpdate := make(chan string)
		go printStatusUpdates(statusUpdate)

		fmt.Println("Waiting for the job to complete...")
		resp, err := clientObj.WaitForJobWithUpdates(ctx, jobID, statusUpdate)

		if resp == nil {
			t.Fatalf("Failed to wait for job: %v", err)
			return
		}

		if assert.Condition(t, func() bool {
			return resp.Result == COMPLETED || resp.Result == ERROR
		}, "Unexpected job result. Expected 'completed' or 'error'") {
			fmt.Printf("Job finished with result: %s\n\n", resp.Result)
		}
	})

	t.Run("Max Attempts Error Test", func(t *testing.T) {
		fmt.Println("**** Starting Max Attempts Error Test ****")
		ts, clientObj := setupTestServer(t)
		defer ts.Close()

		ctx := context.Background()
		err := clientObj.SetMaxAttempts(1)
		if err != nil {
			t.Fatalf("Failed to set maximum retry attempts: %v", err)
			return
		}

		fmt.Println("Creating a new job...")
		jobID, err := clientObj.CreateJob(ctx)
		if err != nil {
			t.Fatalf("Failed to create job: %v", err)
			return
		}
		fmt.Printf("Job created successfully! Job ID: %s\n", jobID)

		statusUpdate := make(chan string)
		go printStatusUpdates(statusUpdate)

		fmt.Println("Waiting for the job to complete...")
		resp, err := clientObj.WaitForJobWithUpdates(ctx, jobID, statusUpdate)
		expectedError := errors.New(MAX_RETRY_ERROR)

		if resp == nil {
			if !assert.Equal(t, err.Error(), expectedError.Error(), "Unexpected error") {
				t.Fatalf("Failed to wait for job: %v", err)
				return
			} else {
				fmt.Printf("Error: %s\n", err.Error())
			}
		}
	})
}
