package nntpReaderWriter

import (
	"fmt"
	"net"
)

// ErrInvalidResponseLine is returned when the server returns a response line that is too short
// or does not start with a valid 3-digit code.
var ErrInvalidResponseLine = fmt.Errorf("invalid response line")

// ErrInvalidWriteline is returned when attempting to write a line containing a invalid character,
// like a newline character, a carriage return or a null byte.
var ErrInvalidWriteline = func(char string) error {
	return fmt.Errorf("invalid character %q in line to write", char)
}

// ErrNntpError represents an error response (4xx/5xx) from the NNTP server.
type ErrNntpError struct {
	Code int
	Msg  string
}

// ErrUnexpectedResponseCode is returned when the server returns a syntactically valid but unexpected response code.
type ErrUnexpectedResponseCode struct {
	Code int
	Msg  string
}

// Error returns a string representation of the ErrNntpError, including the code and message.
func (e *ErrNntpError) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Msg)
}

// Error returns a string representation of the ErrUnexpectedResponseCode.
func (e *ErrUnexpectedResponseCode) Error() string {
	return fmt.Sprintf("unexpected response code %d %s", e.Code, e.Msg)
}

// IsNntpError returns true if err is a NntpError.
func IsNntpError(err error) bool {
	_, ok := err.(*ErrNntpError)
	return ok
}

// IsUnexpectedResponseCodeError returns true if err is an ErrUnexpectedResponseCode.
func IsUnexpectedResponseCodeError(err error) bool {
	_, ok := err.(*ErrUnexpectedResponseCode)
	return ok
}

// IsInvalidResponseLineError returns true if err is ErrInvalidResponseLine.
func IsInvalidResponseLineError(err error) bool {
	return err == ErrInvalidResponseLine
}

// IsTimeOutError returns true if err is a net.Error with Timeout() == true.
func IsTimeOutError(err error) bool {
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	return false
}

// nntpError returns an *ErrNntpError if code >= 400, otherwise nil.
func nntpError(code int, msg string) error {
	if code >= 400 {
		return &ErrNntpError{Code: code, Msg: msg}
	}
	return nil
}

// nntpUnexpectedCodeError returns an *ErrUnexpectedResponseCode.
func nntpUnexpectedCodeError(code int, msg string) error {
	return &ErrUnexpectedResponseCode{Code: code, Msg: msg}
}
