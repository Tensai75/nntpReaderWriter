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
	readLines   = 16384 //100000
	readLineLen = 128   //1024
)

func BenchmarkReadLinesAsReader(b *testing.B) {
	response := prepareReadBenchmark()
	b.SetBytes(int64(readLines * readLineLen))
	for b.Loop() {
		mock := mockScriptedConn.NewScriptedConn([]mockScriptedConn.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: response},
		})
		client := NewNntpReaderWriter(mock, NntpReaderWriterOptions{})
		code, msg, r, err := client.DotLinesReadCmdAsReader("OVER 1-100000")
		if err != nil || code != 224 {
			b.Fatalf("unexpected: code=%d msg=%q err=%v", code, msg, err)
		}
		_, err = io.Copy(io.Discard, r)
		if err != nil {
			b.Fatalf("io.Copy error: %v", err)
		}
		r.Close()
	}
}

func BenchmarkReadLinesAsReader_Textproto(b *testing.B) {
	response := prepareReadBenchmark()
	b.SetBytes(int64(readLines * readLineLen))
	for b.Loop() {
		mock := mockScriptedConn.NewScriptedConn([]mockScriptedConn.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: response},
		})
		tr := textproto.NewReader(bufio.NewReaderSize(mock, 16*1024))
		// Write the OVER command (simulate client send)
		_, _ = mock.Write([]byte("OVER 1-100000\r\n"))
		// Read the status line
		code, msg, err := tr.ReadCodeLine(224)
		if err != nil && code != 224 {
			b.Fatalf("unexpected: code=%d msg=%q err=%v", code, msg, err)
		}
		if err != nil {
			b.Fatalf("ReadLine error: %v", err)
		}
		// Read dot-encoded lines using DotReader and copy to io.Discard
		dr := tr.DotReader()
		_, err = io.Copy(io.Discard, dr)
		if err != nil {
			b.Fatalf("io.Copy error: %v", err)
		}
	}
}

func BenchmarkReadLinesAsStrings(b *testing.B) {
	response := prepareReadBenchmark()
	b.SetBytes(int64(readLines * readLineLen))
	for b.Loop() {
		mock := mockScriptedConn.NewScriptedConn([]mockScriptedConn.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: response},
		})
		client := NewNntpReaderWriter(mock, NntpReaderWriterOptions{})
		code, msg, _, err := client.DotLinesReadCmdAsStrings("OVER 1-100000")
		if err != nil || code != 224 {
			b.Fatalf("unexpected: code=%d msg=%q err=%v", code, msg, err)
		}
	}
}

func BenchmarkReadLinesAsStrings_Textproto(b *testing.B) {
	response := prepareReadBenchmark()
	b.SetBytes(int64(readLines * readLineLen))
	for b.Loop() {
		mock := mockScriptedConn.NewScriptedConn([]mockScriptedConn.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: response},
		})
		tr := textproto.NewReader(bufio.NewReader(mock))
		// Write the OVER command (simulate client send)
		_, _ = mock.Write([]byte("OVER 1-100000\r\n"))
		// Read the status line
		code, msg, err := tr.ReadCodeLine(224)
		if err != nil && code != 224 {
			b.Fatalf("unexpected: code=%d msg=%q err=%v", code, msg, err)
		}
		if err != nil {
			b.Fatalf("ReadLine error: %v", err)
		}
		// Read dot-encoded lines using DotReader and copy to io.Discard
		_, err = tr.ReadDotLines()
		if err != nil {
			b.Fatalf("io.Copy error: %v", err)
		}
	}
}

func BenchmarkReadLinesAsBytesWithCallback(b *testing.B) {
	response := prepareReadBenchmark()
	b.SetBytes(int64(readLines * readLineLen))
	for b.Loop() {
		mock := mockScriptedConn.NewScriptedConn([]mockScriptedConn.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: response},
		})
		client := NewNntpReaderWriter(mock, NntpReaderWriterOptions{})
		callback := func(line []byte) error {
			// In a real use case, you might process the line here.
			// For benchmarking, we just discard it.
			return nil
		}
		code, msg, err := client.DotLinesReadCmdAsBytesWithCallback("OVER 1-100000", callback)
		if err != nil || code != 224 {
			b.Fatalf("unexpected: code=%d msg=%q err=%v", code, msg, err)
		}
	}
}

func BenchmarkReadLinesAsStringsWithCallback(b *testing.B) {
	response := prepareReadBenchmark()
	b.SetBytes(int64(readLines * readLineLen))
	for b.Loop() {
		mock := mockScriptedConn.NewScriptedConn([]mockScriptedConn.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: response},
		})
		client := NewNntpReaderWriter(mock, NntpReaderWriterOptions{})
		callback := func(line string) error {
			// In a real use case, you might process the line here.
			// For benchmarking, we just discard it.
			return nil
		}
		code, msg, err := client.DotLinesReadCmdAsStringsWithCallback("OVER 1-100000", callback)
		if err != nil || code != 224 {
			b.Fatalf("unexpected: code=%d msg=%q err=%v", code, msg, err)
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
