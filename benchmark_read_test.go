package nntpReaderWriter

import (
	"bytes"
	"io"
	"testing"

	"github.com/Tensai75/nntpReaderWriter/mockScriptedConn"
)

const (
	readLines   = 100000
	readLineLen = 1024
)

func BenchmarkReadLinesReader(b *testing.B) {
	response := prepareReadBenchmark()
	b.SetBytes(int64(readLines * readLineLen))
	for b.Loop() {
		b.StopTimer()
		mock := mockScriptedConn.NewScriptedConn([]mockScriptedConn.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: response},
		})
		client := NewNntpReaderWriter(mock, NntpReaderWriterOptions{})
		err := client.WriteLine("OVER 1-100000")
		if err != nil {
			b.Fatalf("unexpected error writing command: %v", err)
		}
		_, _, err = client.ReadCodeResponseLine()
		if err != nil {
			b.Fatalf("unexpected error reading code line: err=%v", err)
		}
		b.StartTimer()
		r, err := client.ReadDotLinesReader(nil)
		if err != nil {
			b.Fatalf("unexpected error getting dot lines reader: %v", err)
		}
		_, err = io.Copy(io.Discard, r)
		if err != nil {
			b.Fatalf("io.Copy error: %v", err)
		}
	}
}

func BenchmarkReadLinesStrings(b *testing.B) {
	response := prepareReadBenchmark()
	b.SetBytes(int64(readLines * readLineLen))
	for b.Loop() {
		b.StopTimer()
		mock := mockScriptedConn.NewScriptedConn([]mockScriptedConn.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: response},
		})
		client := NewNntpReaderWriter(mock, NntpReaderWriterOptions{})
		err := client.WriteLine("OVER 1-100000")
		if err != nil {
			b.Fatalf("unexpected error writing command: %v", err)
		}
		_, _, err = client.ReadCodeResponseLine()
		if err != nil {
			b.Fatalf("unexpected error reading code line: err=%v", err)
		}
		b.StartTimer()
		_, err = client.ReadDotLines()
		if err != nil {
			b.Fatalf("unexpected error reading dot lines as strings: %v", err)
		}
	}
}

func BenchmarkReadLinesCallback(b *testing.B) {
	response := prepareReadBenchmark()
	b.SetBytes(int64(readLines * readLineLen))
	for b.Loop() {
		b.StopTimer()
		mock := mockScriptedConn.NewScriptedConn([]mockScriptedConn.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: response},
		})
		client := NewNntpReaderWriter(mock, NntpReaderWriterOptions{})
		callback := func(line string) error {
			// In a real use case, you might process the line here.
			// For benchmarking, we just discard it.
			return nil
		}
		err := client.WriteLine("OVER 1-100000")
		if err != nil {
			b.Fatalf("unexpected error writing command: %v", err)
		}
		_, _, err = client.ReadCodeResponseLine()
		if err != nil {
			b.Fatalf("unexpected error reading code line: err=%v", err)
		}
		b.StartTimer()
		err = client.ReadDotLinesCallback(callback)
		if err != nil {
			b.Fatalf("unexpected error reading dot lines with callback: %v", err)
		}
	}
}

func BenchmarkReadLinesCallbackBytes(b *testing.B) {
	response := prepareReadBenchmark()
	b.SetBytes(int64(readLines * readLineLen))
	for b.Loop() {
		b.StopTimer()
		mock := mockScriptedConn.NewScriptedConn([]mockScriptedConn.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: response},
		})
		client := NewNntpReaderWriter(mock, NntpReaderWriterOptions{})
		callback := func(line []byte) error {
			// In a real use case, you might process the line here.
			// For benchmarking, we just discard it.
			return nil
		}
		err := client.WriteLine("OVER 1-100000")
		if err != nil {
			b.Fatalf("unexpected error writing command: %v", err)
		}
		_, _, err = client.ReadCodeResponseLine()
		if err != nil {
			b.Fatalf("unexpected error reading code line: err=%v", err)
		}
		b.StartTimer()
		err = client.ReadDotLinesCallbackBytes(callback)
		if err != nil {
			b.Fatalf("unexpected error reading dot lines with callback: %v", err)
		}
	}
}

func prepareReadBenchmark() []byte {
	var resp bytes.Buffer
	resp.WriteString("224 Overview information follows\r\n")
	line := bytes.Repeat([]byte("X"), readLineLen-2)
	for range readLines {
		resp.Write(line)
		resp.WriteString("\r\n")
	}
	resp.WriteString(".\r\n")
	return resp.Bytes()
}
