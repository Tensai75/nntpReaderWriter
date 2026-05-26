package nntpReaderWriter

import (
	"bytes"
	"io"
	"strings"
	"time"
)

// WriteDotLines writes the provided lines as dot-encoded lines.
// It handles dot-stuffing for lines that start with a dot and ensures that each line is properly terminated with CRLF.
// After writing all lines, it writes a single dot on a line by itself to indicate the end of the block.
// If a line contains any invalid characters (CR, LF, or NUL), it returns an error.
func (rw *NntpReaderWriter) WriteDotLines(lines []string) error {
	for _, line := range lines {
		if err := rw.WriteDotLine(line); err != nil {
			return err
		}
	}
	return rw.WriteLine(".")
}

// WriteDotLinesReader writes lines from the provided io.Reader as dot-encoded lines.
// It handles dot-stuffing for lines that start with a dot and ensures that each line is properly terminated with CRLF.
// After writing all lines, it writes a single dot on a line by itself to indicate the end of the block.
// If a line contains any invalid characters (CR, LF, or NUL), it returns an error.
func (rw *NntpReaderWriter) WriteDotLinesReader(r io.Reader) error {
	rw.lineReader.Reset()
	for {
		line, err := rw.lineReader.ReadLineBuffered(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := rw.WriteDotLineBytes(line); err != nil {
			return err
		}
	}
	return rw.WriteLine(".")
}

// WriteDotLinesChannel writes lines from the provided channel as dot-encoded lines.
// It handles dot-stuffing for lines that start with a dot and ensures that each line is properly terminated with CRLF.
// After writing all lines, it writes a single dot on a line by itself to indicate the end of the block.
// If a line contains any invalid characters (CR, LF, or NUL), it returns an error.
func (rw *NntpReaderWriter) WriteDotLinesChannel(ch chan string) error {
	var err error
	for line := range ch {
		err = rw.WriteDotLine(line)
		if err != nil {
			return err
		}
	}
	err = rw.WriteLine(".")
	return err
}

// WriteDotLine writes a single line as a dot-encoded line, handling dot-stuffing if necessary.
// If the line contains any invalid characters (CR, LF, or NUL), it returns an error.
func (rw *NntpReaderWriter) WriteDotLine(line string) error {
	prefix := ""
	if len(line) > 0 && line[0] == '.' {
		prefix = "."
	}
	return rw.WriteLine(line, prefix)
}

// WriteDotLineBytes writes a single line as a dot-encoded line, handling dot-stuffing if necessary.
// If the line contains any invalid characters (CR, LF, or NUL), it returns an error.
func (rw *NntpReaderWriter) WriteDotLineBytes(line []byte) error {
	prefix := ""
	if len(line) > 0 && line[0] == '.' {
		prefix = "."
	}
	return rw.WriteLineBytes(line, prefix)
}

// WriteLine writes a single line from a string, ensuring it is properly terminated with CRLF.
// It also accepts optional prefixes that can be used for dot-stuffing.
// If the line contains any invalid characters (CR, LF, or NUL), it returns an error.
func (rw *NntpReaderWriter) WriteLine(line string, prefixes ...string) error {
	if err := checkStringLineForInvalidChars(line); err != nil {
		return err
	}
	rw.writeBuffer = rw.writeBuffer[:0]
	for _, p := range prefixes {
		rw.writeBuffer = append(rw.writeBuffer, p...)
	}
	rw.writeBuffer = append(rw.writeBuffer, line...)
	rw.writeBuffer = append(rw.writeBuffer, '\r', '\n')
	_, err := rw.writeBytes(rw.writeBuffer)
	if err == io.EOF {
		err = nil
	}
	return err
}

// WriteLineBytes writes a single line from a byte slice, ensuring it is properly terminated with CRLF.
// It also accepts optional prefixes that can be used for dot-stuffing.
// If the line contains any invalid characters (CR, LF, or NUL), it returns an error.
func (rw *NntpReaderWriter) WriteLineBytes(line []byte, prefixes ...string) error {
	if err := checkBytesLineForInvalidChars(line); err != nil {
		return err
	}
	rw.writeBuffer = rw.writeBuffer[:0]
	for _, p := range prefixes {
		rw.writeBuffer = append(rw.writeBuffer, p...)
	}
	rw.writeBuffer = append(rw.writeBuffer, line...)
	rw.writeBuffer = append(rw.writeBuffer, '\r', '\n')
	_, err := rw.writeBytes(rw.writeBuffer)
	if err == io.EOF {
		err = nil
	}
	return err
}

func (rw *NntpReaderWriter) writeBytes(p []byte) (n int, err error) {
	if rw.writeTimeout > 0 {
		rw.conn.SetWriteDeadline(time.Now().Add(rw.writeTimeout))
	}
	return rw.conn.Write(p)
}

func checkStringLineForInvalidChars(line string) error {
	if strings.Contains(line, "\n") {
		return errInvalidWriteLine("\\n")
	}
	if strings.Contains(line, "\r") {
		return errInvalidWriteLine("\\r")
	}
	if strings.Contains(line, "\x00") {
		return errInvalidWriteLine("\\x00")
	}
	return nil
}

func checkBytesLineForInvalidChars(line []byte) error {
	if bytes.Contains(line, []byte{'\n'}) {
		return errInvalidWriteLine("\\n")
	}
	if bytes.Contains(line, []byte{'\r'}) {
		return errInvalidWriteLine("\\r")
	}
	if bytes.Contains(line, []byte{'\x00'}) {
		return errInvalidWriteLine("\\x00")
	}
	return nil
}
