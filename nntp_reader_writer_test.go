package nntpReaderWriter

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Tensai75/nntpReaderWriter/mockScriptedConn"
)

type TestStep struct {
	ScriptSteps       []mockScriptedConn.ScriptStep
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
	var scriptSteps []mockScriptedConn.ScriptStep
	for _, step := range script.Steps {
		scriptSteps = append(scriptSteps, step.ScriptSteps...)
	}
	conn := mockScriptedConn.NewScriptedConn(scriptSteps)
	return NewNntpReaderWriter(conn, option)
}

var readAndCheckCodeResponseLineTests = []TestScript{
	{
		Name: "NoCheck",
		Steps: []TestStep{
			{
				ScriptSteps: []mockScriptedConn.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("111 20260518101530\r\n"),
					},
				},
				Command:       "DATE",
				ExpectedCode:  0,
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "SingleDigitExpectedCode",
		Steps: []TestStep{
			{
				ScriptSteps: []mockScriptedConn.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("111 20260518101530\r\n"),
					},
				},
				Command:       "DATE",
				ExpectedCode:  1,
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "DoubleDigitExpectedCode",
		Steps: []TestStep{
			{
				ScriptSteps: []mockScriptedConn.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("111 20260518101530\r\n"),
					},
				},
				Command:       "DATE",
				ExpectedCode:  11,
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "ExactExpectedCode",
		Steps: []TestStep{
			{
				ScriptSteps: []mockScriptedConn.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("111 20260518101530\r\n"),
					},
				},
				Command:       "DATE",
				ExpectedCode:  111,
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "ErrorCode",
		Steps: []TestStep{
			{
				ScriptSteps: []mockScriptedConn.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("411 20260518101530\r\n"),
					},
				},
				Command:       "DATE",
				ExpectedCode:  111,
				ExpectedError: errNntpError(411, "20260518101530"),
			},
		},
	},
	{
		Name: "SingleDigitUnexpectedCode",
		Steps: []TestStep{
			{
				ScriptSteps: []mockScriptedConn.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("211 20260518101530\r\n"),
					},
				},
				Command:       "DATE",
				ExpectedCode:  1,
				ExpectedError: errUnexpectedResponseCodeError(211, "20260518101530"),
			},
		},
	},
	{
		Name: "DoubleDigitUnexpectedCode",
		Steps: []TestStep{
			{
				ScriptSteps: []mockScriptedConn.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("211 20260518101530\r\n"),
					},
				},
				Command:       "DATE",
				ExpectedCode:  11,
				ExpectedError: errUnexpectedResponseCodeError(211, "20260518101530"),
			},
		},
	},
	{
		Name: "ExactUnexpectedCode",
		Steps: []TestStep{
			{
				ScriptSteps: []mockScriptedConn.ScriptStep{
					{
						ExpectedWrite: []byte("DATE\r\n"),
						Response:      []byte("211 20260518101530\r\n"),
					},
				},
				Command:       "DATE",
				ExpectedCode:  111,
				ExpectedError: errUnexpectedResponseCodeError(211, "20260518101530"),
			},
		},
	},
}

func TestCheckCode(t *testing.T) {
	for _, test := range readAndCheckCodeResponseLineTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				err := rw.WriteLine(step.Command)
				if err != nil {
					t.Fatalf("unexpected error writing command: %v", err)
				}
				_, _, err = rw.ReadCodeResponseLine()
				err = rw.CheckCode(step.ExpectedCode)
				testError(t, err, step)
			}
		})
	}
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	conn := mockScriptedConn.NewConcurrencyTestConn()
	rw := NewNntpReaderWriter(conn, NntpReaderWriterOptions{})
	done := sync.WaitGroup{}
	errs := make(chan error, 100)
	for i := 1; i <= 100; i++ {
		done.Add(1)
		go func(i int) {
			id := rw.StartSequence()
			defer func() { done.Done(); rw.EndSequence(id) }()
			err := rw.WriteDotLine(fmt.Sprintf("line%d", i))
			if err != nil {
				errs <- fmt.Errorf("unexpected error in write sequence: %v", err)
				return
			}
			_, msg, err := rw.ReadCodeResponseLine()
			if err != nil {
				errs <- fmt.Errorf("unexpected error in read sequence: %v", err)
				return
			}
			if msg != fmt.Sprintf("line%d", i) {
				errs <- fmt.Errorf("expected message %q, got %q", fmt.Sprintf("line%d", i), msg)
			}
		}(i)
		time.Sleep(10 * time.Millisecond) // Stagger the goroutines slightly
	}
	done.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
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
	if code != step.ExpectedCode && step.ExpectedError == nil {
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
