package mockScriptedConn

import (
	"errors"
	"io"
	"testing"
	"time"
)

func TestScriptedConn_ReadWrite_Success(t *testing.T) {
	steps := []ScriptStep{
		{ExpectedWrite: []byte("hello"), Response: []byte("world")},
		{ExpectedWrite: []byte("foo"), Response: []byte("bar")},
	}
	conn := NewScriptedConn(steps)

	// Write and read first step
	n, err := conn.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("Write failed: n=%d, err=%v", n, err)
	}
	buf := make([]byte, 8)
	n, err = conn.Read(buf)
	if err != nil || string(buf[:n]) != "world" {
		t.Fatalf("Read failed: n=%d, err=%v, got=%q", n, err, buf[:n])
	}

	// Write and read second step
	n, err = conn.Write([]byte("foo"))
	if err != nil || n != 3 {
		t.Fatalf("Write failed: n=%d, err=%v", n, err)
	}
	n, err = conn.Read(buf)
	if err != nil || string(buf[:n]) != "bar" {
		t.Fatalf("Read failed: n=%d, err=%v, got=%q", n, err, buf[:n])
	}
}

func TestScriptedConn_Timeouts(t *testing.T) {
	steps := []ScriptStep{
		{ExpectedWrite: []byte("hi"), Response: []byte("ok"), WriteTimeout: true},
		{ExpectedWrite: []byte("next"), Response: []byte("done"), ReadTimeout: true},
	}
	conn := NewScriptedConn(steps)

	// WriteTimeout
	_, err := conn.Write([]byte("hi"))
	if err == nil || !isTimeoutErr(err) {
		t.Errorf("expected write timeout error, got %v", err)
	}

	// Advance to next step for read timeout
	conn.stepIdx = 1
	conn.writeOk = true
	_, err = conn.Read(make([]byte, 8))
	if err == nil || !isTimeoutErr(err) {
		t.Errorf("expected read timeout error, got %v", err)
	}
}

func TestScriptedConn_UnexpectedWrite(t *testing.T) {
	steps := []ScriptStep{
		{ExpectedWrite: []byte("abc"), Response: []byte("123")},
	}
	conn := NewScriptedConn(steps)
	_, err := conn.Write([]byte("xyz"))
	if err == nil || err.Error() != "unexpected write: did not match expected" {
		t.Errorf("expected unexpected write error, got %v", err)
	}
}

func TestScriptedConn_Closed(t *testing.T) {
	steps := []ScriptStep{
		{ExpectedWrite: []byte("a"), Response: []byte("b")},
	}
	conn := NewScriptedConn(steps)
	_ = conn.Close()
	_, err := conn.Write([]byte("a"))
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Errorf("expected ErrClosedPipe, got %v", err)
	}
	_, err = conn.Read(make([]byte, 8))
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestScriptedConn_Deadlines(t *testing.T) {
	steps := []ScriptStep{
		{ExpectedWrite: []byte("a"), Response: []byte("b")},
	}
	conn := NewScriptedConn(steps)
	conn.SetReadDeadline(time.Now().Add(-time.Second))
	_, err := conn.Read(make([]byte, 8))
	if err == nil || !isTimeoutErr(err) {
		t.Errorf("expected read deadline timeout, got %v", err)
	}
	conn.SetWriteDeadline(time.Now().Add(-time.Second))
	_, err = conn.Write([]byte("a"))
	if err == nil || !isTimeoutErr(err) {
		t.Errorf("expected write deadline timeout, got %v", err)
	}
}

func isTimeoutErr(err error) bool {
	t, ok := err.(interface{ Timeout() bool })
	return ok && t.Timeout()
}
