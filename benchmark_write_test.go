package nntpReaderWriter

import (
	"bufio"
	"bytes"
	"net/textproto"
	"testing"

	mock "github.com/Tensai75/nntpReaderWriter/testutils"
)

func BenchmarkWriteLinesFromReader(b *testing.B) {
	const (
		lines   = 16384
		lineLen = 128
	)

	b.SetBytes(int64(lines * lineLen))

	// Prepare the scripted response and mock once
	var linesToWrite bytes.Buffer
	var steps []mock.ScriptStep
	line := bytes.Repeat([]byte("X"), lineLen)
	for range lines {
		linesToWrite.Write(line)
		linesToWrite.Write([]byte{'\n'})
		steps = append(steps, mock.ScriptStep{AwaitNextWrite: true})
	}
	steps = append(steps, mock.ScriptStep{})

	for b.Loop() {
		mock := mock.NewScriptedConn(steps)
		opts := NntpReaderWriterOptions{}
		client := NewNntpReaderWriter(mock, opts)

		err := client.writeDotLinesFromReader(&linesToWrite)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWriteLinesFromReader_Textproto(b *testing.B) {
	const (
		lines   = 16384
		lineLen = 128
	)

	b.SetBytes(int64(lines * lineLen))

	// Prepare the scripted response and mock once
	var linesToWrite bytes.Buffer
	var steps []mock.ScriptStep
	line := bytes.Repeat([]byte("X"), lineLen)
	for range lines {
		linesToWrite.Write(line)
		linesToWrite.Write([]byte{'\n'})
		steps = append(steps, mock.ScriptStep{AwaitNextWrite: true})
	}
	steps = append(steps, mock.ScriptStep{})

	for b.Loop() {
		mock := mock.NewScriptedConn(steps)
		tw := textproto.NewWriter(bufio.NewWriter(mock))
		writer := tw.DotWriter()
		_, err := writer.Write(linesToWrite.Bytes())
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		writer.Close()
	}
}

func BenchmarkWriteLinesFromStrings(b *testing.B) {
	const (
		lines   = 16384
		lineLen = 128
	)

	b.SetBytes(int64(lines * lineLen))

	// Prepare the scripted response and mock once
	var linesToWrite []string
	var steps []mock.ScriptStep
	line := bytes.Repeat([]byte("X"), lineLen)
	for range lines {
		linesToWrite = append(linesToWrite, string(line))
		steps = append(steps, mock.ScriptStep{AwaitNextWrite: true})
	}
	steps = append(steps, mock.ScriptStep{})

	for b.Loop() {
		mock := mock.NewScriptedConn(steps)
		opts := NntpReaderWriterOptions{}
		client := NewNntpReaderWriter(mock, opts)

		err := client.writeDotLinesFromStrings(linesToWrite)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWriteLinesFromStrings_Textproto(b *testing.B) {
	const (
		lines   = 16384
		lineLen = 128
	)

	b.SetBytes(int64(lines * lineLen))

	// Prepare the scripted response and mock once
	var linesToWrite []string
	var steps []mock.ScriptStep
	line := bytes.Repeat([]byte("X"), lineLen)
	for range lines {
		linesToWrite = append(linesToWrite, string(line))
		steps = append(steps, mock.ScriptStep{AwaitNextWrite: true})
	}
	steps = append(steps, mock.ScriptStep{})

	for b.Loop() {
		mock := mock.NewScriptedConn(steps)
		tw := textproto.NewWriter(bufio.NewWriter(mock))
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
