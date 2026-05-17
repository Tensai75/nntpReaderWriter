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
				ScriptStep:       mock.ScriptStep{Response: []byte("hello world\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte("hello world"),
				ExpectedError:    nil,
			},
		},
	},
	{
		Name: "UTF-8",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte("こんにちは世界\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte("こんにちは世界"),
				ExpectedError:    nil,
			},
		},
	},
	{
		Name: "VeryLongLine",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte(strings.Repeat("a", 70*1024) + "\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte(strings.Repeat("a", 70*1024)),
				ExpectedError:    nil,
			},
		},
	},
	{
		Name: "EmptyLine",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte("\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte(""),
				ExpectedError:    nil,
			},
		},
	},
	{
		Name: "TwoDotsToOne",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte("..\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte("."),
				ExpectedError:    nil,
			},
		},
	},
	{
		Name: "DotStuffed",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte("..hello\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte(".hello"),
				ExpectedError:    nil,
			},
		},
	},
	{
		Name: "DotStuffedThree",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte("...hello\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte("..hello"),
				ExpectedError:    nil,
			},
		},
	},
	{
		// CR without LF should be treated as part of the line, not a line terminator
		Name: "CRNoLF",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte("foo\r.\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte("foo\r."),
				ExpectedError:    nil,
			},
		},
	},
	{
		// Only the last CR before LF should be treated as a line terminator
		Name: "MultipleCR",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte("foo\r\r\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte("foo\r\r"),
				ExpectedError:    nil,
			},
		},
	},
	{
		Name: "DotInMiddle",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte("foo.bar.baz\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte("foo.bar.baz"),
				ExpectedError:    nil,
			},
		},
	},
	{
		Name: "WhitespaceOnly",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte("   \t  \r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte("   \t  "),
				ExpectedError:    nil,
			},
		},
	},
	{
		Name: "SingleDotEOF",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: []byte(".\r\n")},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: []byte(""),
				ExpectedError:    io.EOF,
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
				if err != step.ExpectedError || string(line) != string(step.ExpectedReadData) {
					t.Errorf("readDotLine(%q) = %q, err=%v; want %q", test.Name, line, err, step.ExpectedReadData)
				}
			}
		})
	}
}

func TestReadDotLines(t *testing.T) {
	var response []byte
	var ExpectedLines []string
	var ReadData []byte
	for test := range readDotLineTests {
		for _, step := range readDotLineTests[test].Steps {
			if step.ExpectedError == io.EOF {
				continue
			}
			response = append(response, step.ScriptStep.Response...)
			ExpectedLines = append(ExpectedLines, string(step.ExpectedReadData))
			ReadData = append(ReadData, append(step.ExpectedReadData, '\n')...)
		}
	}
	testScriptAsStrings := TestScript{
		Name: "AsStrings",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{Response: response},
				WriteData:     []byte("\r\n"),
				ExpectedLines: ExpectedLines,
				ExpectedError: nil,
			},
		},
	}
	t.Run(testScriptAsStrings.Name, func(t *testing.T) {
		rw := NewTestReaderWriterWithTestScript(testScriptAsStrings, NntpReaderWriterOptions{})
		for _, step := range testScriptAsStrings.Steps {
			rw.writeBytes(step.WriteData)
			lines, err := rw.readDotLinesAsStrings()
			if err != step.ExpectedError {
				t.Errorf("readDotLines(%q) error = %v; want %v", testScriptAsStrings.Name, err, step.ExpectedError)
			}
			for i, line := range lines {
				if line != step.ExpectedLines[i] {
					t.Errorf("readDotLines(%q) line %d = %q; want %q", testScriptAsStrings.Name, i, line, step.ExpectedLines[i])
				}
			}
		}
	})

	testScriptAsReader := TestScript{
		Name: "AsReader",
		Steps: []TestStep{
			{
				ScriptStep:       mock.ScriptStep{Response: response},
				WriteData:        []byte("\r\n"),
				ExpectedReadData: ReadData,
				ExpectedError:    nil,
			},
		},
	}
	t.Run(testScriptAsReader.Name, func(t *testing.T) {
		rw := NewTestReaderWriterWithTestScript(testScriptAsReader, NntpReaderWriterOptions{})
		for _, step := range testScriptAsReader.Steps {
			rw.writeBytes(step.WriteData)
			linesReader, err := rw.readDotLinesAsReader(func() {})
			if err != step.ExpectedError {
				t.Errorf("readDotLines(%q) error = %v; want %v", testScriptAsReader.Name, err, step.ExpectedError)
			}
			lines, err := io.ReadAll(linesReader)
			if err != step.ExpectedError {
				t.Errorf("readDotLines(%q) error = %v; want %v", testScriptAsReader.Name, err, step.ExpectedError)
			}
			if string(lines) != string(step.ExpectedReadData) {
				t.Errorf("readDotLines(%q) = %q; want %q", testScriptAsReader.Name, lines, step.ExpectedReadData)
			}

		}
	})

	testScriptAsBytesWithCallback := TestScript{
		Name: "AsBytesWithCallback",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{Response: response},
				WriteData:     []byte("\r\n"),
				ExpectedLines: ExpectedLines,
				ExpectedError: nil,
			},
		},
	}
	t.Run(testScriptAsBytesWithCallback.Name, func(t *testing.T) {
		rw := NewTestReaderWriterWithTestScript(testScriptAsBytesWithCallback, NntpReaderWriterOptions{})
		for _, step := range testScriptAsBytesWithCallback.Steps {
			rw.writeBytes(step.WriteData)
			i := 0
			callback := func(line []byte) error {
				if string(line) != step.ExpectedLines[i] {
					t.Errorf("readDotLines(%q) line %d = %q; want %q", testScriptAsBytesWithCallback.Name, i, line, step.ExpectedLines[i])
				}
				i++
				return nil
			}
			err := rw.readDotLinesAsBytesWithCallback(callback)
			if err != step.ExpectedError {
				t.Errorf("readDotLines(%q) error = %v; want %v", testScriptAsBytesWithCallback.Name, err, step.ExpectedError)
			}

		}
	})

	testScriptAsStringsWithCallback := TestScript{
		Name: "AsStringsWithCallback",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{Response: response},
				WriteData:     []byte("\r\n"),
				ExpectedLines: ExpectedLines,
				ExpectedError: nil,
			},
		},
	}
	t.Run(testScriptAsStringsWithCallback.Name, func(t *testing.T) {
		rw := NewTestReaderWriterWithTestScript(testScriptAsStringsWithCallback, NntpReaderWriterOptions{})
		for _, step := range testScriptAsStringsWithCallback.Steps {
			rw.writeBytes(step.WriteData)
			i := 0
			callback := func(line string) error {
				if line != step.ExpectedLines[i] {
					t.Errorf("readDotLines(%q) line %d = %q; want %q", testScriptAsStringsWithCallback.Name, i, line, step.ExpectedLines[i])
				}
				i++
				return nil
			}
			err := rw.readDotLinesAsStringsWithCallback(callback)
			if err != step.ExpectedError {
				t.Errorf("readDotLines(%q) error = %v; want %v", testScriptAsStringsWithCallback.Name, err, step.ExpectedError)
			}

		}
	})

}

var readCodeResponseLineTests = []TestScript{
	{
		Name: "Basic",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{Response: []byte("200 Hello\r\n")},
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
				ScriptStep:    mock.ScriptStep{Response: []byte("200 こんにちは\r\n")},
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
				ScriptStep:    mock.ScriptStep{Response: []byte("200\r\n")},
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
				ScriptStep:    mock.ScriptStep{Response: []byte("abc Invalid\r\n")},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  0,
				ExpectedMsg:   "",
				ExpectedError: ErrInvalidResponseLine,
			},
		},
	},
	{
		Name: "ShortLine",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{Response: []byte("20\r\n")},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  0,
				ExpectedMsg:   "",
				ExpectedError: ErrInvalidResponseLine,
			},
		},
	},
	{
		Name: "WhitespaceInMessage",
		Steps: []TestStep{
			{
				ScriptStep:    mock.ScriptStep{Response: []byte("200    Hello World   \r\n")},
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
				ScriptStep:    mock.ScriptStep{Response: []byte("599 Custom Code\r\n")},
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
				ScriptStep:    mock.ScriptStep{Response: []byte("020 Hello\r\n")},
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
				ScriptStep:    mock.ScriptStep{Response: []byte("200 Hello   \r\n")},
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
				ScriptStep:    mock.ScriptStep{Response: []byte(" 200 Hello\r\n")},
				WriteData:     []byte("\r\n"),
				ExpectedCode:  0,
				ExpectedMsg:   "",
				ExpectedError: ErrInvalidResponseLine,
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
				if err != step.ExpectedError || code != step.ExpectedCode || msg != step.ExpectedMsg {
					t.Errorf("readCodeResponseLine(%q) = code=%d msg=%q err=%v; want code=%d msg=%q err=%v",
						test.Name, code, msg, err, step.ExpectedCode, step.ExpectedMsg, step.ExpectedError)
				}
			}
		})
	}
}
