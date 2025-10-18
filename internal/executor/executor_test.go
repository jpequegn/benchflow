package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jpequegn/benchflow/internal/parser"
)

func TestExecutor_Execute_Success(t *testing.T) {
	executor := NewExecutor(nil)
	registry := setupTestRegistry()

	config := &BenchmarkConfig{
		Name:     "test-echo",
		Language: "rust",
		Command:  "echo 'test bench_test ... bench:   1,234 ns/iter (+/- 56)'",
		Timeout:  5 * time.Second,
	}

	result, err := executor.Execute(context.Background(), config, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("result has error: %v", result.Error)
	}

	if result.Suite == nil {
		t.Fatal("expected suite, got nil")
	}

	if len(result.Suite.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(result.Suite.Results))
	}

	if result.Duration == 0 {
		t.Error("expected non-zero duration")
	}
}

func TestExecutor_Execute_CommandFailure(t *testing.T) {
	executor := NewExecutor(nil)
	registry := setupTestRegistry()

	config := &BenchmarkConfig{
		Name:     "test-fail",
		Language: "rust",
		Command:  "exit 1",
		Timeout:  5 * time.Second,
	}

	_, err := executor.Execute(context.Background(), config, registry)
	if err == nil {
		t.Fatal("expected error for failed command")
	}
}

func TestExecutor_Execute_Timeout(t *testing.T) {
	executor := NewExecutor(nil)
	registry := setupTestRegistry()

	config := &BenchmarkConfig{
		Name:     "test-timeout",
		Language: "rust",
		Command:  "sleep 5",
		Timeout:  100 * time.Millisecond,
	}

	_, err := executor.Execute(context.Background(), config, registry)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !strings.Contains(err.Error(), "context deadline exceeded") &&
		!strings.Contains(err.Error(), "killed") {
		t.Errorf("expected timeout/killed error, got: %v", err)
	}
}

func TestExecutor_Execute_ContextCancellation(t *testing.T) {
	executor := NewExecutor(nil)
	registry := setupTestRegistry()

	config := &BenchmarkConfig{
		Name:     "test-cancel",
		Language: "rust",
		Command:  "sleep 10",
		Timeout:  0,
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := executor.Execute(ctx, config, registry)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestExecutor_Execute_ParserNotFound(t *testing.T) {
	executor := NewExecutor(nil)
	registry := NewParserRegistry() // Empty registry

	config := &BenchmarkConfig{
		Name:     "test-noparser",
		Language: "python",
		Command:  "echo 'test'",
		Timeout:  5 * time.Second,
	}

	_, err := executor.Execute(context.Background(), config, registry)
	if err == nil {
		t.Fatal("expected parser not found error")
	}

	if !strings.Contains(err.Error(), "parser not found") {
		t.Errorf("expected 'parser not found' error, got: %v", err)
	}
}

func TestExecutor_Execute_ParsingFailure(t *testing.T) {
	executor := NewExecutor(nil)
	registry := setupTestRegistry()

	config := &BenchmarkConfig{
		Name:     "test-parserrr",
		Language: "rust",
		Command:  "echo 'invalid benchmark output'",
		Timeout:  5 * time.Second,
	}

	_, err := executor.Execute(context.Background(), config, registry)
	if err == nil {
		t.Fatal("expected parsing error")
	}

	if !strings.Contains(err.Error(), "parsing failed") {
		t.Errorf("expected 'parsing failed' error, got: %v", err)
	}
}

func TestExecutor_ExecuteBatch_Success(t *testing.T) {
	var events []*ProgressEvent
	var mu sync.Mutex

	progressHandler := func(event *ProgressEvent) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, event)
	}

	executor := NewExecutor(progressHandler)
	registry := setupTestRegistry()

	configs := []*BenchmarkConfig{
		{
			Name:     "test-1",
			Language: "rust",
			Command:  "echo 'test bench_1 ... bench:   100 ns/iter (+/- 10)'",
			Timeout:  5 * time.Second,
		},
		{
			Name:     "test-2",
			Language: "rust",
			Command:  "echo 'test bench_2 ... bench:   200 ns/iter (+/- 20)'",
			Timeout:  5 * time.Second,
		},
		{
			Name:     "test-3",
			Language: "rust",
			Command:  "echo 'test bench_3 ... bench:   300 ns/iter (+/- 30)'",
			Timeout:  5 * time.Second,
		},
	}

	execConfig := &ExecutionConfig{
		Parallel: 2,
		Retry:    1,
		FailFast: false,
	}

	results, err := executor.ExecuteBatch(context.Background(), configs, execConfig, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		if result.Error != nil {
			t.Errorf("result %d has error: %v", i, result.Error)
		}
		if result.Suite == nil {
			t.Errorf("result %d has nil suite", i)
		}
	}

	// Check progress events
	mu.Lock()
	defer mu.Unlock()

	if len(events) == 0 {
		t.Error("expected progress events")
	}

	// Should have started and completed events for each benchmark
	startedCount := 0
	completedCount := 0
	for _, event := range events {
		if event.Type == EventStarted {
			startedCount++
		}
		if event.Type == EventCompleted {
			completedCount++
		}
	}

	if startedCount != 3 {
		t.Errorf("expected 3 started events, got %d", startedCount)
	}
	if completedCount != 3 {
		t.Errorf("expected 3 completed events, got %d", completedCount)
	}
}

func TestExecutor_ExecuteBatch_WithRetry(t *testing.T) {
	var events []*ProgressEvent
	var mu sync.Mutex

	progressHandler := func(event *ProgressEvent) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, event)
	}

	executor := NewExecutor(progressHandler)
	registry := setupTestRegistry()

	// First command fails, should retry
	configs := []*BenchmarkConfig{
		{
			Name:     "test-retry",
			Language: "rust",
			Command:  "exit 1",
			Timeout:  5 * time.Second,
		},
	}

	execConfig := &ExecutionConfig{
		Parallel: 1,
		Retry:    2,
		FailFast: false,
	}

	results, err := executor.ExecuteBatch(context.Background(), configs, execConfig, registry)
	if err != nil {
		t.Fatalf("unexpected batch error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Error == nil {
		t.Error("expected result to have error")
	}

	if result.Attempts != 3 {
		t.Errorf("expected 3 attempts (1 initial + 2 retries), got %d", result.Attempts)
	}

	// Check for retry events
	mu.Lock()
	defer mu.Unlock()

	retryCount := 0
	for _, event := range events {
		if event.Type == EventRetrying {
			retryCount++
		}
	}

	if retryCount != 2 {
		t.Errorf("expected 2 retry events, got %d", retryCount)
	}
}

func TestExecutor_ExecuteBatch_FailFast(t *testing.T) {
	executor := NewExecutor(nil)
	registry := setupTestRegistry()

	configs := []*BenchmarkConfig{
		{
			Name:     "test-fail",
			Language: "rust",
			Command:  "exit 1",
			Timeout:  5 * time.Second,
		},
		{
			Name:     "test-ok",
			Language: "rust",
			Command:  "echo 'test bench_ok ... bench:   100 ns/iter (+/- 10)'",
			Timeout:  5 * time.Second,
		},
	}

	execConfig := &ExecutionConfig{
		Parallel: 1,
		Retry:    0,
		FailFast: true,
	}

	results, err := executor.ExecuteBatch(context.Background(), configs, execConfig, registry)
	if err == nil {
		t.Fatal("expected error with FailFast=true")
	}

	// Should have at least one result (the failed one)
	if len(results) == 0 {
		t.Error("expected at least one result")
	}

	// First result should have error
	if results[0].Error == nil {
		t.Error("expected first result to have error")
	}
}

func TestExecutor_ExecuteBatch_ContextCancellation(t *testing.T) {
	executor := NewExecutor(nil)
	registry := setupTestRegistry()

	configs := []*BenchmarkConfig{
		{
			Name:     "test-1",
			Language: "rust",
			Command:  "sleep 5",
			Timeout:  0,
		},
		{
			Name:     "test-2",
			Language: "rust",
			Command:  "sleep 5",
			Timeout:  0,
		},
	}

	execConfig := &ExecutionConfig{
		Parallel: 1,
		Retry:    0,
		FailFast: false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	results, err := executor.ExecuteBatch(ctx, configs, execConfig, registry)

	// Should return context error
	if err == nil {
		t.Error("expected context error")
	}

	// Should have partial results
	if len(results) == 0 {
		t.Error("expected at least some results")
	}
}

func TestExecutor_ExecuteBatch_Parallel(t *testing.T) {
	executor := NewExecutor(nil)
	registry := setupTestRegistry()

	// Create 10 benchmarks that each take 100ms
	configs := make([]*BenchmarkConfig, 10)
	for i := 0; i < 10; i++ {
		configs[i] = &BenchmarkConfig{
			Name:     fmt.Sprintf("test-%d", i),
			Language: "rust",
			Command:  fmt.Sprintf("sleep 0.1 && echo 'test bench_%d ... bench:   100 ns/iter (+/- 10)'", i),
			Timeout:  5 * time.Second,
		}
	}

	execConfig := &ExecutionConfig{
		Parallel: 5, // 5 parallel workers
		Retry:    0,
		FailFast: false,
	}

	start := time.Now()
	results, err := executor.ExecuteBatch(context.Background(), configs, execConfig, registry)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("expected 10 results, got %d", len(results))
	}

	// With 5 workers and 10 tasks of 100ms each, should complete in ~200ms
	// Allow for overhead, so check < 500ms
	if duration > 500*time.Millisecond {
		t.Errorf("expected parallel execution to complete quickly, took %v", duration)
	}

	t.Logf("Parallel execution of 10 benchmarks with 5 workers took %v", duration)
}

func TestEventType_String(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventStarted, "started"},
		{EventRetrying, "retrying"},
		{EventCompleted, "completed"},
		{EventFailed, "failed"},
		{EventCancelled, "cancelled"},
		{EventType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.eventType.String(); got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// setupTestRegistry creates a registry with a Rust parser for testing
func setupTestRegistry() *DefaultParserRegistry {
	registry := NewParserRegistry()
	registry.RegisterParser("rust", parser.NewRustParser())
	return registry
}
