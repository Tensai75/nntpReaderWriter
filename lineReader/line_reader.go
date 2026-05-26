// Package lineReader provides a utility for reading lines from a io.Reader with support for both buffered and unbuffered reading.
// If the io.Reader supports setting read deadlines, the LineReader can also be configured with a timeout for read operations.
package lineReader

import (
	"bytes"
	"errors"
	"io"
	"time"
)

// LineReader is a utility for reading lines from a io.Reader with support for both buffered and unbuffered reading.
// If the io.Reader supports setting read deadlines, the LineReader can also be configured with a timeout for read operations.
type LineReader struct {
	reader     io.Reader
	timeout    time.Duration
	singleByte []byte
	buffer     []byte
	lineBuffer []byte
	bytes      int
	offset     int
	lines      int
	err        error
}

// LineReaderOptions configures the timeout and buffer size for a LineReader.
type LineReaderOptions struct {
	Timeout    time.Duration // Timeout for read operations. A zero value means no timeout.
	BufferSize int           // Size of the buffer used for buffered reading. A zero or negative value will default to 16KB.
}

const (
	cr = '\r'
	lf = '\n'
)

// NewLineReader creates a new LineReader with the specified io.Reader and options.
func NewLineReader(r io.Reader, options LineReaderOptions) *LineReader {
	bufferSize := options.BufferSize
	if bufferSize <= 0 {
		bufferSize = 16 * 1024
	}
	return &LineReader{
		timeout:    options.Timeout,
		reader:     r,
		singleByte: make([]byte, 1),
		buffer:     make([]byte, bufferSize),
		lineBuffer: make([]byte, 512),
	}
}

// SetReader allows setting or changing the underlying io.Reader for the LineReader.
func (r *LineReader) SetReader(reader io.Reader) {
	r.reader = reader
}

// ReadLineUnbuffered reads a line from the reader one byte at a time until a newline character is encountered or an error occurs.
// It accepts an optional io.Reader argument to specify the reader to read from. If no reader is provided, it uses the LineReader's underlying reader.
// It returns the line without the line ending and any error that occurred during reading.
// This method should be used when the caller expects to read only a short single line
// or if the reader should not be consumed more than the necessary amount.
// Before calling ReadLineUnbuffered the first time, the caller should call Reset to clear any previous state.
// If several lines are to be read, the caller should subsequently call ReadLineUnbuffered repeatedly without calling Reset,
// until all lines have been read.
// ReadLineUnbuffered returns io.ErrUnexpectedEOF if the reader returns an EOF before a newline character is encountered.
// If the reader does not send an EOF, the caller must check the read lines for any end-of-stream conditions.
func (r *LineReader) ReadLineUnbuffered(readers ...io.Reader) ([]byte, error) {
	reader := r.getReader(readers...)
	// Clear the line buffer for the new line.
	r.lineBuffer = r.lineBuffer[:0]
	for {
		// Read one byte at a time until we encounter a newline character or an error occurs.
		r.setTimeout()
		r.bytes, r.err = reader.Read(r.singleByte)
		// If the reader returns a newline character, we return the current line buffer.
		if r.singleByte[0] == lf {
			// We trim any trailing CR characters from the line and return it.
			r.lineBuffer = bytes.TrimSuffix(r.lineBuffer, []byte{cr})
			r.lines++
			return r.lineBuffer, r.err
		}
		// If an error occurs during reading, we return it along with any line that was read so far.
		if r.err != nil {
			if r.bytes == 1 {
				r.lineBuffer = append(r.lineBuffer, r.singleByte[0])
			}
			if r.err == io.EOF && !(r.lines > 0 && len(r.lineBuffer) == 0) {
				// If the last character read was not a newline, we return an unexpected EOF error
				// unless we have read at least one line and the line buffer is empty, in which case we return the EOF error as is.
				r.err = io.ErrUnexpectedEOF
			}
			if len(r.lineBuffer) == 0 {
				r.lineBuffer = nil
			}
			return r.lineBuffer, r.err
		}
		r.lineBuffer = append(r.lineBuffer, r.singleByte[0])
	}
}

// ReadLineBuffered reads a line from the reader using a buffer to minimize the number of read operations.
// It accepts an optional io.Reader argument to specify the reader to read from. If no reader is provided, it uses the LineReader's underlying reader.
// It returns the line without the line ending and any error that occurred during reading.
// This method should be used when the caller expects to read multiple lines, as it is more efficient than ReadLineUnbuffered.
// Before calling ReadLineBuffered the first time, the caller should call Reset to clear any previous state.
// If several lines are to be read, the caller should subsequently call ReadLineBuffered repeatedly without calling Reset,
// until all lines have been read.
// ReadLineBuffered returns io.ErrUnexpectedEOF if the reader returns an EOF before a newline character is encountered.
// If the reader does not send an EOF, the caller must check the read lines for any end-of-stream conditions.
func (r *LineReader) ReadLineBuffered(readers ...io.Reader) ([]byte, error) {
	reader := r.getReader(readers...)
	// If there was a previous read error, we return it.
	if r.err != nil {
		if (r.lines == 0 && len(r.lineBuffer) == 0) || (len(r.lineBuffer) > 0 && r.lineBuffer[len(r.lineBuffer)-1] != lf) {
			// If the last character read was not a LF, we return an unexpected EOF error.
			r.err = io.ErrUnexpectedEOF
		}
		if len(r.lineBuffer) == 0 {
			r.lineBuffer = nil
		}
		return r.lineBuffer, r.err
	}
	// Clear the line buffer for the new line.
	r.lineBuffer = r.lineBuffer[:0]
	for {
		// If the read buffer is exhausted, read more data into it.
		if r.offset >= r.bytes {
			// If there was a previous read error, we return it.
			if r.err != nil {
				if (r.lines == 0 && len(r.lineBuffer) == 0) || (len(r.lineBuffer) > 0 && r.lineBuffer[len(r.lineBuffer)-1] != lf) {
					// If the last character read was not a LF, we return an unexpected EOF error.
					r.err = io.ErrUnexpectedEOF
				}
				if len(r.lineBuffer) == 0 {
					r.lineBuffer = nil
				}
				return r.lineBuffer, r.err
			}
			// We read more data into the read buffer.
			r.setTimeout()
			r.bytes, r.err = reader.Read(r.buffer)
			// Reset the read buffer offset for the new data.
			r.offset = 0
		}
		// We set chunk to the unread portion of the read buffer.
		chunk := r.buffer[r.offset:r.bytes]
		// We look for a newline character in the chunk.
		if i := bytes.IndexByte(chunk, '\n'); i >= 0 {
			// If we find a newline, we append the line up to and including the newline to the line buffer.
			r.lineBuffer = append(r.lineBuffer, chunk[:i+1]...)
			// We advance the read buffer offset past the line we just read.
			r.offset += i + 1
			// We trim any trailing LF or CR characters from the line and return it.
			r.lineBuffer = bytes.TrimSuffix(r.lineBuffer, []byte{lf})
			r.lineBuffer = bytes.TrimSuffix(r.lineBuffer, []byte{cr})
			r.lines++
			return r.lineBuffer, r.err
		}
		// If we don't find a newline, we append the entire chunk to the line buffer.
		r.lineBuffer = append(r.lineBuffer, chunk...)
		// We advance the read buffer offset to the end of the chunk since we've consumed it all.
		r.offset = r.bytes
	}
}

// Reset clears the internal state of the LineReader, allowing it to be reused for reading new lines from the connection.
func (r *LineReader) Reset() {
	r.bytes = 0
	r.offset = 0
	r.lines = 0
	r.singleByte[0] = 0
	r.lineBuffer = r.lineBuffer[:0]
	r.err = nil
}

// GetRemainingData returns any remaining data in the read buffer.
// If there is remaining data, it returns the remaining data and a nil error.
// If there is no remaining data, it returns nil and an error.
func (r *LineReader) GetRemainingData() (data []byte, err error) {
	if r.offset >= r.bytes {
		return nil, errors.New("buffer is empty")
	}
	remaining := r.buffer[r.offset:r.bytes]
	data = append(data, remaining...)
	r.offset = r.bytes
	return data, nil
}

func (r *LineReader) setTimeout() {
	if r.timeout <= 0 {
		return
	}
	if timoutReader, ok := r.reader.(interface{ SetReadDeadline(time.Time) error }); ok {
		timoutReader.SetReadDeadline(time.Now().Add(r.timeout))
	}
}

func (r *LineReader) getReader(readers ...io.Reader) io.Reader {
	if len(readers) == 0 {
		return r.reader
	}
	return readers[0]
}
