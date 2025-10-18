package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// DefaultExecutor implements the Executor interface with concurrent execution support
type DefaultExecutor struct {
	progressHandler ProgressHandler
}

// NewExecutor creates a new executor instance
func NewExecutor(progressHandler ProgressHandler) *DefaultExecutor {
	return &DefaultExecutor{
		progressHandler: progressHandler,
	}
}

// Execute runs a single benchmark and returns the result
func (e *DefaultExecutor) Execute(ctx context.Context, config *BenchmarkConfig, registry ParserRegistry) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Config:    config,
		StartTime: time.Now(),
	}

	// Get parser for this language
	p, err := registry.GetParser(config.Language)
	if err != nil {
		result.Error = fmt.Errorf("parser not found: %w", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, result.Error
	}

	// Create context with timeout if specified
	execCtx := ctx
	var cancel context.CancelFunc
	if config.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Execute the benchmark command
	output, err := e.executeCommand(execCtx, config)
	if err != nil {
		result.Error = fmt.Errorf("execution failed: %w", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, result.Error
	}

	// Parse the output
	suite, err := p.Parse(output)
	if err != nil {
		result.Error = fmt.Errorf("parsing failed: %w", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, result.Error
	}

	result.Suite = suite
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// executeCommand executes the benchmark command and captures output
func (e *DefaultExecutor) executeCommand(ctx context.Context, config *BenchmarkConfig) ([]byte, error) {
	// Parse command string into command and args
	// For simplicity, we'll use sh -c to handle complex commands
	cmd := exec.CommandContext(ctx, "sh", "-c", config.Command)

	// Set working directory if specified
	if config.WorkDir != "" {
		cmd.Dir = config.WorkDir
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()
	if err != nil {
		// Include stderr in error message
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("%w: %s", err, stderr.String())
		}
		return nil, err
	}

	// Return stdout (benchmark output)
	return stdout.Bytes(), nil
}

// ExecuteBatch runs multiple benchmarks concurrently using a worker pool
func (e *DefaultExecutor) ExecuteBatch(
	ctx context.Context,
	configs []*BenchmarkConfig,
	execConfig *ExecutionConfig,
	registry ParserRegistry,
) ([]*ExecutionResult, error) {
	// Create channels for work distribution
	jobs := make(chan *BenchmarkConfig, len(configs))
	results := make(chan *ExecutionResult, len(configs))
	errors := make(chan error, len(configs))

	// Context for cancellation propagation
	batchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start worker pool
	var wg sync.WaitGroup
	numWorkers := execConfig.Parallel
	if numWorkers <= 0 {
		numWorkers = 1
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go e.worker(batchCtx, jobs, results, execConfig, registry, &wg)
	}

	// Send jobs to workers
	go func() {
		for _, config := range configs {
			select {
			case jobs <- config:
			case <-batchCtx.Done():
				return
			}
		}
		close(jobs)
	}()

	// Collect results in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Collect all results
	var allResults []*ExecutionResult
	var firstError error

	for result := range results {
		allResults = append(allResults, result)

		// Handle fail-fast mode
		if execConfig.FailFast && result.Error != nil {
			if firstError == nil {
				firstError = result.Error
			}
			cancel() // Cancel remaining work
		}
	}

	// Check for context cancellation
	if ctx.Err() != nil {
		return allResults, ctx.Err()
	}

	// Return first error if fail-fast was triggered
	if execConfig.FailFast && firstError != nil {
		return allResults, firstError
	}

	return allResults, nil
}

// worker processes benchmark jobs from the jobs channel
func (e *DefaultExecutor) worker(
	ctx context.Context,
	jobs <-chan *BenchmarkConfig,
	results chan<- *ExecutionResult,
	execConfig *ExecutionConfig,
	registry ParserRegistry,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for config := range jobs {
		select {
		case <-ctx.Done():
			// Context cancelled, send cancelled result
			result := &ExecutionResult{
				Config: config,
				Error:  ctx.Err(),
			}
			e.sendProgressEvent(EventCancelled, config, result, ctx.Err())
			results <- result
			return
		default:
			// Execute benchmark with retry logic
			result := e.executeWithRetry(ctx, config, execConfig.Retry, registry)
			results <- result
		}
	}
}

// executeWithRetry executes a benchmark with retry logic
func (e *DefaultExecutor) executeWithRetry(
	ctx context.Context,
	config *BenchmarkConfig,
	maxRetries int,
	registry ParserRegistry,
) *ExecutionResult {
	var lastResult *ExecutionResult
	attempts := 0

	// Send started event
	e.sendProgressEvent(EventStarted, config, nil, nil)

	for attempts <= maxRetries {
		attempts++

		// Execute benchmark
		result, err := e.Execute(ctx, config, registry)
		result.Attempts = attempts
		lastResult = result

		// Success
		if err == nil {
			e.sendProgressEvent(EventCompleted, config, result, nil)
			return result
		}

		// Check if context was cancelled
		if ctx.Err() != nil {
			result.Error = ctx.Err()
			e.sendProgressEvent(EventCancelled, config, result, ctx.Err())
			return result
		}

		// Retry if not last attempt
		if attempts <= maxRetries {
			e.sendProgressEvent(EventRetrying, config, result, err)
			// Small backoff before retry
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				result.Error = ctx.Err()
				e.sendProgressEvent(EventCancelled, config, result, ctx.Err())
				return result
			}
		}
	}

	// All retries exhausted
	e.sendProgressEvent(EventFailed, config, lastResult, lastResult.Error)
	return lastResult
}

// sendProgressEvent sends a progress event if handler is configured
func (e *DefaultExecutor) sendProgressEvent(eventType EventType, config *BenchmarkConfig, result *ExecutionResult, err error) {
	if e.progressHandler == nil {
		return
	}

	event := &ProgressEvent{
		Type:      eventType,
		Config:    config,
		Result:    result,
		Error:     err,
		Timestamp: time.Now(),
	}

	// Generate human-readable message
	switch eventType {
	case EventStarted:
		event.Message = fmt.Sprintf("Starting benchmark: %s", config.Name)
	case EventRetrying:
		event.Message = fmt.Sprintf("Retrying benchmark: %s (attempt %d)", config.Name, result.Attempts)
	case EventCompleted:
		event.Message = fmt.Sprintf("Completed benchmark: %s (%d results, %v)", config.Name, len(result.Suite.Results), result.Duration)
	case EventFailed:
		event.Message = fmt.Sprintf("Failed benchmark: %s after %d attempts: %v", config.Name, result.Attempts, err)
	case EventCancelled:
		event.Message = fmt.Sprintf("Cancelled benchmark: %s", config.Name)
	}

	e.progressHandler(event)
}
