package mockScriptedConn

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// ConcurrencyTestConn is a mock net.Conn that simulates concurrent reads and writes for testing purposes.
type ConcurrencyTestConn struct {
	readBuf    *bytes.Buffer
	writeIndex int
	readIndex  int
	mu         sync.Mutex
	closed     bool
}

// NewConcurrencyTestConn creates a new ConcurrencyTestConn.
func NewConcurrencyTestConn() *ConcurrencyTestConn {
	return &ConcurrencyTestConn{
		readBuf: &bytes.Buffer{},
	}
}

// Read implements the net.Conn Read method, providing scripted responses for concurrent reads.
// Each read will return a line with an incrementing number (e.g., "line1\r\n", "line2\r\n", etc.).
func (c *ConcurrencyTestConn) Read(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, io.EOF
	}
	c.readIndex++
	c.readBuf.WriteString("200 line" + fmt.Sprintf("%d", c.readIndex) + "\r\n")
	return c.readBuf.Read(b)
}

// Write implements the net.Conn Write method, checking against the expected writes for concurrent writes.
// Each write is expected to be a line with an incrementing number (e.g., "line1\r\n", "line2\r\n", etc.).
func (c *ConcurrencyTestConn) Write(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, io.ErrClosedPipe
	}
	c.writeIndex++
	if !bytes.Equal(b, []byte("line"+fmt.Sprintf("%d", c.writeIndex)+"\r\n")) {
		return 0, errors.New("unexpected write: wanted line" + fmt.Sprintf("%d", c.writeIndex) + "\r\n" + " but got " + string(b))
	}
	return len(b), nil
}

// Close implements the net.Conn Close method, marking the connection as closed.
func (c *ConcurrencyTestConn) Close() error { c.closed = true; return nil }

// LocalAddr returns nil as this is a mock connection without a real local address.
func (c *ConcurrencyTestConn) LocalAddr() net.Addr { return nil }

// RemoteAddr returns nil as this is a mock connection without a real remote address.
func (c *ConcurrencyTestConn) RemoteAddr() net.Addr { return nil }

// SetDeadline sets both read and write deadlines for the connection.
func (c *ConcurrencyTestConn) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline sets the read deadline for the connection.
func (c *ConcurrencyTestConn) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline sets the write deadline for the connection.
func (c *ConcurrencyTestConn) SetWriteDeadline(t time.Time) error { return nil }
