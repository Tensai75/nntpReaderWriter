package nntpReaderWriter

import (
	"io"
	"strings"
	"testing"

	mock "github.com/Tensai75/nntpReaderWriter/testutils"
)

var readDotLineTests = []TestScript{
	{
		Name: "Basic",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("hello world\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte("hello world"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "UTF-8",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("こんにちは世界\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte("こんにちは世界"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "VeryLongLine",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte(strings.Repeat("a", 70*1024) + "\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte(strings.Repeat("a", 70*1024)),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "EmptyLine",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte(""),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "TwoDotsToOne",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("..\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte("."),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "DotStuffed",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("..hello\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte(".hello"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "DotStuffedThree",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("...hello\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte("..hello"),
				ExpectedError: nil,
			},
		},
	},
	{
		// CR without LF should be treated as part of the line, not a line terminator
		Name: "CRNoLF",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("foo\r.\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte("foo\r."),
				ExpectedError: nil,
			},
		},
	},
	{
		// Only the last CR before LF should be treated as a line terminator
		Name: "MultipleCR",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("foo\r\r\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte("foo\r\r"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "DotInMiddle",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("foo.bar.baz\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte("foo.bar.baz"),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "WhitespaceOnly",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("   \t  \r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte("   \t  "),
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "SingleDotEOF",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte(".\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedBytes: []byte(""),
				ExpectedError: io.EOF,
			},
		},
	},
}

func TestReadDotLine(t *testing.T) {
	for _, test := range readDotLineTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				rw.writeBytes(step.WriteData)
				line, err := rw.readDotLine()
				testError(t, err, step)
				testBytes(t, line, step)
			}
		})
	}
}

func TestReadDotLines(t *testing.T) {
	var response []byte
	var expectedLines []string
	var expectedBytes []byte
	for test := range readDotLineTests {
		for _, step := range readDotLineTests[test].Steps {
			if step.ExpectedError == io.EOF {
				continue
			}
			response = append(response, step.ScriptSteps[0].Response...)
			expectedLines = append(expectedLines, string(step.ExpectedBytes))
			expectedBytes = append(expectedBytes, append(step.ExpectedBytes, '\n')...)
		}
	}
	testSteps := []TestStep{
		{
			ScriptSteps:   []mock.ScriptStep{{Response: response}},
			WriteData:     []byte("\r\n"),
			ExpectedLines: expectedLines,
			ExpectedBytes: expectedBytes,
			ExpectedError: nil,
		},
	}

	testScriptAsStrings := TestScript{
		Name:  "AsStrings",
		Steps: testSteps,
	}
	t.Run(testScriptAsStrings.Name, func(t *testing.T) {
		rw := NewTestReaderWriterWithTestScript(testScriptAsStrings, NntpReaderWriterOptions{})
		for _, step := range testScriptAsStrings.Steps {
			rw.writeBytes(step.WriteData)
			lines, err := rw.readDotLinesAsStrings()
			testError(t, err, step)
			testLines(t, lines, step)
		}
	})

	testScriptAsReader := TestScript{
		Name:  "AsReader",
		Steps: testSteps,
	}
	t.Run(testScriptAsReader.Name, func(t *testing.T) {
		rw := NewTestReaderWriterWithTestScript(testScriptAsReader, NntpReaderWriterOptions{})
		for _, step := range testScriptAsReader.Steps {
			rw.writeBytes(step.WriteData)
			linesReader, err := rw.readDotLinesAsReader(func() {})
			testError(t, err, step)
			lines, err := io.ReadAll(linesReader)
			if err != nil {
				t.Errorf("unexpected reader error: %v", err)
			}
			testBytes(t, lines, step)
		}
	})

	testScriptAsBytesWithCallback := TestScript{
		Name:  "AsBytesWithCallback",
		Steps: testSteps,
	}
	t.Run(testScriptAsBytesWithCallback.Name, func(t *testing.T) {
		rw := NewTestReaderWriterWithTestScript(testScriptAsBytesWithCallback, NntpReaderWriterOptions{})
		for _, step := range testScriptAsBytesWithCallback.Steps {
			rw.writeBytes(step.WriteData)
			bytes := make([]byte, 0)
			callback := func(line []byte) error {
				bytes = append(bytes, append(line, '\n')...)
				return nil
			}
			err := rw.readDotLinesAsBytesWithCallback(callback)
			testError(t, err, step)
			testBytes(t, bytes, step)

		}
	})

	testScriptAsStringsWithCallback := TestScript{
		Name:  "AsStringsWithCallback",
		Steps: testSteps,
	}
	t.Run(testScriptAsStringsWithCallback.Name, func(t *testing.T) {
		rw := NewTestReaderWriterWithTestScript(testScriptAsStringsWithCallback, NntpReaderWriterOptions{})
		for _, step := range testScriptAsStringsWithCallback.Steps {
			rw.writeBytes(step.WriteData)
			lines := make([]string, 0)
			callback := func(line string) error {
				lines = append(lines, line)
				return nil
			}
			err := rw.readDotLinesAsStringsWithCallback(callback)
			testError(t, err, step)
			testLines(t, lines, step)
		}
	})

}

var readCodeResponseLineTests = []TestScript{
	{
		Name: "Basic",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("200 Hello\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  200,
				ExpectedMsg:   "Hello",
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "UTF-8",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("200 こんにちは\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  200,
				ExpectedMsg:   "こんにちは",
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "NoMessage",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("200\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  200,
				ExpectedMsg:   "",
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "InvalidCode",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("abc Invalid\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  0,
				ExpectedMsg:   "",
				ExpectedError: errInvalidResponseLine(),
			},
		},
	},
	{
		Name: "ShortLine",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("20\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  0,
				ExpectedMsg:   "",
				ExpectedError: errInvalidResponseLine(),
			},
		},
	},
	{
		Name: "WhitespaceInMessage",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("200    Hello World   \r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  200,
				ExpectedMsg:   "   Hello World   ",
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "NonStandardCode",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("599 Custom Code\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  599,
				ExpectedMsg:   "Custom Code",
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "CodeWithLeadingZeros",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("020 Hello\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  20,
				ExpectedMsg:   "Hello",
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "CodeWithTrailingWhitespace",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte("200 Hello   \r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  200,
				ExpectedMsg:   "Hello   ",
				ExpectedError: nil,
			},
		},
	},
	{
		Name: "CodeWithLeadingWhitespace",
		Steps: []TestStep{
			{
				ScriptSteps:   []mock.ScriptStep{{Response: []byte(" 200 Hello\r\n")}},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  0,
				ExpectedMsg:   "",
				ExpectedError: errInvalidResponseLine(),
			},
		},
	},
}

func TestReadCodeResponseLine(t *testing.T) {
	for _, test := range readCodeResponseLineTests {
		t.Run(test.Name, func(t *testing.T) {
			rw := NewTestReaderWriterWithTestScript(test, NntpReaderWriterOptions{})
			for _, step := range test.Steps {
				rw.writeBytes(step.WriteData)
				code, msg, err := rw.readCodeResponseLine()
				testError(t, err, step)
				testCode(t, code, step)
				testMsg(t, msg, step)
			}
		})
	}
}
