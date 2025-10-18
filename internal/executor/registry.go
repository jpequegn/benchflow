package executor

import (
	"fmt"
	"sync"

	"github.com/jpequegn/benchflow/internal/parser"
)

// DefaultParserRegistry provides a thread-safe parser registry
type DefaultParserRegistry struct {
	mu      sync.RWMutex
	parsers map[string]parser.Parser
}

// NewParserRegistry creates a new parser registry
func NewParserRegistry() *DefaultParserRegistry {
	return &DefaultParserRegistry{
		parsers: make(map[string]parser.Parser),
	}
}

// GetParser returns a parser for the specified language
func (r *DefaultParserRegistry) GetParser(language string) (parser.Parser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.parsers[language]
	if !ok {
		return nil, fmt.Errorf("no parser registered for language: %s", language)
	}
	return p, nil
}

// RegisterParser registers a parser for a language
func (r *DefaultParserRegistry) RegisterParser(language string, p parser.Parser) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.parsers[language] = p
}
