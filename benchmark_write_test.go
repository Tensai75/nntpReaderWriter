package nntpReaderWriter

import (
	"bufio"
	"bytes"
	"io"
	"net/textproto"
	"testing"

	mockScriptedConn "github.com/Tensai75/nntpReaderWriter/testutils"
)

const (
	writeLines   = 16384
	writeLineLen = 128
)

func BenchmarkWriteLinesFromReader(b *testing.B) {
	_, bytesToWrite, steps := prepareBenchmark(false)
	reader := bytes.NewReader(bytesToWrite)
	b.SetBytes(int64(writeLines * writeLineLen))
	for b.Loop() {
		client := NewNntpReaderWriter(mockScriptedConn.NewScriptedConn(steps), NntpReaderWriterOptions{})
		reader.Reset(bytesToWrite)
		err := client.writeDotLinesFromReader(reader)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWriteLinesFromReader_Textproto(b *testing.B) {
	_, bytesToWrite, steps := prepareBenchmark(true)
	reader := bytes.NewReader(bytesToWrite)
	b.SetBytes(int64(writeLines * writeLineLen))
	for b.Loop() {
		conn := mockScriptedConn.NewScriptedConn(steps)
		tw := textproto.NewWriter(bufio.NewWriter(conn))
		reader.Reset(bytesToWrite)
		writer := tw.DotWriter()
		_, err := io.Copy(writer, reader)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		writer.Close()
	}
}

func BenchmarkWriteLinesFromStrings(b *testing.B) {
	linesToWrite, _, steps := prepareBenchmark(false)
	b.SetBytes(int64(writeLines * writeLineLen))
	for b.Loop() {
		client := NewNntpReaderWriter(mockScriptedConn.NewScriptedConn(steps), NntpReaderWriterOptions{})
		err := client.writeDotLinesFromStrings(linesToWrite)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWriteLinesFromStrings_Textproto(b *testing.B) {
	linesToWrite, _, steps := prepareBenchmark(true)
	b.SetBytes(int64(writeLines * writeLineLen))
	for b.Loop() {
		conn := mockScriptedConn.NewScriptedConn(steps)
		tw := textproto.NewWriter(bufio.NewWriter(conn))
		writer := tw.DotWriter()
		for _, line := range linesToWrite {
			_, err := writer.Write([]byte(line))
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
		writer.Close()
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
