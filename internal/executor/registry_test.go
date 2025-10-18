package executor

import (
	"testing"

	"github.com/jpequegn/benchflow/internal/parser"
)

func TestParserRegistry(t *testing.T) {
	t.Run("RegisterAndGetParser", func(t *testing.T) {
		registry := NewParserRegistry()
		rustParser := parser.NewRustParser()

		// Register parser
		registry.RegisterParser("rust", rustParser)

		// Get parser
		p, err := registry.GetParser("rust")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p == nil {
			t.Fatal("expected parser, got nil")
		}
		if p.Language() != "rust" {
			t.Errorf("expected rust parser, got %s", p.Language())
		}
	})

	t.Run("GetNonExistentParser", func(t *testing.T) {
		registry := NewParserRegistry()

		// Try to get unregistered parser
		_, err := registry.GetParser("python")
		if err == nil {
			t.Fatal("expected error for unregistered parser")
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		registry := NewParserRegistry()
		rustParser := parser.NewRustParser()
		registry.RegisterParser("rust", rustParser)

		// Concurrent reads should be safe
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				_, err := registry.GetParser("rust")
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
