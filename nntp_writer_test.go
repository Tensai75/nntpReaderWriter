package nntpReaderWriter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Tensai75/nntpReaderWriter/mockScriptedConn"
)

var writeDotLineTests = []TestScript{
	{
		Name: "Basic",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("hello world\r\n")}},
				WriteData:     []byte("hello world"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "UTF-8",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("こんにちは世界\r\n")}},
				WriteData:     []byte("こんにちは世界"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "VeryLongLine",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte(strings.Repeat("a", 70*1024) + "\r\n")}},
				WriteData:     []byte(strings.Repeat("a", 70*1024)),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "EmptyLine",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("\r\n")}},
				WriteData:     []byte(""),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "DotPrefix",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("..hello\r\n")}},
				WriteData:     []byte(".hello"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "TwoDotsPrefix",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("...hello\r\n")}},
				WriteData:     []byte("..hello"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "DotOnlyLine",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("..\r\n")}},
				WriteData:     []byte("."),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "LeadingSpace",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte(" hello\r\n")}},
				WriteData:     []byte(" hello"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "TrailingSpace",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("hello \r\n")}},
				WriteData:     []byte("hello "),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "LeadingTab",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("\thello\r\n")}},
				WriteData:     []byte("\thello"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "TrailingTab",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("hello\t\r\n")}},
				WriteData:     []byte("hello\t"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "LineWithNewline",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("hello\nworld\r\n")}},
				WriteData:     []byte("hello\nworld"),
				ExpectedError: errInvalidWriteLine("\\n"),
			},
		},
	},
	{
		Name: "LineWithCarriageReturn",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("hello\rworld\r\n")}},
				WriteData:     []byte("hello\rworld"),
				ExpectedError: errInvalidWriteLine("\\r"),
			},
		},
	},
	{
		Name: "LineWithNullByte",
		Steps: []TestStep{
			{
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte("hello\x00world\r\n")}},
				WriteData:     []byte("hello\x00world"),
				ExpectedError: errInvalidWriteLine("\\x00"),
			},
		},
	},
}

func TestWriteDotLineFromString(t *testing.T) {
	for _, test := range writeDotLineTests {
		t.Run(test.Name, func(t *testing.T) {
			for _, step := range test.Steps {
				rw := NewTestReaderWriterWithTestScript(TestScript{Steps: []TestStep{step}}, NntpReaderWriterOptions{})
				err := rw.WriteDotLine(string(step.WriteData))
				testError(t, err, step)
			}
		})
	}
}

func TestWriteDotLineFromBytes(t *testing.T) {
	for _, test := range writeDotLineTests {
		t.Run(test.Name, func(t *testing.T) {
			for _, step := range test.Steps {
				rw := NewTestReaderWriterWithTestScript(TestScript{Steps: []TestStep{step}}, NntpReaderWriterOptions{})
				err := rw.WriteDotLineBytes(step.WriteData)
				testError(t, err, step)
			}
		})
	}
}

func TestWriteDotLinesFromStrings(t *testing.T) {
	testScript, stringsToWrite, _ := prepareDotLinesTestScript()
	t.Run(testScript.Name, func(t *testing.T) {
		for _, step := range testScript.Steps {
			rw := NewTestReaderWriterWithTestScript(testScript, NntpReaderWriterOptions{})
			err := rw.WriteDotLines(stringsToWrite)
			testError(t, err, step)
		}
	})
}

func TestWriteDotLinesFromReader(t *testing.T) {
	testScript, _, bytesToWrite := prepareDotLinesTestScript()
	t.Run(testScript.Name, func(t *testing.T) {
		for _, step := range testScript.Steps {
			rw := NewTestReaderWriterWithTestScript(testScript, NntpReaderWriterOptions{})
			err := rw.WriteDotLinesReader(bytes.NewReader(bytesToWrite))
			testError(t, err, step)
		}
	})
}

func TestWriteDotLinesFromChan(t *testing.T) {
	testScript, stringsToWrite, _ := prepareDotLinesTestScript()
	t.Run(testScript.Name, func(t *testing.T) {
		for _, step := range testScript.Steps {
			rw := NewTestReaderWriterWithTestScript(testScript, NntpReaderWriterOptions{})
			ch := make(chan string, 100)
			for _, line := range stringsToWrite {
				ch <- line
			}
			close(ch)
			err := rw.WriteDotLinesChannel(ch)
			testError(t, err, step)
		}
	})
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
				ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: step.ScriptSteps[0].ExpectedWrite, AwaitNextWrite: true}},
				ExpectedError: step.ExpectedError,
			}
			steps = append(steps, step)
		}
	}
	steps = append(steps, TestStep{ // Final step for the dot line
		ScriptSteps:   []mockScriptedConn.ScriptStep{{ExpectedWrite: []byte(".\r\n")}},
		ExpectedError: nil,
	})
	testScript = TestScript{
		Name:  "WriteDotLinesFromStrings",
		Steps: steps,
	}
	return
}
