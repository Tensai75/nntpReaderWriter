package nntpReaderWriter

import (
	"bufio"
	"bytes"
	"io"
	"net/textproto"
	"testing"

	mock "github.com/Tensai75/nntpReaderWriter/testutils"
)

func BenchmarkDotLinesReadCmdAsReader(b *testing.B) {
	const (
		headers = 100000
		lineLen = 1024
	)

	b.SetBytes(int64(headers * lineLen))

	// Prepare the scripted response and mock once
	var resp bytes.Buffer
	resp.WriteString("224 Overview information follows\r\n")
	line := bytes.Repeat([]byte("X"), lineLen-2)
	for range headers {
		resp.Write(line)
		resp.WriteString("\r\n")
	}
	resp.WriteString(".\r\n")
	responseBytes := resp.Bytes()

	for b.Loop() {
		mock := mock.NewScriptedConn([]mock.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: responseBytes},
		})
		opts := NntpReaderWriterOptions{}
		client := NewNntpReaderWriter(mock, opts)

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

func BenchmarkDotLinesReadCmdAsReader_Textproto(b *testing.B) {
	const (
		headers = 100000
		lineLen = 1024
	)
	// Prepare the scripted response and mock once
	var resp bytes.Buffer
	resp.WriteString("224 Overview information follows\r\n")
	line := bytes.Repeat([]byte("X"), lineLen-2)
	for range headers {
		resp.Write(line)
		resp.WriteString("\r\n")
	}
	resp.WriteString(".\r\n")
	responseBytes := resp.Bytes()

	b.SetBytes(int64(headers * lineLen))

	for b.Loop() {
		mock := mock.NewScriptedConn([]mock.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: responseBytes},
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

func BenchmarkDotLinesReadCmdAsLines(b *testing.B) {
	const (
		headers = 100000
		lineLen = 1024
	)

	b.SetBytes(int64(headers * lineLen))

	// Prepare the scripted response and mock once
	var resp bytes.Buffer
	resp.WriteString("224 Overview information follows\r\n")
	line := bytes.Repeat([]byte("X"), lineLen-2)
	for range headers {
		resp.Write(line)
		resp.WriteString("\r\n")
	}
	resp.WriteString(".\r\n")
	responseBytes := resp.Bytes()

	for b.Loop() {
		mock := mock.NewScriptedConn([]mock.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: responseBytes},
		})
		opts := NntpReaderWriterOptions{}
		client := NewNntpReaderWriter(mock, opts)

		code, msg, _, err := client.DotLinesReadCmdAsStrings("OVER 1-100000")
		if err != nil || code != 224 {
			b.Fatalf("unexpected: code=%d msg=%q err=%v", code, msg, err)
		}
	}
}

func BenchmarkDotLinesReadCmdAsLines_Textproto(b *testing.B) {
	const (
		headers = 100000
		lineLen = 1024
	)
	// Prepare the scripted response and mock once
	var resp bytes.Buffer
	resp.WriteString("224 Overview information follows\r\n")
	line := bytes.Repeat([]byte("X"), lineLen-2)
	for range headers {
		resp.Write(line)
		resp.WriteString("\r\n")
	}
	resp.WriteString(".\r\n")
	responseBytes := resp.Bytes()

	b.SetBytes(int64(headers * lineLen))

	for b.Loop() {
		mock := mock.NewScriptedConn([]mock.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: responseBytes},
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

func BenchmarkDotLinesReadCmdAsBytesWithCallback(b *testing.B) {
	const (
		headers = 100000
		lineLen = 1024
	)

	b.SetBytes(int64(headers * lineLen))

	// Prepare the scripted response and mock once
	var resp bytes.Buffer
	resp.WriteString("224 Overview information follows\r\n")
	line := bytes.Repeat([]byte("X"), lineLen-2)
	for range headers {
		resp.Write(line)
		resp.WriteString("\r\n")
	}
	resp.WriteString(".\r\n")
	responseBytes := resp.Bytes()

	for b.Loop() {
		mock := mock.NewScriptedConn([]mock.ScriptStep{
			{ExpectedWrite: []byte("OVER 1-100000\r\n"), Response: responseBytes},
		})
		opts := NntpReaderWriterOptions{}
		client := NewNntpReaderWriter(mock, opts)
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
