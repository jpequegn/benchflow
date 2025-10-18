package executor

import (
	"context"
	"time"

	"github.com/jpequegn/benchflow/internal/parser"
)

// BenchmarkConfig represents a single benchmark configuration
type BenchmarkConfig struct {
	Name     string        // Benchmark name
	Language string        // Language (rust, python, go)
	Command  string        // Command to execute
	WorkDir  string        // Working directory for execution
	Timeout  time.Duration // Execution timeout (0 = no timeout)
}

// ExecutionConfig represents executor configuration
type ExecutionConfig struct {
	Parallel int  // Number of parallel executions
	Retry    int  // Number of retries on failure
	FailFast bool // Stop on first failure
}

// ExecutionResult represents the result of executing a benchmark
type ExecutionResult struct {
	Config    *BenchmarkConfig       // Original config
	Suite     *parser.BenchmarkSuite // Parsed results
	Error     error                  // Execution or parsing error
	Duration  time.Duration          // Total execution time
	Attempts  int                    // Number of attempts made
	StartTime time.Time              // Start timestamp
	EndTime   time.Time              // End timestamp
}

// ProgressEvent represents a progress update during execution
type ProgressEvent struct {
	Type      EventType        // Event type
	Config    *BenchmarkConfig // Benchmark config
	Result    *ExecutionResult // Result (if completed)
	Error     error            // Error (if failed)
	Message   string           // Human-readable message
	Timestamp time.Time        // Event timestamp
}

// EventType represents the type of progress event
type EventType int

const (
	EventStarted   EventType = iota // Benchmark execution started
	EventRetrying                   // Retrying after failure
	EventCompleted                  // Benchmark completed successfully
	EventFailed                     // Benchmark failed permanently
	EventCancelled                  // Benchmark cancelled
)

// String returns string representation of EventType
func (e EventType) String() string {
	switch e {
	case EventStarted:
		return "started"
	case EventRetrying:
		return "retrying"
	case EventCompleted:
		return "completed"
	case EventFailed:
		return "failed"
	case EventCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Executor defines the interface for benchmark execution
type Executor interface {
	// Execute runs a single benchmark and returns the result
	Execute(ctx context.Context, config *BenchmarkConfig, parserRegistry ParserRegistry) (*ExecutionResult, error)

	// ExecuteBatch runs multiple benchmarks concurrently
	ExecuteBatch(ctx context.Context, configs []*BenchmarkConfig, execConfig *ExecutionConfig, parserRegistry ParserRegistry) ([]*ExecutionResult, error)
}

// ParserRegistry provides parsers for different languages
type ParserRegistry interface {
	// GetParser returns a parser for the specified language
	GetParser(language string) (parser.Parser, error)

	// RegisterParser registers a parser for a language
	RegisterParser(language string, parser parser.Parser)
}

// ProgressHandler is called for progress updates during batch execution
type ProgressHandler func(event *ProgressEvent)
