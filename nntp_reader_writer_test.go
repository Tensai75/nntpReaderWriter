package nntpReaderWriter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	mock "github.com/Tensai75/nntpReaderWriter/testutils"
)

type TestStep struct {
	ScriptSteps       []mock.ScriptStep
	Command           string
	WriteLines        []string
	WriteData         []byte
	ExpectedLines     []string
	ExpectedBytes     []byte
	ExpectedCode      int
	ExpectedMsg       string
	ExpectedError     error
	ExpectedErrorType func(error) bool
}

type TestScript struct {
	Name  string
	Steps []TestStep
}

func NewTestReaderWriterWithTestScript(script TestScript, option NntpReaderWriterOptions) *NntpReaderWriter {
	var scriptSteps []mock.ScriptStep
	for _, step := range script.Steps {
		scriptSteps = append(scriptSteps, step.ScriptSteps...)
	}
	conn := mock.NewScriptedConn(scriptSteps)
	return NewNntpReaderWriter(conn, option)
}

var singleResponseLineTests = []TestScript{
	{
		Name: "Basic",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("111 20260518101530\r\n"),
					},
				},
				Command:       "DATE",
				ExpectedCode:  111,
				ExpectedMsg:   "20260518101530",
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "UTF-8",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("111 こんにちは世界\r\n"),
					},
				},
				Command:       "DATE",
				ExpectedCode:  111,
				ExpectedMsg:   "こんにちは世界",
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "UnexpectedCode",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("340 Send article to be posted\r\n"),
					},
				},
				Command:           "DATE",
				ExpectedCode:      340,
				ExpectedMsg:       "Send article to be posted",
				ExpectedError:     errUnexpectedResponseCodeError(340, "Send article to be posted"),
				ExpectedErrorType: IsUnexpectedResponseCodeError,
			},
		},
	},
	{
		Name: "ErrorCode",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("500 Internal Server Error\r\n"),
					},
				},
				Command:           "DATE",
				ExpectedCode:      500,
				ExpectedMsg:       "Internal Server Error",
				ExpectedError:     errNntpError(500, "Internal Server Error"),
				ExpectedErrorType: IsNntpError,
			},
		},
	},
	{
		Name: "InvalidResponseLine",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("Invalid Response\r\n"),
					},
				},
				Command:           "DATE",
				ExpectedCode:      0,
				ExpectedMsg:       "",
				ExpectedError:     errInvalidResponseLine(),
				ExpectedErrorType: IsInvalidResponseLineError,
			},
		},
	},
}

func TestSingleLineCmd(t *testing.T) {
	for _, test := range singleResponseLineTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				code, msg, err := rw.SingleResponseLineCmd(step.Command)
				testError(t, err, step)
				testCode(t, code, step)
				testMsg(t, msg, step)
			}
		})
	}
}

var readerTests = []TestScript{
	{
		Name: "Basic",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("OVER 1-3\r\n"),
						Response:      []byte("224 Overview information follows\r\nline1\r\nline2\r\nline3\r\n.\r\n"),
					},
				},
				Command:       "OVER 1-3",
				ExpectedCode:  224,
				ExpectedMsg:   "Overview information follows",
				ExpectedLines: []string{"line1", "line2", "line3"},
				ExpectedBytes: []byte("line1\nline2\nline3\n"),
			},
		},
	},
	{
		Name: "EmptyLines",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("OVER 1-3\r\n"),
						Response:      []byte("224 Overview information follows\r\n\r\n\r\n.\r\n"),
					},
				},
				Command:       "OVER 1-3",
				ExpectedCode:  224,
				ExpectedMsg:   "Overview information follows",
				ExpectedLines: []string{"", ""},
				ExpectedBytes: []byte("\n\n"),
			},
		},
	},
	{
		Name: "UTF-8",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("OVER 1-3\r\n"),
						Response:      []byte("224 Overview information follows\r\nこんにちは\r\n世界\r\n.\r\n"),
					},
				},
				Command:       "OVER 1-3",
				ExpectedCode:  224,
				ExpectedMsg:   "Overview information follows",
				ExpectedLines: []string{"こんにちは", "世界"},
				ExpectedBytes: []byte("こんにちは\n世界\n"),
			},
		},
	},
	{
		Name: "ReadTimeout",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("OVER 1-3\r\n"),
						Response:      []byte("224 Overview information follows\r\nline1\r\nline2\r\nline3\r\n.\r\n"),
						ReadTimeout:   true,
					},
				},
				Command:           "OVER 1-3",
				ExpectedCode:      0,
				ExpectedMsg:       "",
				ExpectedLines:     []string{},
				ExpectedBytes:     []byte(""),
				ExpectedError:     &mock.TimeoutError{Op: "read", Net: "mock", Err: errors.New("i/o timeout")},
				ExpectedErrorType: IsTimeOutError,
			},
		},
	},
}

func TestDotLinesReadCmdAsStrings(t *testing.T) {
	for _, test := range readerTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				code, msg, lines, err := rw.DotLinesReadCmdAsStrings(step.Command)
				testError(t, err, step)
				testCode(t, code, step)
				testMsg(t, msg, step)
				testLines(t, lines, step)
			}
		})
	}
}

func TestDotLinesReadCmdAsReader(t *testing.T) {
	for _, test := range readerTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				code, msg, reader, err := rw.DotLinesReadCmdAsReader(step.Command)
				testError(t, err, step)
				testCode(t, code, step)
				testMsg(t, msg, step)
				if reader == nil && err == nil {
					t.Fatalf("expected reader, got nil")
				}
				if reader != nil {
					lines, err := io.ReadAll(reader)
					if err != nil {
						t.Fatalf("unexpected error reading from reader: %v", err)
					}
					testBytes(t, lines, step)
				}
			}
		})
	}
}

func TestDotLinesReadCmdAsStringsWithCallback(t *testing.T) {
	for _, test := range readerTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				lines := make([]string, 0)
				callback := func(line string) error {
					lines = append(lines, line)
					return nil
				}
				code, msg, err := rw.DotLinesReadCmdAsStringsWithCallback(step.Command, callback)
				testError(t, err, step)
				testCode(t, code, step)
				testMsg(t, msg, step)
				testLines(t, lines, step)
			}
		})
	}
}

func TestDotLinesReadCmdAsBytesWithCallback(t *testing.T) {
	for _, test := range readerTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				lines := make([]byte, 0)
				callback := func(line []byte) error {
					lines = append(lines, append(line, '\n')...)
					return nil
				}
				code, msg, err := rw.DotLinesReadCmdAsBytesWithCallback(step.Command, callback)
				testError(t, err, step)
				testCode(t, code, step)
				testMsg(t, msg, step)
				testBytes(t, lines, step)
			}
		})
	}
}

var writerTests = []TestScript{
	{
		Name: "ValidLines",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("POST\r\n"),
						Response:      []byte("340 Send article to be posted\r\n"),
					},
					{
						ExpectedWrite:  []byte("line1\r\n"),
						AwaitNextWrite: true,
					},
					{
						ExpectedWrite:  []byte("line2\r\n"),
						AwaitNextWrite: true,
					},
					{
						ExpectedWrite:  []byte("line3\r\n"),
						AwaitNextWrite: true,
					},
					{
						ExpectedWrite: []byte(".\r\n"),
						Response:      []byte("240 Article received OK\r\n"),
					},
				},
				Command:       "POST",
				WriteLines:    []string{"line1", "line2", "line3"},
				WriteData:     []byte("line1\nline2\nline3\n"),
				ExpectedCode:  240,
				ExpectedMsg:   "Article received OK",
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "InvalidLine",
		Steps: []TestStep{
			{
				ScriptSteps: []mock.ScriptStep{
					{
						ExpectedWrite: []byte("POST\r\n"),
						Response:      []byte("340 Send article to be posted\r\n"),
					},
					{
						ExpectedWrite:  []byte("line1\r\n"),
						AwaitNextWrite: true,
					},
				},
				Command:           "POST",
				WriteLines:        []string{"line1", "line2\rline3"},
				WriteData:         []byte("line1\nline2\rline3\n"),
				ExpectedCode:      0,
				ExpectedMsg:       "",
				ExpectedError:     fmt.Errorf("invalid character %q in line to write", "\\r"),
				ExpectedErrorType: IsInvalidWriteLineError,
			},
		},
	},
}

func TestDotLinesWriteCmdFromStrings(t *testing.T) {
	for _, test := range writerTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				code, msg, err := rw.DotLinesWriteCmdFromStrings(step.Command, step.WriteLines)
				testError(t, err, step)
				testCode(t, code, step)
				testMsg(t, msg, step)
			}
		})
	}
}

func TestDotLinesWriteCmdFromReader(t *testing.T) {
	for _, test := range writerTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				reader := bytes.NewReader(step.WriteData)
				code, msg, err := rw.DotLinesWriteCmdFromReader(step.Command, reader)
				testError(t, err, step)
				testCode(t, code, step)
				testMsg(t, msg, step)
			}
		})
	}
}

func TestDotLinesWriteCmdFromChan(t *testing.T) {
	for _, test := range writerTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				lineChan := make(chan string)
				go func() {
					for _, line := range step.WriteLines {
						lineChan <- line
					}
					close(lineChan)
				}()
				code, msg, err := rw.DotLinesWriteCmdFromChan(step.Command, lineChan)
				testError(t, err, step)
				testCode(t, code, step)
				testMsg(t, msg, step)
			}
		})
	}
}

// helper functions

func testError(t *testing.T, err error, step TestStep) {
	if err != nil && step.ExpectedError == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err == nil && step.ExpectedError != nil {
		t.Fatalf("expected error %v, got nil", step.ExpectedError)
	}
	if err != nil && step.ExpectedError != nil && err.Error() != step.ExpectedError.Error() {
		t.Fatalf("expected error %v, got %v", step.ExpectedError, err)
	}
	if step.ExpectedErrorType != nil && !step.ExpectedErrorType(err) {
		t.Fatalf("expected error type %v, got %v", step.ExpectedError, err)
	}
}

func testCode(t *testing.T, code int, step TestStep) {
	if code != step.ExpectedCode {
		t.Fatalf("expected code %v, got %v", step.ExpectedCode, code)
	}
}

func testMsg(t *testing.T, msg string, step TestStep) {
	if msg != step.ExpectedMsg {
		t.Fatalf("expected msg %v, got %v", step.ExpectedMsg, msg)
	}
}

func testLines(t *testing.T, lines []string, step TestStep) {
	if len(lines) != len(step.ExpectedLines) {
		t.Fatalf("expected number of lines %v, got %v", len(step.ExpectedLines), len(lines))
	}
	for i, line := range lines {
		testLine(t, line, step.ExpectedLines[i])
	}
}

func testLine(t *testing.T, line string, expectedLine string) {
	if line != expectedLine {
		t.Fatalf("expected line %v, got %v", expectedLine, line)
	}
}

func testBytes(t *testing.T, data []byte, step TestStep) {
	if len(data) != len(step.ExpectedBytes) {
		t.Fatalf("expected number of bytes %v, got %v", len(step.ExpectedBytes), len(data))
	}
	if !bytes.Equal(data, step.ExpectedBytes) {
		t.Fatalf("expected bytes %v, got %v", string(step.ExpectedBytes), string(data))
	}
}
