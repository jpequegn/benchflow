// Package executor provides concurrent benchmark execution with retry logic and progress tracking.
//
// # Overview
//
// The executor package implements a worker pool pattern for running benchmarks in parallel
// across multiple languages. It handles:
//
//   - Concurrent execution with configurable parallelism
//   - Context-based timeout and cancellation
//   - Automatic retry with backoff
//   - Progress tracking via events
//   - Graceful shutdown
//
// # Architecture
//
// The executor follows a worker pool pattern:
//
//	┌──────────────┐
//	│   Client     │
//	└──────┬───────┘
//	       │ ExecuteBatch()
//	       ▼
//	┌──────────────┐      ┌──────────────┐
//	│  Executor    │─────>│ Worker Pool  │
//	└──────────────┘      └──────┬───────┘
//	                             │
//	                    ┌────────┼────────┐
//	                    ▼        ▼        ▼
//	                 Worker   Worker   Worker
//	                    │        │        │
//	                    └────────┼────────┘
//	                             ▼
//	                      ┌──────────────┐
//	                      │   Results    │
//	                      └──────────────┘
//
// # Usage
//
// Basic single benchmark execution:
//
//	executor := executor.NewExecutor(nil)
//	registry := executor.NewParserRegistry()
//	registry.RegisterParser("rust", parser.NewRustParser())
//
//	config := &executor.BenchmarkConfig{
//	    Name:     "rust-sort",
//	    Language: "rust",
//	    Command:  "cargo bench --bench sort",
//	    WorkDir:  "./rust-examples",
//	    Timeout:  5 * time.Minute,
//	}
//
//	result, err := executor.Execute(ctx, config, registry)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Batch execution with progress tracking:
//
//	progressHandler := func(event *executor.ProgressEvent) {
//	    log.Printf("[%s] %s\n", event.Type, event.Message)
//	}
//
//	executor := executor.NewExecutor(progressHandler)
//	execConfig := &executor.ExecutionConfig{
//	    Parallel: 4,
//	    Retry:    3,
//	    FailFast: false,
//	}
//
//	results, err := executor.ExecuteBatch(ctx, configs, execConfig, registry)
//
// # Context and Cancellation
//
// The executor fully supports context-based cancellation. When a context is cancelled:
//
//   - Running benchmarks are terminated immediately
//   - Queued benchmarks are not started
//   - Workers exit gracefully
//   - Partial results are returned
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
//	defer cancel()
//
//	results, err := executor.ExecuteBatch(ctx, configs, execConfig, registry)
//	if err == context.DeadlineExceeded {
//	    log.Println("Batch execution timed out")
//	}
//
// # Retry Logic
//
// Failed benchmarks are automatically retried with exponential backoff:
//
//   - Initial retry delay: 1 second
//   - Maximum retries: configurable via ExecutionConfig.Retry
//   - Context cancellation terminates retries immediately
//
// # Error Handling
//
// The executor distinguishes between:
//
//   - Execution errors (command failed, timeout)
//   - Parsing errors (invalid output format)
//   - Context errors (cancellation, deadline exceeded)
//
// All errors are captured in ExecutionResult.Error and can be inspected by the caller.
//
// # Progress Events
//
// Progress events provide real-time updates during batch execution:
//
//   - EventStarted: Benchmark execution began
//   - EventRetrying: Retrying after failure
//   - EventCompleted: Benchmark succeeded
//   - EventFailed: Benchmark failed after all retries
//   - EventCancelled: Benchmark cancelled by context
//
// # Thread Safety
//
// All executor methods are safe for concurrent use. The ParserRegistry is also
// thread-safe and can be accessed from multiple goroutines.
//
// # Performance
//
// The executor is optimized for parallel execution:
//
//   - Worker pool pattern minimizes goroutine overhead
//   - Bounded channels prevent memory exhaustion
//   - Context propagation enables fast cancellation
//   - No blocking operations in hot paths
//
// Typical performance:
//   - 100 benchmarks × 4 workers: ~60 seconds (vs 240 seconds sequential)
//   - Memory overhead: ~1 MB per worker + output buffers
//   - Context cancellation latency: <100ms
package executor
