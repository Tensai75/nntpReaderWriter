package lineReader

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"
)

type errorAfterReader struct {
	r   io.Reader
	err error
}

func (e *errorAfterReader) Read(p []byte) (int, error) {
	n, err := e.r.Read(p)
	if err == io.EOF {
		return n, e.err // return your custom error instead of io.EOF
	}
	return n, err
}

func newErrorAfterReader(r io.Reader, err error) *errorAfterReader {
	return &errorAfterReader{r: r, err: err}
}

var lineReaderBufferedTests = []struct {
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
		name:          "NoFinalNewlineSingleLine",
		input:         "line1",
		expectedLines: []string{"line1"},
		expectedError: io.ErrUnexpectedEOF,
	},
	{
		name:          "EmptyInput",
		input:         "",
		expectedLines: nil,
		expectedError: io.ErrUnexpectedEOF,
	},
	{
		name:          "LongLines",
		input:         string(bytes.Repeat([]byte("x"), 100000)) + "\nshort\n" + string(bytes.Repeat([]byte("x"), 100000)) + "\nshort\n" + string(bytes.Repeat([]byte("x"), 100000)) + "\nshort\n",
		expectedLines: []string{string(bytes.Repeat([]byte("x"), 100000)), "short", string(bytes.Repeat([]byte("x"), 100000)), "short", string(bytes.Repeat([]byte("x"), 100000)), "short"},
		expectedError: io.EOF,
	},
}

func TestLineReaderBuffered(t *testing.T) {
	for _, tt := range lineReaderBufferedTests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tt.input))
			reader := NewLineReader(r, LineReaderOptions{})
			reader.Reset()
			var lines []string
			var err error
			var line []byte
			for {
				line, err = reader.ReadLineBuffered()
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

func TestLineReaderBuffered_Reset(t *testing.T) {
	input := "line1\nline2\nline3\n"
	r := bytes.NewReader([]byte(input))
	reader := NewLineReader(r, LineReaderOptions{})
	// With the first read, the buffer is filled with "line1\nline2\nline3\n" and "line1" is returned as the first line.
	line, err := reader.ReadLineBuffered()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(line) != "line1" {
		t.Fatalf("expected %q, got %q", "line1", string(line))
	}
	reader.Reset()
	// After Reset(), the buffer is cleared and the next read should result in a new read on the exhausted reader.
	// Since the reader is exhausted, we expect to get an unexpected EOF error and no line.
	line, err = reader.ReadLineBuffered()
	if err != io.ErrUnexpectedEOF {
		t.Fatalf("unexpected error after reset: expected %v, got %v", io.ErrUnexpectedEOF, err)
	}
	if line != nil {
		t.Fatalf("expected %v after reset, got %q", nil, line)
	}

	r = bytes.NewReader([]byte(input))
	// We create a new LineReader with a small buffer size to force multiple reads from the reader.
	reader = NewLineReader(r, LineReaderOptions{
		BufferSize: 12,
	})
	// With the first read, the buffer is filled with "line1\nline2\n" and "line1" is returned as the first line.
	line, err = reader.ReadLineBuffered()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(line) != "line1" {
		t.Fatalf("expected %q, got %q", "line1", string(line))
	}
	reader.Reset()
	// With Reset(), the buffer is cleared and "line3\n" is read into the buffer, so the next read should return "line3".
	line, err = reader.ReadLineBuffered()
	if err != nil {
		t.Fatalf("unexpected error after reset: expected %v, got %v", nil, err)
	}
	if string(line) != "line3" {
		t.Fatalf("expected %q after reset, got %q", "line3", string(line))
	}
}

func TestLineReaderBuffered_ReadError(t *testing.T) {
	for _, tt := range lineReaderBufferedTests {
		t.Run(tt.name, func(t *testing.T) {
			r := newErrorAfterReader(bytes.NewReader([]byte(tt.input)), fmt.Errorf("TESTERROR"))
			reader := NewLineReader(r, LineReaderOptions{})
			reader.Reset()
			var lines []string
			var err error
			var line []byte
			for {
				line, err = reader.ReadLineBuffered()
				if err != nil && !(err == io.ErrUnexpectedEOF || err.Error() == "TESTERROR") {
					t.Fatalf("unexpected error: %v", err)
				}
				if line == nil {
					break
				}
				lines = append(lines, string(line))
				if err != nil {
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
			if err != nil && !(err == io.ErrUnexpectedEOF || err.Error() == "TESTERROR") {
				t.Errorf("expected error %v, got %v", "TESTERROR", err)
			}
		})
	}

}

var lineReaderUnbufferedTests = []struct {
	name          string
	input         string
	expectedLine  string
	expectedError error
}{
	{
		name:          "BasicLines",
		input:         "line1\nline2\nline3\n",
		expectedLine:  "line1",
		expectedError: nil,
	},
	{
		name:          "CRLFLines",
		input:         "line1\r\nline2\r\nline3\r\n",
		expectedLine:  "line1",
		expectedError: nil,
	},
	{
		name:          "NoFinalNewline",
		input:         "line1",
		expectedLine:  "line1",
		expectedError: io.ErrUnexpectedEOF,
	},
	{
		name:          "EmptyLine",
		input:         "\n",
		expectedLine:  "",
		expectedError: nil,
	},
	{
		name:          "EmptyInput",
		input:         "",
		expectedLine:  "",
		expectedError: io.ErrUnexpectedEOF,
	},
	{
		name:          "LongLine",
		input:         string(bytes.Repeat([]byte("x"), 10000)) + "\nshort\n",
		expectedLine:  string(bytes.Repeat([]byte("x"), 10000)),
		expectedError: nil,
	},
}

func TestLineReaderUnbuffered(t *testing.T) {
	lineReaderOptions := LineReaderOptions{
		Timeout: 100 * time.Millisecond,
	}
	for _, tt := range lineReaderUnbufferedTests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tt.input))
			reader := NewLineReader(r, lineReaderOptions)
			reader.Reset()
			line, err := reader.ReadLineUnbuffered()
			fmt.Printf("expected error %v, got %v\n", tt.expectedError, err)
			if err != tt.expectedError {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}
			if string(line) != tt.expectedLine {
				t.Errorf("expected line %q, got %q", tt.expectedLine, string(line))
			}
		})
	}
}
