package parser

import (
	"os"
	"testing"
	"time"
)

func TestRustParser_Language(t *testing.T) {
	parser := NewRustParser()
	if got := parser.Language(); got != "rust" {
		t.Errorf("Language() = %v, want %v", got, "rust")
	}
}

func TestRustParser_Parse_BasicBencher(t *testing.T) {
	input := []byte(`running 5 tests
test bench_bubble_sort ... bench:   1,234 ns/iter (+/- 56)
test bench_quick_sort  ... bench:     567 ns/iter (+/- 23)
test bench_merge_sort  ... bench:     890 ns/iter (+/- 45)
test bench_heap_sort   ... bench:   1,100 ns/iter (+/- 78)
test bench_insertion   ... bench:   2,345 ns/iter (+/- 123)

test result: ok. 5 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out`)

	parser := NewRustParser()
	suite, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if suite == nil {
		t.Fatal("Parse() returned nil suite")
	}

	// Check suite metadata
	if suite.Language != "rust" {
		t.Errorf("Suite.Language = %v, want %v", suite.Language, "rust")
	}

	// Check number of results
	if len(suite.Results) != 5 {
		t.Fatalf("len(Results) = %d, want %d", len(suite.Results), 5)
	}

	// Verify first benchmark
	first := suite.Results[0]
	if first.Name != "bench_bubble_sort" {
		t.Errorf("Results[0].Name = %v, want %v", first.Name, "bench_bubble_sort")
	}
	if first.Time != 1234*time.Nanosecond {
		t.Errorf("Results[0].Time = %v, want %v", first.Time, 1234*time.Nanosecond)
	}
	if first.StdDev != 56*time.Nanosecond {
		t.Errorf("Results[0].StdDev = %v, want %v", first.StdDev, 56*time.Nanosecond)
	}
	if first.Language != "rust" {
		t.Errorf("Results[0].Language = %v, want %v", first.Language, "rust")
	}

	// Verify last benchmark
	last := suite.Results[4]
	if last.Name != "bench_insertion" {
		t.Errorf("Results[4].Name = %v, want %v", last.Name, "bench_insertion")
	}
	if last.Time != 2345*time.Nanosecond {
		t.Errorf("Results[4].Time = %v, want %v", last.Time, 2345*time.Nanosecond)
	}
}

func TestRustParser_Parse_WithWarnings(t *testing.T) {
	data, err := os.ReadFile("../../testdata/rust/cargo_bench_with_warnings.txt")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewRustParser()
	suite, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Should parse benchmarks despite warnings
	if len(suite.Results) != 3 {
		t.Errorf("len(Results) = %d, want %d", len(suite.Results), 3)
	}

	// Verify specific benchmark
	for _, result := range suite.Results {
		if result.Name == "bench_fast_operation" {
			if result.Time != 42*time.Nanosecond {
				t.Errorf("bench_fast_operation Time = %v, want %v", result.Time, 42*time.Nanosecond)
			}
			if result.StdDev != 3*time.Nanosecond {
				t.Errorf("bench_fast_operation StdDev = %v, want %v", result.StdDev, 3*time.Nanosecond)
			}
		}
	}
}

func TestRustParser_Parse_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantResults    int
		wantErr        bool
		checkFirstName string
		checkFirstTime time.Duration
	}{
		{
			name: "zero nanoseconds",
			input: `test bench_very_fast ... bench:       0 ns/iter (+/- 0)
test result: ok.`,
			wantResults:    1,
			wantErr:        false,
			checkFirstName: "bench_very_fast",
			checkFirstTime: 0,
		},
		{
			name: "large numbers",
			input: `test bench_slow ... bench:  12,345,678 ns/iter (+/- 987,654)
test result: ok.`,
			wantResults:    1,
			wantErr:        false,
			checkFirstName: "bench_slow",
			checkFirstTime: 12345678 * time.Nanosecond,
		},
		{
			name: "single digit",
			input: `test bench_fast ... bench:       5 ns/iter (+/- 1)
test result: ok.`,
			wantResults:    1,
			wantErr:        false,
			checkFirstName: "bench_fast",
			checkFirstTime: 5 * time.Nanosecond,
		},
		{
			name: "with underscores in name",
			input: `test bench_with_many_underscores ... bench:     100 ns/iter (+/- 10)
test result: ok.`,
			wantResults:    1,
			wantErr:        false,
			checkFirstName: "bench_with_many_underscores",
			checkFirstTime: 100 * time.Nanosecond,
		},
		{
			name:        "empty input",
			input:       "",
			wantResults: 0,
			wantErr:     true,
		},
		{
			name: "no benchmarks found",
			input: `running 0 tests
test result: ok. 0 passed`,
			wantResults: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewRustParser()
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

			if tt.wantResults > 0 && tt.checkFirstName != "" {
				if suite.Results[0].Name != tt.checkFirstName {
					t.Errorf("Results[0].Name = %v, want %v", suite.Results[0].Name, tt.checkFirstName)
				}
				if suite.Results[0].Time != tt.checkFirstTime {
					t.Errorf("Results[0].Time = %v, want %v", suite.Results[0].Time, tt.checkFirstTime)
				}
			}
		})
	}
}

func TestRustParser_Parse_FromFile(t *testing.T) {
	data, err := os.ReadFile("../../testdata/rust/cargo_bench_bencher.txt")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewRustParser()
	suite, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if len(suite.Results) != 5 {
		t.Errorf("len(Results) = %d, want %d", len(suite.Results), 5)
	}
}

func TestRustParser_Parse_SkipsFailedTests(t *testing.T) {
	input := []byte(`running 3 tests
test bench_success ... bench:     100 ns/iter (+/- 10)
test bench_failed ... FAILED
test bench_another ... bench:     200 ns/iter (+/- 20)

test result: FAILED. 2 passed; 1 failed`)

	parser := NewRustParser()
	suite, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Should only parse successful benchmarks
	if len(suite.Results) != 2 {
		t.Errorf("len(Results) = %d, want %d (failed test should be skipped)", len(suite.Results), 2)
	}
}

func TestRustParser_Parse_IgnoredTests(t *testing.T) {
	input := []byte(`running 2 tests
test bench_normal ... bench:     100 ns/iter (+/- 10)
test bench_ignored ... ignored

test result: ok. 1 passed; 0 failed; 1 ignored`)

	parser := NewRustParser()
	suite, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Should only parse non-ignored benchmarks
	if len(suite.Results) != 1 {
		t.Errorf("len(Results) = %d, want %d (ignored test should be skipped)", len(suite.Results), 1)
	}
}
