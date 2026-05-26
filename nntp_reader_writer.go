// Package nntpReaderWriter provides a high-performance, protocol-correct NNTP client reader/writer.
// It supports sequencerd command execution, dot-encoded multi-line responses, and efficient buffer management.
//
// Note: NntpReaderWriter is not safe for concurrent use by multiple goroutines. All access must be externally synchronized.
package nntpReaderWriter

import (
	"net"
	"time"

	"github.com/Tensai75/nntpReaderWriter/lineReader"
)

// NntpReaderWriter provides high-level NNTP command methods and manages protocol state, buffers, and pipelining.
// Not safe for concurrent use by multiple goroutines.
type NntpReaderWriter struct {
	conn         net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
	sequencer    sequencer
	lineReader   *lineReader.LineReader
	lineBuffer   []byte
	writeBuffer  []byte
	code         int
	msg          string
	err          error
}

// NntpReaderWriterOptions configures timeouts for a new NntpReaderWriter.
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
		lineReader:   lineReader.NewLineReader(conn, lineReader.LineReaderOptions{Timeout: options.ReadTimeout}),
		lineBuffer:   make([]byte, 0, 4096),
		writeBuffer:  make([]byte, 0, 4096),
	}
}

// StartSequence starts a new command sequence and returns its ID. The caller must call EndSequence with the same ID when the sequence is complete.
func (rw *NntpReaderWriter) StartSequence() (id uint) {
	id = rw.sequencer.Next()
	rw.sequencer.Start(id)
	return id
}

// EndSequence ends the command sequence with the given ID. This should be called when a command sequence is complete, even if an error occurred, to ensure proper pipelining behavior.
func (rw *NntpReaderWriter) EndSequence(id uint) {
	rw.sequencer.End(id)
}

// CheckCode checks the response code against the expected code and returns any error.
// If the response code indicates an error (4xx or 5xx), it returns an error of type NntpError.
// If the response code does not match the expected code, it returns an error of type UnexpectedResponseCodeError.
// If expectedCode is 0 or negative, it does not check the code against any expected value and only returns an error for 4xx/5xx codes.
// If expectedCode is between 1 and 3, it checks that the code is in the corresponding range (100-199 for 1, 200-299 for 2, etc.).
// If expectedCode is between 10 and 30, it checks that the code is in the corresponding range (100-109 for 10, 110-119 for 11, ..., 500-509 for 50).
// If expectedCode is between 100 and 399, it checks that the code matches exactly.
func (rw *NntpReaderWriter) CheckCode(expectedCode int) (err error) {
	if rw.code >= 400 {
		err = errNntpError(rw.code, rw.msg)
		return
	}
	if expectedCode > 0 && expectedCode < 400 {
		var min, max int
		switch {
		case expectedCode > 0 && expectedCode < 4:
			min, max = expectedCode*100, expectedCode*100+99
		case expectedCode >= 10 && expectedCode < 40:
			min, max = expectedCode*10, expectedCode*10+9
		case expectedCode >= 100 && expectedCode < 400:
			min, max = expectedCode, expectedCode
		}
		if min > 0 && (rw.code < min || rw.code > max) {
			err = errUnexpectedResponseCodeError(rw.code, rw.msg)
			return
		}
	}
	return
}
