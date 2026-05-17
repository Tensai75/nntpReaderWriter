package nntpReaderWriter

import (
	"bytes"
	"strings"
	"testing"

	mock "github.com/Tensai75/nntpReaderWriter/testutils"
)

var writeDotLineTests = []TestScript{
	{
		Name: "Basic",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("hello world\r\n")},
				WriteData:     []byte("hello world"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "UTF-8",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("こんにちは世界\r\n")},
				WriteData:     []byte("こんにちは世界"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "VeryLongLine",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte(strings.Repeat("a", 70*1024) + "\r\n")},
				WriteData:     []byte(strings.Repeat("a", 70*1024)),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "EmptyLine",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("\r\n")},
				WriteData:     []byte(""),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "DotPrefix",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("..hello\r\n")},
				WriteData:     []byte(".hello"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "TwoDotsPrefix",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("...hello\r\n")},
				WriteData:     []byte("..hello"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "DotOnlyLine",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("..\r\n")},
				WriteData:     []byte("."),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "LeadingSpace",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte(" hello\r\n")},
				WriteData:     []byte(" hello"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "TrailingSpace",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("hello \r\n")},
				WriteData:     []byte("hello "),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "LeadingTab",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("\thello\r\n")},
				WriteData:     []byte("\thello"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "TrailingTab",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("hello\t\r\n")},
				WriteData:     []byte("hello\t"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "LineWithNewline",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("hello\nworld\r\n")},
				WriteData:     []byte("hello\nworld"),
				ExpectedError: ErrInvalidWriteline("\\n"),
			},
		},
	},
	{
		Name: "LineWithCarriageReturn",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("hello\rworld\r\n")},
				WriteData:     []byte("hello\rworld"),
				ExpectedError: ErrInvalidWriteline("\\r"),
			},
		},
	},
	{
		Name: "LineWithNullByte",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte("hello\x00world\r\n")},
				WriteData:     []byte("hello\x00world"),
				ExpectedError: ErrInvalidWriteline("\\x00"),
			},
		},
	},
}

func TestWriteDotLineFromString(t *testing.T) {
	for _, test := range writeDotLineTests {
		t.Run(test.Name, func(t *testing.T) {
			for _, step := range test.Steps {
				rw := NewTestReaderWriterWithTestScript(TestScript{Steps: []TestStep{step}}, NntpReaderWriterOptions{})
				err := rw.writeDotLineFromString(string(step.WriteData))
				if errorToString(err) != errorToString(step.ExpectedError) {
					t.Errorf("writeDotLineFromString(%q) error = %v; want %v", step.WriteData, err, step.ExpectedError)
				}
			}
		})
	}
}

func TestWriteDotLineFromBytes(t *testing.T) {
	for _, test := range writeDotLineTests {
		t.Run(test.Name, func(t *testing.T) {
			for _, step := range test.Steps {
				rw := NewTestReaderWriterWithTestScript(TestScript{Steps: []TestStep{step}}, NntpReaderWriterOptions{})
				err := rw.writeDotLineFromBytes(step.WriteData)
				if errorToString(err) != errorToString(step.ExpectedError) {
					t.Errorf("writeDotLineFromBytes(%q) error = %v; want %v", step.WriteData, err, step.ExpectedError)
				}
			}
		})
	}
}

func TestWriteDotLinesFromStrings(t *testing.T) {
	testScript, stringsToWrite, _ := prepareDotLinesTestScript()
	rw := NewTestReaderWriterWithTestScript(testScript, NntpReaderWriterOptions{})
	err := rw.writeDotLinesFromStrings(stringsToWrite)
	if errorToString(err) != errorToString(testScript.Steps[0].ExpectedError) {
		t.Errorf("writeDotLinesFromStrings error = %v; want %v", err, testScript.Steps[0].ExpectedError)
	}
}

func TestWriteDotLinesFromReader(t *testing.T) {
	testScript, _, bytesToWrite := prepareDotLinesTestScript()
	rw := NewTestReaderWriterWithTestScript(testScript, NntpReaderWriterOptions{})
	err := rw.writeDotLinesFromReader(bytes.NewReader(bytesToWrite))
	if errorToString(err) != errorToString(testScript.Steps[0].ExpectedError) {
		t.Errorf("writeDotLinesFromReader error = %v; want %v", err, testScript.Steps[0].ExpectedError)
	}
}

func errorToString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

func prepareDotLinesTestScript() (testScript TestScript, stringsToWrite []string, bytesToWrite []byte) {
	var steps []TestStep
	for _, test := range writeDotLineTests {
		for _, step := range test.Steps {
			if step.ExpectedError != nil {
				continue
			}
			stringsToWrite = append(stringsToWrite, string(step.WriteData))
			bytesToWrite = append(bytesToWrite, append(step.WriteData, []byte("\n")...)...)
			step := TestStep{
				ScriptStep:    mock.ScriptStep{ExpectedWrite: step.ScriptStep.ExpectedWrite, AwaitNextWrite: true},
				ExpectedError: step.ExpectedError,
			}
			steps = append(steps, step)
		}
	}
	steps = append(steps, TestStep{ // Final step for the dot line
		ScriptStep:    mock.ScriptStep{ExpectedWrite: []byte(".\r\n")},
		ExpectedError: nil,
	})
	testScript = TestScript{
		Name:  "WriteDotLinesFromStrings",
		Steps: steps,
	}
	return
}
