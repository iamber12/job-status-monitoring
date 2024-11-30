# Job Status Monitoring

## Table of Contents
1. [Spinning up the Server](#spinning-up-the-server)
2. [Client Library](#client-library)
    - [Installation](#installation)
    - [Usage](#usage)
        - [Importing the Package](#importing-the-package)
        - [Creating a Client](#creating-a-client)
        - [Creating a Job](#creating-a-job)
        - [Waiting for Job Completion](#waiting-for-job-completion)
        - [Real-Time Status Updates](#real-time-status-updates)
3. [Configuration](#configuration)
4. [Running the Program](#running-the-program)
5. [Running Integration Tests](#running-integration-tests)
6. [Troubleshooting](#troubleshooting)
    - [Error: package your-project-name/job-status-monitoring/client is not in std (#GOROOT)](#error-package-your-project-namejob-status-monitoringclient-is-not-in-std-goroot)
7. [Algorithm Used](#algorithm-used)

## Spinning up the server
`server` is the package used to launch the Job Status server.

Fetch the package using:
```bash
git clone https://github.com/iamber12/job-status-monitoring.git
```

Run server
```bash
cd job-status-monitoring/server
go build -o ./bin/server ./cmd/main.go 
./bin/server serve
```
## Client library
`client` is the Go package for interacting with a Job Status server. It provides methods to create jobs, monitor their statuses, and configure retry/backoff mechanisms for job execution workflows.

## Installation

Fetch the package using:
```bash
git clone https://github.com/iamber12/job-status-monitoring.git
```

## Usage

### Importing the Package

#### Assuming the following project structure
```css
├── your-project-name
    ├── go.mod                  
    ├── main.go                 
    ├── job-status-monitoring
        ├── client
            ├── client.go       

```

#### Import the package
```go
import (
    "context"
    "fmt"
    "time"
    "your-project-name/job-status-monitoring/client"
)
```

### Creating a Client

To create a new client instance, initialize it with the server's base URL:

```go
jobClient := client.NewClient("http://localhost:8080")
```

### Creating a Job

Use `CreateJob` to create a new job and retrieve its Job ID:

```go
ctx := context.Background()
jobID, err := jobClient.CreateJob(ctx)
if err != nil {
    fmt.Printf("Error creating job: %v\n", err)
    return
}
fmt.Printf("Created job with ID: %s\n", jobID)
```

### Waiting for Job Completion

Poll the API until the job completes, fails, or times out:

```go
ctx := context.Background()
status, err := jobClient.WaitForJob(ctx, jobID)
if err != nil {
    fmt.Printf("Error waiting for job: %v\n", err)
    return
}
fmt.Printf("Job completed with status: %s\n", status.Result)
```

### Real-Time Status Updates

Use `WaitForJobWithUpdates` to receive status updates in real-time during polling:

```go
statusUpdate := make(chan string)
go func() {
    for update := range statusUpdate {
        fmt.Println(update)
    }
}()

ctx := context.Background()
status, err := jobClient.WaitForJobWithUpdates(ctx, jobID, statusUpdate)
if err != nil {
    fmt.Printf("Error waiting for job: %v\n", err)
    return
}
fmt.Printf("Job completed with status: %s\n", status.Result)
```

## Configuration

Customize the client with the following methods:

- `SetMaxAttempts(maxAttempts int) - Default: 10`
- `SetBaseDelay(baseDelay time.Duration) - Default: 100 ms`
- `SetMaxDelay(maxDelay time.Duration) - Default: 10 s`
- `SetTimeout(timeout time.Duration) - Default: 2 minutes`

Example:

```go
err := jobClient.SetMaxAttempts(5)
if err != nil {
    fmt.Printf("Error setting max attempts: %v\n", err)
}
```

## Running the program
```go
go run main.go
```

## Running Integration Tests

To execute the integration test, use the following command:

```bash
go test -v -count=1 ./test/integration
```

## Troubleshooting

### Error: `package your-project-name/job-status-monitoring/client is not in std (#GOROOT)`

If you encounter this error, it means Go is trying to resolve the package from the standard library (`#GOROOT`) instead of your local module. To fix this issue, make sure your `go.mod` file includes the following entries:

```go
require (
    your-project-name/job-status-monitoring v0.0.0
)

replace your-project-name/job-status-monitoring => ./job-status-monitoring
```

## Algorithm Used

The client library uses a **polling with exponential backoff** approach to monitor job statuses.

### Approach
1. **Polling**: Repeatedly queries the server for job status until it completes, fails, or times out.
2. **Exponential Backoff**: Wait times between requests increase exponentially with each retry, capped at a maximum delay. Random jitter is added to prevent server overload.
3. **Configurable Parameters**: Users can adjust retry attempts, delays, and timeouts to fit their specific needs.

This algorithm ensures efficient resource usage and minimizes unnecessary API calls.


---


