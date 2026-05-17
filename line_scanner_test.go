package nntpReaderWriter

import (
	"bytes"
	"io"
	"testing"
)

var lineScannerTests = []struct {
	name          string
	input         string
	expectedLines []string
	expectedError error
}{
	{
		name:          "BasicLines",
		input:         "line1\nline2\nline3\n",
		expectedLines: []string{"line1", "line2", "line3"},
		expectedError: io.EOF,
	},
	{
		name:          "CRLFLines",
		input:         "line1\r\nline2\r\nline3\r\n",
		expectedLines: []string{"line1", "line2", "line3"},
		expectedError: io.EOF,
	},
	{
		name:          "MixedLineEndings",
		input:         "line1\r\nline2\nline3\r\nline4\n",
		expectedLines: []string{"line1", "line2", "line3", "line4"},
		expectedError: io.EOF,
	},
	{
		name:          "NoFinalNewline",
		input:         "line1\nline2\nline3",
		expectedLines: []string{"line1", "line2", "line3"},
		expectedError: io.ErrUnexpectedEOF,
	},
	{
		name:          "EmptyInput",
		input:         "",
		expectedLines: nil,
		expectedError: io.EOF,
	},
	{
		name:          "LongLine",
		input:         string(bytes.Repeat([]byte("x"), 10000)) + "\nshort\n",
		expectedLines: []string{string(bytes.Repeat([]byte("x"), 10000)), "short"},
		expectedError: io.EOF,
	},
}

func TestLineScanner(t *testing.T) {
	for _, tt := range lineScannerTests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tt.input))
			scanner := NewLineScanner()

			var lines []string
			var err error
			var line []byte
			for {
				line, err = scanner.ScanLine(r)
				if err != nil && err != tt.expectedError {
					t.Fatalf("unexpected error: %v", err)
				}
				if line == nil {
					break
				}
				lines = append(lines, string(line))
				if err == tt.expectedError {
					break
				}
			}

			expected := tt.expectedLines
			if len(lines) != len(expected) {
				t.Fatalf("expected %d lines, got %d", len(expected), len(lines))
			}
			for i, l := range expected {
				if lines[i] != l {
					t.Errorf("line %d: expected %q, got %q", i, l, lines[i])
				}
			}
			if tt.expectedError != nil && err != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}
