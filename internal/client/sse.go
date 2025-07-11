package client

import (
	"bufio"
	"io"
)

// SSEScanner scans Server-Sent Events from a reader
type SSEScanner struct {
	scanner *bufio.Scanner
	err     error
}

// NewSSEScanner creates a new SSE scanner
func NewSSEScanner(r io.Reader) *SSEScanner {
	return &SSEScanner{
		scanner: bufio.NewScanner(r),
	}
}

// Scan advances the scanner to the next line
func (s *SSEScanner) Scan() bool {
	for s.scanner.Scan() {
		line := s.scanner.Text()

		// Skip empty lines and comments
		if line == "" || line[0] == ':' {
			continue
		}

		return true
	}

	s.err = s.scanner.Err()
	return false
}

// Text returns the current line
func (s *SSEScanner) Text() string {
	return s.scanner.Text()
}

// Err returns any error that occurred during scanning
func (s *SSEScanner) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.scanner.Err()
}
