package nntpReaderWriter

import (
	"bytes"
	"testing"

	"github.com/Tensai75/nntpReaderWriter/mockScriptedConn"
)

const (
	writeLines   = 100000
	writeLineLen = 1024
)

func BenchmarkWriteLinesReader(b *testing.B) {
	_, bytesToWrite, steps := prepareBenchmark(false)
	reader := bytes.NewReader(bytesToWrite)
	b.SetBytes(int64(writeLines * writeLineLen))
	for b.Loop() {
		b.StopTimer()
		client := NewNntpReaderWriter(mockScriptedConn.NewScriptedConn(steps), NntpReaderWriterOptions{})
		reader.Reset(bytesToWrite)
		b.StartTimer()
		err := client.WriteDotLinesReader(reader)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWriteLinesStrings(b *testing.B) {
	linesToWrite, _, steps := prepareBenchmark(false)
	b.SetBytes(int64(writeLines * writeLineLen))
	for b.Loop() {
		b.StopTimer()
		client := NewNntpReaderWriter(mockScriptedConn.NewScriptedConn(steps), NntpReaderWriterOptions{})
		b.StartTimer()
		err := client.WriteDotLines(linesToWrite)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWriteLinesChannel(b *testing.B) {
	linesToWrite, _, steps := prepareBenchmark(false)
	b.SetBytes(int64(writeLines * writeLineLen))
	for b.Loop() {
		b.StopTimer()
		lchan := make(chan string, len(linesToWrite))
		var line string
		for _, line = range linesToWrite {
			lchan <- line
		}
		close(lchan)
		client := NewNntpReaderWriter(mockScriptedConn.NewScriptedConn(steps), NntpReaderWriterOptions{})
		b.StartTimer()
		err := client.WriteDotLinesChannel(lchan)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func prepareBenchmark(textproto bool) (linesToWrite []string, bytesToWrite []byte, steps []mockScriptedConn.ScriptStep) {
	line := bytes.Repeat([]byte("X"), writeLineLen)
	var expectedWrite []byte
	for range writeLines {
		linesToWrite = append(linesToWrite, string(line))
		bytesToWrite = append(bytesToWrite, line...)
		bytesToWrite = append(bytesToWrite, '\n')
		if textproto {
			expectedWrite = nil
		} else {
			expectedWrite = append(line, '\r', '\n')
		}
		steps = append(steps, mockScriptedConn.ScriptStep{
			ExpectedWrite:  expectedWrite,
			AwaitNextWrite: true,
		})
	}
	if textproto {
		expectedWrite = nil
	} else {
		expectedWrite = []byte(".\r\n")
	}
	steps = append(steps, mockScriptedConn.ScriptStep{
		ExpectedWrite: []byte(".\r\n"),
	})
	return
}
