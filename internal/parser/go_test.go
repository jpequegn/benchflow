package parser

import (
	"os"
	"testing"
	"time"
)

func TestGoParser_Language(t *testing.T) {
	parser := NewGoParser()
	if got := parser.Language(); got != "go" {
		t.Errorf("Language() = %v, want %v", got, "go")
	}
}

func TestGoParser_Parse_BasicBenchmarks(t *testing.T) {
	input := []byte(`goos: darwin
goarch: arm64
pkg: github.com/example/benchmarks
cpu: Apple M1

BenchmarkSort-8         1000000              1234 ns/op             512 B/op          10 allocs/op
BenchmarkSearch-8       5000000               234 ns/op               0 B/op           0 allocs/op
BenchmarkInsert-8        500000              3456 ns/op            1024 B/op          20 allocs/op

PASS
ok      github.com/example/benchmarks    2.456s`)

	parser := NewGoParser()
	suite, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if suite == nil {
		t.Fatal("Parse() returned nil suite")
	}

	// Check suite metadata
	if suite.Language != "go" {
		t.Errorf("Suite.Language = %v, want %v", suite.Language, "go")
	}

	// Check number of results
	if len(suite.Results) != 3 {
		t.Fatalf("len(Results) = %d, want %d", len(suite.Results), 3)
	}

	// Verify first benchmark
	first := suite.Results[0]
	if first.Name != "BenchmarkSort-8" {
		t.Errorf("Results[0].Name = %v, want %v", first.Name, "BenchmarkSort-8")
	}
	if first.Language != "go" {
		t.Errorf("Results[0].Language = %v, want %v", first.Language, "go")
	}
	if first.Time != 1234*time.Nanosecond {
		t.Errorf("Results[0].Time = %v, want %v", first.Time, 1234*time.Nanosecond)
	}
	if first.Iterations != 1000000 {
		t.Errorf("Results[0].Iterations = %d, want %d", first.Iterations, 1000000)
	}
	if first.Metadata["bytes_per_op"] != "512" {
		t.Errorf("Results[0].Metadata['bytes_per_op'] = %v, want %v", first.Metadata["bytes_per_op"], "512")
	}
	if first.Metadata["allocs_per_op"] != "10" {
		t.Errorf("Results[0].Metadata['allocs_per_op'] = %v, want %v", first.Metadata["allocs_per_op"], "10")
	}

	// Verify second benchmark (no memory allocation)
	second := suite.Results[1]
	if second.Name != "BenchmarkSearch-8" {
		t.Errorf("Results[1].Name = %v, want %v", second.Name, "BenchmarkSearch-8")
	}
	if second.Time != 234*time.Nanosecond {
		t.Errorf("Results[1].Time = %v, want %v", second.Time, 234*time.Nanosecond)
	}
	if second.Iterations != 5000000 {
		t.Errorf("Results[1].Iterations = %d, want %d", second.Iterations, 5000000)
	}
	// Should not have memory metadata for zero allocation
	if _, ok := second.Metadata["bytes_per_op"]; ok {
		t.Errorf("Results[1] should not have bytes_per_op for zero allocation")
	}
	if _, ok := second.Metadata["allocs_per_op"]; ok {
		t.Errorf("Results[1] should not have allocs_per_op for zero allocation")
	}
}

func TestGoParser_Parse_FromFile(t *testing.T) {
	data, err := os.ReadFile("../../testdata/go/testing_b_basic.txt")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewGoParser()
	suite, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if len(suite.Results) != 3 {
		t.Errorf("len(Results) = %d, want %d", len(suite.Results), 3)
	}
}

func TestGoParser_Parse_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantResults     int
		wantErr         bool
		wantName        string
		wantTime        time.Duration
		wantIterations  int64
		wantBytesPerOp  string
		wantAllocsPerOp string
	}{
		{
			name: "zero allocations",
			input: `goos: linux
goarch: amd64

BenchmarkZeroAllocs-8             1000000              1000 ns/op               0 B/op           0 allocs/op

PASS`,
			wantResults:     1,
			wantErr:         false,
			wantName:        "BenchmarkZeroAllocs-8",
			wantTime:        1000 * time.Nanosecond,
			wantIterations:  1000000,
			wantBytesPerOp:  "",
			wantAllocsPerOp: "",
		},
		{
			name: "large numbers",
			input: `goos: linux
goarch: amd64

BenchmarkLargeNumbers-8              100            123456789 ns/op        1048576 B/op         512 allocs/op

PASS`,
			wantResults:     1,
			wantErr:         false,
			wantName:        "BenchmarkLargeNumbers-8",
			wantTime:        123456789 * time.Nanosecond,
			wantIterations:  100,
			wantBytesPerOp:  "1048576",
			wantAllocsPerOp: "512",
		},
		{
			name: "float time value",
			input: `BenchmarkFastOp-8               10000000                10.5 ns/op              0 B/op           0 allocs/op

PASS`,
			wantResults:    1,
			wantErr:        false,
			wantName:       "BenchmarkFastOp-8",
			wantTime:       10 * time.Nanosecond, // truncates to int
			wantIterations: 10000000,
		},
		{
			name: "no memory metrics",
			input: `goos: windows
goarch: amd64

BenchmarkAdd-8                  1000000000                1.5 ns/op

PASS`,
			wantResults:    1,
			wantErr:        false,
			wantName:       "BenchmarkAdd-8",
			wantTime:       1 * time.Nanosecond, // truncates to int
			wantIterations: 1000000000,
		},
		{
			name:        "empty input",
			input:       "",
			wantResults: 0,
			wantErr:     true,
		},
		{
			name: "no benchmarks found",
			input: `goos: linux
goarch: amd64

PASS
ok      github.com/test    1.234s`,
			wantResults: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewGoParser()
			suite, err := parser.Parse([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(suite.Results) != tt.wantResults {
				t.Errorf("len(Results) = %d, want %d", len(suite.Results), tt.wantResults)
			}

			if tt.wantResults > 0 {
				result := suite.Results[0]
				if result.Name != tt.wantName {
					t.Errorf("Results[0].Name = %v, want %v", result.Name, tt.wantName)
				}
				if result.Time != tt.wantTime {
					t.Errorf("Results[0].Time = %v, want %v", result.Time, tt.wantTime)
				}
				if result.Iterations != tt.wantIterations {
					t.Errorf("Results[0].Iterations = %d, want %d", result.Iterations, tt.wantIterations)
				}
				if tt.wantBytesPerOp != "" {
					if result.Metadata["bytes_per_op"] != tt.wantBytesPerOp {
						t.Errorf("Results[0].Metadata['bytes_per_op'] = %v, want %v",
							result.Metadata["bytes_per_op"], tt.wantBytesPerOp)
					}
				}
				if tt.wantAllocsPerOp != "" {
					if result.Metadata["allocs_per_op"] != tt.wantAllocsPerOp {
						t.Errorf("Results[0].Metadata['allocs_per_op'] = %v, want %v",
							result.Metadata["allocs_per_op"], tt.wantAllocsPerOp)
					}
				}
			}
		})
	}
}

func TestGoParser_Parse_WithWarnings(t *testing.T) {
	data, err := os.ReadFile("../../testdata/go/testing_b_with_warnings.txt")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewGoParser()
	suite, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Should parse benchmarks despite debug output
	if len(suite.Results) != 3 {
		t.Errorf("len(Results) = %d, want %d", len(suite.Results), 3)
	}

	// Verify specific benchmark
	for _, result := range suite.Results {
		if result.Name == "BenchmarkQuickSort-8" {
			if result.Time != 1500*time.Nanosecond {
				t.Errorf("BenchmarkQuickSort-8 Time = %v, want %v", result.Time, 1500*time.Nanosecond)
			}
			if result.Iterations != 1000000 {
				t.Errorf("BenchmarkQuickSort-8 Iterations = %d, want %d", result.Iterations, 1000000)
			}
		}
	}
}

func TestGoParser_Parse_NoMemory(t *testing.T) {
	data, err := os.ReadFile("../../testdata/go/testing_b_no_memory.txt")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewGoParser()
	suite, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Should parse benchmarks without memory metrics
	if len(suite.Results) != 3 {
		t.Errorf("len(Results) = %d, want %d", len(suite.Results), 3)
	}

	// Verify first benchmark has no memory metadata
	first := suite.Results[0]
	if first.Name != "BenchmarkAdd-8" {
		t.Errorf("Results[0].Name = %v, want %v", first.Name, "BenchmarkAdd-8")
	}
	if _, ok := first.Metadata["bytes_per_op"]; ok {
		t.Errorf("Results[0] should not have bytes_per_op when not reported")
	}
	if _, ok := first.Metadata["allocs_per_op"]; ok {
		t.Errorf("Results[0] should not have allocs_per_op when not reported")
	}
}

func TestGoParser_Parse_EdgeCasesFromFile(t *testing.T) {
	data, err := os.ReadFile("../../testdata/go/testing_b_edge_cases.txt")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewGoParser()
	suite, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if len(suite.Results) != 4 {
		t.Errorf("len(Results) = %d, want %d", len(suite.Results), 4)
	}

	// Check for expected benchmarks
	names := make(map[string]*BenchmarkResult)
	for _, result := range suite.Results {
		names[result.Name] = result
	}

	// Check zero allocs case
	if zeroAllocs, ok := names["BenchmarkZeroAllocs-8"]; ok {
		if zeroAllocs.Time != 1000*time.Nanosecond {
			t.Errorf("ZeroAllocs time = %v, want %v", zeroAllocs.Time, 1000*time.Nanosecond)
		}
	} else {
		t.Error("BenchmarkZeroAllocs-8 not found")
	}

	// Check large numbers case
	if large, ok := names["BenchmarkLargeNumbers-8"]; ok {
		if large.Time != 123456789*time.Nanosecond {
			t.Errorf("Large time = %v, want %v", large.Time, 123456789*time.Nanosecond)
		}
		if large.Iterations != 100 {
			t.Errorf("Large iterations = %d, want %d", large.Iterations, 100)
		}
	} else {
		t.Error("BenchmarkLargeNumbers-8 not found")
	}

	// Check float time case
	if fast, ok := names["BenchmarkFastOp-8"]; ok {
		if fast.Iterations != 10000000 {
			t.Errorf("Fast iterations = %d, want %d", fast.Iterations, 10000000)
		}
	} else {
		t.Error("BenchmarkFastOp-8 not found")
	}
}

func TestGoParser_Parse_SkipsDebugOutput(t *testing.T) {
	input := []byte(`goos: darwin
goarch: arm64
pkg: github.com/example/benchmarks
cpu: Apple M1

BenchmarkSuccess-8            1000000              1234 ns/op             512 B/op          10 allocs/op
--- BENCH: BenchmarkDebugOutput-8
    bench_test.go:42: debug message
BenchmarkAnother-8            2000000              2000 ns/op             256 B/op           5 allocs/op

PASS
ok      github.com/example/benchmarks    2.456s`)

	parser := NewGoParser()
	suite, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Should only parse actual benchmark lines, skip debug output
	if len(suite.Results) != 2 {
		t.Errorf("len(Results) = %d, want %d (should skip debug output)", len(suite.Results), 2)
	}

	if suite.Results[0].Name != "BenchmarkSuccess-8" {
		t.Errorf("Results[0].Name = %v, want BenchmarkSuccess-8", suite.Results[0].Name)
	}
	if suite.Results[1].Name != "BenchmarkAnother-8" {
		t.Errorf("Results[1].Name = %v, want BenchmarkAnother-8", suite.Results[1].Name)
	}
}

func TestGoParser_Parse_HandlesVariousNames(t *testing.T) {
	input := []byte(`BenchmarkSimpleName-1       1000000              1000 ns/op
BenchmarkWith_Underscore-16 2000000              2000 ns/op
Benchmark-32                3000000              3000 ns/op
BenchmarkCamelCase-8        4000000              4000 ns/op

PASS`)

	parser := NewGoParser()
	suite, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if len(suite.Results) != 4 {
		t.Errorf("len(Results) = %d, want %d", len(suite.Results), 4)
	}

	expectedNames := []string{
		"BenchmarkSimpleName-1",
		"BenchmarkWith_Underscore-16",
		"Benchmark-32",
		"BenchmarkCamelCase-8",
	}

	for i, expectedName := range expectedNames {
		if suite.Results[i].Name != expectedName {
			t.Errorf("Results[%d].Name = %v, want %v", i, suite.Results[i].Name, expectedName)
		}
	}
}
