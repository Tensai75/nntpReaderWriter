package nntpReaderWriter

import (
	"errors"
	"fmt"
	"net"
)

// ErrInvalidResponseLine is returned when the server returns a response line that is too short
// or does not start with a valid 3-digit code.
type ErrInvalidResponseLine struct{}

// Error returns a string representation of the ErrInvalidResponseLine error.
func (e *ErrInvalidResponseLine) Error() string {
	return "invalid response line"
}

// IsInvalidResponseLineError returns true if err is an *ErrInvalidResponseLine.
func IsInvalidResponseLineError(err error) bool {
	_, ok := err.(*ErrInvalidResponseLine)
	return ok
}

// errInvalidResponseLine returns a new ErrInvalidResponseLine error.
func errInvalidResponseLine() error {
	return &ErrInvalidResponseLine{}
}

// ErrInvalidWriteLine is returned when attempting to write a line containing an invalid character,
// such as a newline, carriage return, or null byte.
type ErrInvalidWriteLine struct {
	Char string
}

// Error returns a string representation of the ErrInvalidWriteLine error.
func (e *ErrInvalidWriteLine) Error() string {
	return fmt.Sprintf("invalid character %q in line to write", e.Char)
}

// IsInvalidWriteLineError returns true if err is an *ErrInvalidWriteLine.
func IsInvalidWriteLineError(err error) bool {
	_, ok := err.(*ErrInvalidWriteLine)
	return ok
}

// errInvalidWriteLine returns a new ErrInvalidWriteLine error for the given invalid character.
func errInvalidWriteLine(char string) error {
	return &ErrInvalidWriteLine{Char: char}
}

// ErrNntpError represents an error response (4xx/5xx) from the NNTP server.
type ErrNntpError struct {
	Code int    // The NNTP response code
	Msg  string // The error message from the server
}

// Error returns a string representation of the ErrNntpError, including the code and message.
func (e *ErrNntpError) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Msg)
}

// IsNntpError returns true if err is an *ErrNntpError.
func IsNntpError(err error) bool {
	_, ok := err.(*ErrNntpError)
	return ok
}

// errNntpError returns an *ErrNntpError if code >= 400, otherwise nil.
func errNntpError(code int, msg string) error {
	if code >= 400 {
		return &ErrNntpError{Code: code, Msg: msg}
	}
	return nil
}

// ErrUnexpectedResponseCode is returned when the server returns a syntactically valid but unexpected response code.
type ErrUnexpectedResponseCode struct {
	Code int    // The unexpected NNTP response code
	Msg  string // The message associated with the response code
}

// Error returns a string representation of the ErrUnexpectedResponseCode.
func (e *ErrUnexpectedResponseCode) Error() string {
	return fmt.Sprintf("unexpected response code %d %s", e.Code, e.Msg)
}

// IsUnexpectedResponseCodeError returns true if err is an *ErrUnexpectedResponseCode.
func IsUnexpectedResponseCodeError(err error) bool {
	_, ok := err.(*ErrUnexpectedResponseCode)
	return ok
}

// errUnexpectedResponseCodeError returns an *ErrUnexpectedResponseCode for unexpected NNTP response codes.
func errUnexpectedResponseCodeError(code int, msg string) error {
	return &ErrUnexpectedResponseCode{Code: code, Msg: msg}
}

// IsTimeOutError returns true if err is a net.Error with Timeout() == true.
// This can be used to check for network timeout errors in a generic way.
func IsTimeOutError(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	return false
}
