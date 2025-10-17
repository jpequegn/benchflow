package cmd

import (
	"bytes"
	"testing"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "help flag",
			args:    []string{"--help"},
			wantErr: false,
		},
		{
			name:    "version flag",
			args:    []string{"--version"},
			wantErr: false,
		},
		{
			name:    "verbose flag",
			args:    []string{"--verbose", "--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			// Set args
			rootCmd.SetArgs(tt.args)

			// Execute
			err := rootCmd.Execute()

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Reset for next test
			rootCmd.SetArgs(nil)
		})
	}
}

func TestInitConfig(t *testing.T) {
	// Test that config initialization doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("initConfig() panicked: %v", r)
		}
	}()

	initConfig()
}

func TestInitLogger(t *testing.T) {
	// Test that logger initialization doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("initLogger() panicked: %v", r)
		}
	}()

	initLogger()
}
