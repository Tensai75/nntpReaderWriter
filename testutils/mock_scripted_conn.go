package mock

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

func NewScriptedConn(steps []ScriptStep) *ScriptedConn {
	return &ScriptedConn{
		steps:   steps,
		readBuf: &bytes.Buffer{},
	}
}

func (c *ScriptedConn) Read(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, io.EOF
	}
	if !c.readDeadline.IsZero() && time.Now().After(c.readDeadline) {
		return 0, &net.OpError{Op: "read", Net: "mock", Err: errors.New("i/o timeout")}
	}
	for c.readBuf.Len() == 0 && c.stepIdx < len(c.steps) {
		step := c.steps[c.stepIdx]
		if step.ReadTimeout {
			return 0, &net.OpError{Op: "read", Net: "mock", Err: errors.New("i/o timeout (step)")}
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

func (c *ScriptedConn) Write(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, io.ErrClosedPipe
	}
	if !c.writeDeadline.IsZero() && time.Now().After(c.writeDeadline) {
		return 0, &net.OpError{Op: "write", Net: "mock", Err: errors.New("i/o timeout")}
	}
	if c.stepIdx >= len(c.steps) {
		return 0, errors.New("unexpected write: no more script steps")
	}
	step := c.steps[c.stepIdx]
	if step.WriteTimeout {
		return 0, &net.OpError{Op: "write", Net: "mock", Err: errors.New("i/o timeout (step)")}
	}
	if step.ExpectedWrite != nil && !bytes.Equal(b, step.ExpectedWrite) {
		return 0, errors.New("unexpected write: did not match expected")
	}
	// fmt.Printf("Expected write: %q\nActual write:   %q\n", step.ExpectedWrite, b)
	if step.AwaitNextWrite {
		c.stepIdx++
		return len(b), nil
	}
	c.writeOk = true
	return len(b), nil
}

func (c *ScriptedConn) Close() error         { c.closed = true; return nil }
func (c *ScriptedConn) LocalAddr() net.Addr  { return nil }
func (c *ScriptedConn) RemoteAddr() net.Addr { return nil }
func (c *ScriptedConn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readDeadline = t
	c.writeDeadline = t
	return nil
}
func (c *ScriptedConn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readDeadline = t
	return nil
}
func (c *ScriptedConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeDeadline = t
	return nil
}
