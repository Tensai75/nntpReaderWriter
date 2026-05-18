// Package nntpReaderWriter provides a high-performance, protocol-correct NNTP client reader/writer.
// It supports pipelined command execution, dot-encoded multi-line responses, and efficient buffer management.
//
// Note: NntpReaderWriter is not safe for concurrent use by multiple goroutines. All access must be externally synchronized.
package nntpReaderWriter

import (
	"io"
	"net"
	"time"
)

// NntpReaderWriter provides high-level NNTP command methods and manages protocol state, buffers, and pipelining.
// Not safe for concurrent use by multiple goroutines.
type NntpReaderWriter struct {
	conn         net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
	pipeline     pipeline
	lineScanner  *LineScanner
	wbuf         []byte
}

// NntpReaderWriterOptions configures timeouts and line length limits for a new NntpReaderWriter.
type NntpReaderWriterOptions struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewNntpReaderWriter creates a new NNTP reader/writer using the given connection and options.
// The returned instance is not safe for concurrent use.
func NewNntpReaderWriter(conn net.Conn, options NntpReaderWriterOptions) *NntpReaderWriter {
	return &NntpReaderWriter{
		conn:         conn,
		readTimeout:  options.ReadTimeout,
		writeTimeout: options.WriteTimeout,
		lineScanner:  NewLineScanner(),
		wbuf:         make([]byte, 0, 512),
	}
}

// SingleResponseLineCmd sends a command and returns the response code and message for single-line NNTP responses.
func (rw *NntpReaderWriter) SingleResponseLineCmd(cmd string) (code int, msg string, err error) {
	var id uint
	id, code, msg, err = rw.startCmd(cmd)
	if err != nil {
		return
	}
	defer rw.pipeline.End(id)
	if code >= 300 && code < 400 {
		err = errUnexpectedResponseCodeError(code, msg)
	}
	return
}

// DotLinesReadCmdAsStrings sends a command and returns the response code, message, and all dot-encoded lines as strings.
func (rw *NntpReaderWriter) DotLinesReadCmdAsStrings(cmd string) (code int, msg string, lines []string, err error) {
	var id uint
	id, code, msg, err = rw.startCmd(cmd)
	if err != nil {
		return
	}
	defer rw.pipeline.End(id)
	lines, err = rw.readDotLinesAsStrings()
	return
}

// DotLinesReadCmdAsReader sends a command and returns the response code, message, and an io.ReadCloser for dot-encoded lines.
func (rw *NntpReaderWriter) DotLinesReadCmdAsReader(cmd string) (code int, msg string, r io.ReadCloser, err error) {
	var id uint
	id, code, msg, err = rw.startCmd(cmd)
	if err != nil {
		return
	}
	r, err = rw.readDotLinesAsReader(func() { rw.pipeline.End(id) })
	if err != nil {
		rw.pipeline.End(id)
	}
	return
}

// DotLinesReadCmdAsBytesWithCallback sends a command and invokes the callback for each dot-encoded line as bytes.
func (rw *NntpReaderWriter) DotLinesReadCmdAsBytesWithCallback(cmd string, callback func([]byte) error) (code int, msg string, err error) {
	var id uint
	id, code, msg, err = rw.startCmd(cmd)
	if err != nil {
		return
	}
	err = rw.readDotLinesAsBytesWithCallback(callback)
	if err != nil {
		rw.pipeline.End(id)
	}
	return
}

// DotLinesReadCmdAsStringsWithCallback sends a command and invokes the callback for each dot-encoded line as a string.
func (rw *NntpReaderWriter) DotLinesReadCmdAsStringsWithCallback(cmd string, callback func(string) error) (code int, msg string, err error) {
	var id uint
	id, code, msg, err = rw.startCmd(cmd)
	if err != nil {
		return
	}
	err = rw.readDotLinesAsStringsWithCallback(callback)
	if err != nil {
		rw.pipeline.End(id)
	}
	return
}

// DotLinesWriteCmdFromStrings sends a command and writes the provided lines as a dot-encoded block.
func (rw *NntpReaderWriter) DotLinesWriteCmdFromStrings(cmd string, lines []string) (code int, msg string, err error) {
	return rw.dotLinesWriteCmd(cmd, func() error {
		return rw.writeDotLinesFromStrings(lines)
	})
}

// DotLinesWriteCmdFromReader sends a command and writes data from the provided reader as a dot-encoded block.
func (rw *NntpReaderWriter) DotLinesWriteCmdFromReader(cmd string, r io.Reader) (code int, msg string, err error) {
	return rw.dotLinesWriteCmd(cmd, func() error {
		return rw.writeDotLinesFromReader(r)
	})
}

// DotLinesWriteCmdFromChan sends a command and writes lines received from the provided channel as a dot-encoded block.
func (rw *NntpReaderWriter) DotLinesWriteCmdFromChan(cmd string, ch chan string) (code int, msg string, err error) {
	return rw.dotLinesWriteCmd(cmd, func() error {
		return rw.writeDotLinesFromChan(ch)
	})
}

// ReadCodeResponseLine reads a single-line NNTP response and returns the response code and message.
func (rw *NntpReaderWriter) ReadCodeResponseLine() (code int, msg string, err error) {
	id := rw.pipeline.Next()
	rw.pipeline.Start(id)
	defer rw.pipeline.End(id)
	rw.lineScanner.Reset()
	return rw.readCodeResponseLine()
}

func (rw *NntpReaderWriter) startCmd(cmd string) (id uint, code int, msg string, err error) {
	id = rw.pipeline.Next()
	rw.pipeline.Start(id)
	if err = rw.writeLineFromString(cmd); err != nil {
		rw.pipeline.End(id)
		return
	}
	rw.lineScanner.Reset()
	code, msg, err = rw.readCodeResponseLine()
	if err == nil && code >= 400 {
		err = errNntpError(code, msg)
	}
	if err != nil {
		rw.pipeline.End(id)
		return
	}
	return
}

func (rw *NntpReaderWriter) dotLinesWriteCmd(cmd string, write func() error) (code int, msg string, err error) {
	var id uint
	id, code, msg, err = rw.startCmd(cmd)
	if err != nil {
		return
	}
	defer rw.pipeline.End(id)
	err = errNntpError(code, msg)
	if err != nil {
		return
	}
	if code < 300 {
		return code, msg, errUnexpectedResponseCodeError(code, msg)
	}
	err = write()
	if err != nil && err != io.EOF {
		return 0, "", err
	}
	rw.lineScanner.Reset()
	code, msg, err = rw.readCodeResponseLine()
	if err != nil {
		return
	}
	err = errNntpError(code, msg)
	return
}
