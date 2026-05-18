// Package mockScriptedConn provides a mock net.Conn implementation for testing protocol interactions with scripted exchanges.
// It allows defining a sequence of expected writes and corresponding responses, as well as simulating timeouts and connection closure.
package mockScriptedConn

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

// ScriptStep defines one expected write and the response to return on the next read.
type ScriptStep struct {
	ExpectedWrite  []byte // The exact bytes expected to be written (nil to accept any)
	Response       []byte // The bytes to return on the next Read
	WriteTimeout   bool   // If true, the step will cause a write timeout instead of accepting the write
	ReadTimeout    bool   // If true, the step will cause a read timeout instead of providing a response
	AwaitNextWrite bool   // If true, the step will not provide the response until the next write is received
}

// ScriptedConn is a mock net.Conn for protocol testing with a scripted exchange.
type ScriptedConn struct {
	steps         []ScriptStep
	stepIdx       int
	readBuf       *bytes.Buffer
	mu            sync.Mutex
	closed        bool
	writeOk       bool // true if the expected write for this step has been received
	readDeadline  time.Time
	writeDeadline time.Time
}

// NewScriptedConn creates a new ScriptedConn with the given script steps.
func NewScriptedConn(steps []ScriptStep) *ScriptedConn {
	return &ScriptedConn{
		steps:   steps,
		readBuf: &bytes.Buffer{},
	}
}

// TimeoutError is a custom error type that implements net.Error for simulating timeouts in the ScriptedConn.
type TimeoutError struct {
	Op  string
	Net string
	Err error
}

// Timeout returns true to indicate that this error is a timeout.
func (to *TimeoutError) Timeout() bool { return true }

// Temporary returns true to indicate that this error is temporary.
func (to *TimeoutError) Temporary() bool { return true }

// Error returns a string representation of the TimeoutError, including the operation, network, and underlying error message.
func (to *TimeoutError) Error() string { return to.Op + ": " + to.Net + ": " + to.Err.Error() }

// A single instance of TimeoutError to return for all timeout scenarios in the ScriptedConn.
var timeoutError = &TimeoutError{Op: "read", Net: "mock", Err: errors.New("i/o timeout")}

// Read implements the net.Conn Read method, providing scripted responses
// and simulating timeouts as defined in the ScriptStep.
func (c *ScriptedConn) Read(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, io.EOF
	}
	if !c.readDeadline.IsZero() && time.Now().After(c.readDeadline) {
		return 0, timeoutError
	}
	for c.readBuf.Len() == 0 && c.stepIdx < len(c.steps) {
		step := c.steps[c.stepIdx]
		if step.ReadTimeout {
			return 0, timeoutError
		}
		if !c.writeOk {
			// Wait for the expected write before providing a response
			return 0, io.EOF
		}
		resp := step.Response
		if len(resp) > 0 {
			c.readBuf.Write(resp)
		}
		c.stepIdx++
		c.writeOk = false
	}
	return c.readBuf.Read(b)
}

// Write implements the net.Conn Write method, checking against the expected writes in the ScriptStep
// and simulating timeouts as defined in the ScriptStep.
func (c *ScriptedConn) Write(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, io.ErrClosedPipe
	}
	if !c.writeDeadline.IsZero() && time.Now().After(c.writeDeadline) {
		return 0, timeoutError
	}
	if c.stepIdx >= len(c.steps) {
		return 0, errors.New("unexpected write: no more script steps")
	}
	step := c.steps[c.stepIdx]
	if step.WriteTimeout {
		return 0, timeoutError
	}
	if step.ExpectedWrite != nil && !bytes.Equal(b, step.ExpectedWrite) {
		return 0, errors.New("unexpected write: did not match expected")
	}
	if step.AwaitNextWrite {
		c.stepIdx++
		return len(b), nil
	}
	c.writeOk = true
	return len(b), nil
}

// Close implements the net.Conn Close method, marking the connection as closed.
func (c *ScriptedConn) Close() error { c.closed = true; return nil }

// LocalAddr returns nil as this is a mock connection without a real local address.
func (c *ScriptedConn) LocalAddr() net.Addr { return nil }

// RemoteAddr returns nil as this is a mock connection without a real remote address.
func (c *ScriptedConn) RemoteAddr() net.Addr { return nil }

// SetDeadline sets both read and write deadlines for the connection.
func (c *ScriptedConn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readDeadline = t
	c.writeDeadline = t
	return nil
}

// SetReadDeadline sets the read deadline for the connection.
func (c *ScriptedConn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readDeadline = t
	return nil
}

// SetWriteDeadline sets the write deadline for the connection.
func (c *ScriptedConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeDeadline = t
	return nil
}
