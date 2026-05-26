package nntpReaderWriter

import (
	"io"
)

// ReadCodeResponseLine reads a single-line response and parses the response code and message. It returns an error if the line is not a valid response line.
func (rw *NntpReaderWriter) ReadCodeResponseLine() (code int, msg string, err error) {
	rw.lineReader.Reset()
	rw.lineBuffer, err = rw.lineReader.ReadLineUnbuffered()
	if err != nil {
		return
	}
	if len(rw.lineBuffer) < 3 {
		err = errInvalidResponseLine()
		return
	}
	b0, b1, b2 := rw.lineBuffer[0], rw.lineBuffer[1], rw.lineBuffer[2]
	if b0 < '0' || b0 > '9' || b1 < '0' || b1 > '9' || b2 < '0' || b2 > '9' {
		err = errInvalidResponseLine()
		return
	}
	rw.code = int(b0-'0')*100 + int(b1-'0')*10 + int(b2-'0')
	if len(rw.lineBuffer) > 4 {
		rw.msg = string(rw.lineBuffer[4:])
	}
	return rw.code, rw.msg, nil
}

// ReadDotLines reads dot-encoded lines and returns them as a slice of strings.
func (rw *NntpReaderWriter) ReadDotLines() (lines []string, err error) {
	rw.lineReader.Reset()
	for {
		err = rw.readDotLineIntoLineBuffer()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		lines = append(lines, string(rw.lineBuffer))
	}
}

// ReadDotLinesReader returns an io.ReadCloser that can be used to read dot-encoded lines.
// The caller should provide the EndSequence function as the done callback or make sure to call it when finished reading.
func (rw *NntpReaderWriter) ReadDotLinesReader(done func()) (io.ReadCloser, error) {
	rw.lineReader.Reset()
	return &dotLineReader{
		rw:   rw,
		done: done,
	}, nil
}

// ReadDotLinesCallback reads dot-encoded lines and invokes the callback for each line as a string.
func (rw *NntpReaderWriter) ReadDotLinesCallback(callback func(string) error) (err error) {
	rw.lineReader.Reset()
	for {
		err = rw.readDotLineIntoLineBuffer()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if err := callback(string(rw.lineBuffer)); err != nil {
			return err
		}
	}
}

// ReadDotLinesCallbackBytes reads dot-encoded lines and invokes the callback for each line as bytes.
func (rw *NntpReaderWriter) ReadDotLinesCallbackBytes(callback func([]byte) error) (err error) {
	rw.lineReader.Reset()
	for {
		err = rw.readDotLineIntoLineBuffer()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if err := callback(rw.lineBuffer); err != nil {
			return err
		}
	}
}

// readDotLineIntoLineBuffer reads a single dot-encoded line into the line buffer.
// It returns io.EOF when the end of the dot-encoded block is reached.
func (rw *NntpReaderWriter) readDotLineIntoLineBuffer() (err error) {
	rw.lineBuffer = rw.lineBuffer[:0]
	rw.lineBuffer, err = rw.lineReader.ReadLineBuffered()
	if err != nil {
		return
	}
	if len(rw.lineBuffer) == 1 && rw.lineBuffer[0] == '.' {
		rw.lineBuffer = rw.lineBuffer[:0]
		return io.EOF
	}
	if len(rw.lineBuffer) >= 2 && rw.lineBuffer[0] == '.' && rw.lineBuffer[1] == '.' {
		rw.lineBuffer = rw.lineBuffer[1:]
	}
	return
}
