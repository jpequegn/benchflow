package parser

import (
	"os"
	"testing"
	"time"
)

func TestNewTypeScriptParser(t *testing.T) {
	parser := NewTypeScriptParser()
	if parser == nil {
		t.Error("NewTypeScriptParser() returned nil")
	}
}

func TestTypeScriptParserLanguage(t *testing.T) {
	parser := NewTypeScriptParser()
	if parser.Language() != "typescript" {
		t.Errorf("Language() = %q, want %q", parser.Language(), "typescript")
	}
}

func TestTypeScriptParserParseBasic(t *testing.T) {
	data, err := os.ReadFile("../../testdata/typescript/benchmark_js_basic.txt")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewTypeScriptParser()
	suite, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if suite == nil {
		t.Error("Parse() returned nil suite")
		return
	}

	if suite.Language != "typescript" {
		t.Errorf("suite.Language = %q, want %q", suite.Language, "typescript")
	}

	if len(suite.Results) != 3 {
		t.Errorf("suite.Results length = %d, want 3", len(suite.Results))
	}

	// Check first benchmark
	if suite.Results[0].Name != "StringComparison" {
		t.Errorf("first result Name = %q, want %q", suite.Results[0].Name, "StringComparison")
	}

	if suite.Results[0].Throughput == nil {
		t.Error("first result Throughput is nil")
	} else if suite.Results[0].Throughput.Value != 1234567 {
		t.Errorf("first result Throughput.Value = %f, want 1234567", suite.Results[0].Throughput.Value)
	}

	if suite.Results[0].Iterations != 90 {
		t.Errorf("first result Iterations = %d, want 90", suite.Results[0].Iterations)
	}

	// Check metadata
	if margin, ok := suite.Results[0].Metadata["margin_of_error"]; !ok {
		t.Error("margin_of_error not in metadata")
	} else if margin != "1.23%" {
		t.Errorf("margin_of_error = %q, want %q", margin, "1.23%")
	}
}

func TestTypeScriptParserParseEdgeCases(t *testing.T) {
	data, err := os.ReadFile("../../testdata/typescript/benchmark_js_edge_cases.txt")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewTypeScriptParser()
	suite, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(suite.Results) != 6 {
		t.Errorf("suite.Results length = %d, want 6", len(suite.Results))
	}

	// Test single operation
	single := suite.Results[0]
	if single.Name != "Single operation" {
		t.Errorf("single.Name = %q, want %q", single.Name, "Single operation")
	}
	if single.Throughput.Value != 1 {
		t.Errorf("single ops/sec = %f, want 1", single.Throughput.Value)
	}
	if single.Iterations != 5 {
		t.Errorf("single runs = %d, want 5", single.Iterations)
	}

	// Test very large number
	large := suite.Results[1]
	if large.Name != "Very fast" {
		t.Errorf("large.Name = %q, want %q", large.Name, "Very fast")
	}
	if large.Throughput.Value != 123456789 {
		t.Errorf("large ops/sec = %f, want 123456789", large.Throughput.Value)
	}

	// Test special characters in name
	special := suite.Results[2]
	if special.Name != "With-special-chars" {
		t.Errorf("special.Name = %q, want %q", special.Name, "With-special-chars")
	}

	// Test underscores in name
	underscore := suite.Results[3]
	if underscore.Name != "Test_with_underscores" {
		t.Errorf("underscore.Name = %q, want %q", underscore.Name, "Test_with_underscores")
	}

	// Test long name
	long := suite.Results[4]
	if long.Name != "Very_very_long_benchmark_name_with_many_characters" {
		t.Errorf("long.Name = %q, want %q", long.Name, "Very_very_long_benchmark_name_with_many_characters")
	}

	// Test high ops/sec with high margin
	highOps := suite.Results[5]
	if highOps.Name != "run" {
		t.Errorf("highOps.Name = %q, want %q", highOps.Name, "run")
	}
	if highOps.Throughput.Value != 999999999 {
		t.Errorf("highOps ops/sec = %f, want 999999999", highOps.Throughput.Value)
	}
}

func TestTypeScriptParserParseWithFastest(t *testing.T) {
	data, err := os.ReadFile("../../testdata/typescript/benchmark_js_with_fastest.txt")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewTypeScriptParser()
	suite, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	// Should have 5 results (ignoring "Fastest is..." and empty lines)
	if len(suite.Results) != 5 {
		t.Errorf("suite.Results length = %d, want 5", len(suite.Results))
	}

	// First result
	if suite.Results[0].Name != "RegExp#test" {
		t.Errorf("first result Name = %q, want %q", suite.Results[0].Name, "RegExp#test")
	}
	if suite.Results[0].Throughput.Value != 48985511 {
		t.Errorf("first result ops/sec = %f, want 48985511", suite.Results[0].Throughput.Value)
	}
}

func TestTypeScriptParserTableDriven(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCount int
		expectedErr   bool
		validations   func(t *testing.T, suite *BenchmarkSuite)
	}{
		{
			name:          "single benchmark",
			input:         "test x 1,000 ops/sec ±1.0% (10 runs sampled)",
			expectedCount: 1,
			expectedErr:   false,
			validations: func(t *testing.T, suite *BenchmarkSuite) {
				if suite.Results[0].Name != "test" {
					t.Errorf("Name = %q, want %q", suite.Results[0].Name, "test")
				}
				if suite.Results[0].Throughput.Value != 1000 {
					t.Errorf("ops/sec = %f, want 1000", suite.Results[0].Throughput.Value)
				}
				if suite.Results[0].Iterations != 10 {
					t.Errorf("iterations = %d, want 10", suite.Results[0].Iterations)
				}
			},
		},
		{
			name:          "multiple benchmarks",
			input:         "test1 x 1,000 ops/sec ±1.0% (10 runs sampled)\ntest2 x 2,000 ops/sec ±2.0% (20 runs sampled)",
			expectedCount: 2,
			expectedErr:   false,
			validations: func(t *testing.T, suite *BenchmarkSuite) {
				if suite.Results[0].Name != "test1" {
					t.Errorf("first Name = %q, want %q", suite.Results[0].Name, "test1")
				}
				if suite.Results[1].Name != "test2" {
					t.Errorf("second Name = %q, want %q", suite.Results[1].Name, "test2")
				}
			},
		},
		{
			name:          "with empty lines",
			input:         "test1 x 1,000 ops/sec ±1.0% (10 runs sampled)\n\ntest2 x 2,000 ops/sec ±2.0% (20 runs sampled)\n",
			expectedCount: 2,
			expectedErr:   false,
			validations: func(t *testing.T, suite *BenchmarkSuite) {
				if len(suite.Results) != 2 {
					t.Errorf("Results length = %d, want 2", len(suite.Results))
				}
			},
		},
		{
			name:          "with Fastest line",
			input:         "test1 x 1,000 ops/sec ±1.0% (10 runs sampled)\ntest2 x 2,000 ops/sec ±2.0% (20 runs sampled)\nFastest is test2",
			expectedCount: 2,
			expectedErr:   false,
			validations: func(t *testing.T, suite *BenchmarkSuite) {
				if len(suite.Results) != 2 {
					t.Errorf("Results length = %d, want 2 (Fastest line should be skipped)", len(suite.Results))
				}
			},
		},
		{
			name:          "empty input",
			input:         "",
			expectedCount: 0,
			expectedErr:   true,
			validations:   func(t *testing.T, suite *BenchmarkSuite) {},
		},
		{
			name:          "no benchmarks",
			input:         "some random output\nno benchmarks here\n",
			expectedCount: 0,
			expectedErr:   true,
			validations:   func(t *testing.T, suite *BenchmarkSuite) {},
		},
		{
			name:          "throughput calculation",
			input:         "calc x 1,000,000 ops/sec ±0.5% (100 runs sampled)",
			expectedCount: 1,
			expectedErr:   false,
			validations: func(t *testing.T, suite *BenchmarkSuite) {
				result := suite.Results[0]
				// 1,000,000 ops/sec = 1 µs per op = 1000 ns per op
				expectedTime := time.Duration(1000) * time.Nanosecond
				if result.Time != expectedTime {
					t.Errorf("Time = %v, want %v", result.Time, expectedTime)
				}
				if result.Throughput.Unit != "ops/s" {
					t.Errorf("Throughput.Unit = %q, want %q", result.Throughput.Unit, "ops/s")
				}
			},
		},
		{
			name:          "margin of error in metadata",
			input:         "test x 1,000 ops/sec ±5.5% (50 runs sampled)",
			expectedCount: 1,
			expectedErr:   false,
			validations: func(t *testing.T, suite *BenchmarkSuite) {
				margin, ok := suite.Results[0].Metadata["margin_of_error"]
				if !ok {
					t.Error("margin_of_error not in metadata")
				} else if margin != "5.50%" {
					t.Errorf("margin_of_error = %q, want %q", margin, "5.50%")
				}
			},
		},
		{
			name:          "comma-separated large numbers",
			input:         "large x 123,456,789 ops/sec ±0.01% (99 runs sampled)",
			expectedCount: 1,
			expectedErr:   false,
			validations: func(t *testing.T, suite *BenchmarkSuite) {
				if suite.Results[0].Throughput.Value != 123456789 {
					t.Errorf("ops/sec = %f, want 123456789", suite.Results[0].Throughput.Value)
				}
			},
		},
		{
			name:          "singular 'run'",
			input:         "single x 1,000 ops/sec ±1.0% (1 run sampled)",
			expectedCount: 1,
			expectedErr:   false,
			validations: func(t *testing.T, suite *BenchmarkSuite) {
				if suite.Results[0].Iterations != 1 {
					t.Errorf("iterations = %d, want 1", suite.Results[0].Iterations)
				}
			},
		},
		{
			name:          "stddev approximation",
			input:         "test x 1,000 ops/sec ±10.0% (50 runs sampled)",
			expectedCount: 1,
			expectedErr:   false,
			validations: func(t *testing.T, suite *BenchmarkSuite) {
				// stddev should be non-zero when margin > 0
				if suite.Results[0].StdDev == 0 {
					t.Error("StdDev should be non-zero when margin of error > 0")
				}
				if suite.Results[0].StdDev < 0 {
					t.Error("StdDev should be non-negative")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTypeScriptParser()
			suite, err := parser.Parse([]byte(tt.input))

			if (err != nil) != tt.expectedErr {
				t.Errorf("Parse() error = %v, expectedErr = %v", err, tt.expectedErr)
			}

			if !tt.expectedErr {
				if suite == nil {
					t.Error("Parse() returned nil suite")
					return
				}

				if len(suite.Results) != tt.expectedCount {
					t.Errorf("Results count = %d, want %d", len(suite.Results), tt.expectedCount)
				}

				if suite.Language != "typescript" {
					t.Errorf("Language = %q, want %q", suite.Language, "typescript")
				}

				tt.validations(t, suite)
			}
		})
	}
}

func TestTypeScriptParserTimestampIsSet(t *testing.T) {
	parser := NewTypeScriptParser()
	suite, err := parser.Parse([]byte("test x 1,000 ops/sec ±1.0% (10 runs sampled)"))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if suite.Timestamp.IsZero() {
		t.Error("Timestamp is zero")
	}

	// Timestamp should be recent (within last second)
	now := time.Now()
	diff := now.Sub(suite.Timestamp)
	if diff > time.Second {
		t.Errorf("Timestamp difference from now = %v, want < 1 second", diff)
	}
}

func TestTypeScriptParserMetadataPresent(t *testing.T) {
	parser := NewTypeScriptParser()
	suite, err := parser.Parse([]byte("test x 1,000 ops/sec ±1.0% (10 runs sampled)"))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if suite.Metadata == nil {
		t.Error("suite.Metadata is nil")
	}
}

func TestTypeScriptParserErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "empty",
			input:       "",
			expectError: true,
		},
		{
			name:        "only whitespace",
			input:       "   \n  \n   ",
			expectError: true,
		},
		{
			name:        "no valid benchmarks",
			input:       "random text\nmore random text\n",
			expectError: true,
		},
		{
			name:        "malformed ops/sec",
			input:       "test x abc ops/sec ±1.0% (10 runs sampled)",
			expectError: true,
		},
		{
			name:        "malformed margin",
			input:       "test x 1,000 ops/sec ±abc% (10 runs sampled)",
			expectError: true,
		},
		{
			name:        "malformed runs",
			input:       "test x 1,000 ops/sec ±1.0% (abc runs sampled)",
			expectError: true,
		},
		{
			name:        "negative ops/sec",
			input:       "test x -1,000 ops/sec ±1.0% (10 runs sampled)",
			expectError: true,
		},
		{
			name:        "zero ops/sec",
			input:       "test x 0 ops/sec ±1.0% (10 runs sampled)",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTypeScriptParser()
			_, err := parser.Parse([]byte(tt.input))

			if (err != nil) != tt.expectError {
				t.Errorf("Parse() error = %v, expectError = %v", err, tt.expectError)
			}
		})
	}
}

func TestTypeScriptParserResultValues(t *testing.T) {
	input := "test x 10,000 ops/sec ±2.0% (50 runs sampled)"
	parser := NewTypeScriptParser()
	suite, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	result := suite.Results[0]

	// Verify basic fields
	if result.Name != "test" {
		t.Errorf("Name = %q, want %q", result.Name, "test")
	}

	if result.Language != "typescript" {
		t.Errorf("Language = %q, want %q", result.Language, "typescript")
	}

	// 10,000 ops/sec = 100 µs = 100,000 ns
	expectedTime := time.Duration(100000) * time.Nanosecond
	if result.Time != expectedTime {
		t.Errorf("Time = %v, want %v", result.Time, expectedTime)
	}

	if result.Iterations != 50 {
		t.Errorf("Iterations = %d, want 50", result.Iterations)
	}

	// Verify throughput
	if result.Throughput == nil {
		t.Error("Throughput is nil")
	} else {
		if result.Throughput.Value != 10000 {
			t.Errorf("Throughput.Value = %f, want 10000", result.Throughput.Value)
		}
		if result.Throughput.Unit != "ops/s" {
			t.Errorf("Throughput.Unit = %q, want %q", result.Throughput.Unit, "ops/s")
		}
	}

	// Verify StdDev is positive
	if result.StdDev <= 0 {
		t.Errorf("StdDev = %v, should be positive", result.StdDev)
	}
}
