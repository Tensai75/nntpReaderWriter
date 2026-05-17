package nntpReaderWriter

import (
	"bytes"
	"io"
)

var cr = []byte("\r")
var lf = []byte("\n")

// LineScanner provides buffered line scanning functionality.
// It is designed to efficiently read lines of any length from an io.Reader,
// handling line endings CRLF (\r\n) and LF (\n).
type LineScanner struct {
	rbuf    []byte
	rbufOff int
	rbufN   int
	lbuf    []byte
	err     error
}

// NewLineScanner creates a new LineScanner.
func NewLineScanner() *LineScanner {
	return &LineScanner{
		rbuf: make([]byte, 4*1024),
		lbuf: make([]byte, 0, 512),
	}
}

// ScanLine reads a line using the provided reader, handling CRLF (\r\n) and LF (\n) endings.
// It returns the line without the line ending. If an error occurs during reading, it is returned.
func (ls *LineScanner) ScanLine(reader io.Reader) (line []byte, err error) {
	// If there was a previous read error, we return it.
	if ls.err != nil {
		err = ls.err
		line = append(line, ls.rbuf[ls.rbufOff:ls.rbufN]...)
		if len(line) > 0 && line[len(line)-1] != '\n' {
			// If the last character read was not a newline, we return an unexpected EOF error.
			err = io.ErrUnexpectedEOF
		}
		if len(line) == 0 {
			line = nil
		}
		return
	}
	// Clear the line buffer for the new line.
	line = ls.lbuf[:0]
	for {
		// If the read buffer is exhausted, read more data into it.
		if ls.rbufOff >= ls.rbufN {
			// But if there was a previous read error, we return it.
			if ls.err != nil {
				err = ls.err
				if len(line) > 0 && line[len(line)-1] != '\n' {
					// If the last character read was not a newline, we return an unexpected EOF error.
					err = io.ErrUnexpectedEOF
				}
				if len(line) == 0 {
					line = nil
				}
				return
			}
			ls.rbufN, err = reader.Read(ls.rbuf)
			// If an error occurs during reading, we store it and return it on the next call.
			if err != nil {
				ls.err = err
			}
			// Reset the read buffer offset for the new data.
			ls.rbufOff = 0
		}
		// We set chunk to the unread portion of the read buffer.
		chunk := ls.rbuf[ls.rbufOff:ls.rbufN]
		// We look for a newline character in the chunk.
		if i := bytes.IndexByte(chunk, '\n'); i >= 0 {
			// If we find a newline, we append the line up to and including the newline to the line buffer.
			line = append(line, chunk[:i+1]...)
			// We advance the read buffer offset past the line we just read.
			ls.rbufOff += i + 1
			// We trim any trailing CR and LF characters from the line and return it.
			line = bytes.TrimSuffix(line, lf)
			line = bytes.TrimSuffix(line, cr)
			// We update the line buffer for the next call and return the line.
			ls.lbuf = line[:0]
			return
		}
		// If we don't find a newline, we append the entire chunk to the line buffer.
		line = append(line, chunk...)
		// We advance the read buffer offset to the end of the chunk since we've consumed it all.
		ls.rbufOff = ls.rbufN
	}
}

// Reset clears the internal state of the LineScanner for reuse.
func (ls *LineScanner) Reset() {
	ls.rbufOff = 0
	ls.rbufN = 0
	if ls.lbuf != nil {
		ls.lbuf = ls.lbuf[:0]
	}
	ls.err = nil
}
