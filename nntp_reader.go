package nntpReaderWriter

import (
	"io"
	"time"
)

func (rw *NntpReaderWriter) readCodeResponseLine() (code int, msg string, err error) {
	rw.lineScanner.Reset()
	line, err := rw.readRawLine()
	if err != nil {
		return
	}
	if len(line) < 3 {
		err = ErrInvalidResponseLine
		return
	}
	b0, b1, b2 := line[0], line[1], line[2]
	if b0 < '0' || b0 > '9' || b1 < '0' || b1 > '9' || b2 < '0' || b2 > '9' {
		err = ErrInvalidResponseLine
		return
	}
	code = int(b0-'0')*100 + int(b1-'0')*10 + int(b2-'0')
	if len(line) > 4 {
		msg = string(line[4:])
	}
	return
}

func (rw *NntpReaderWriter) readDotLinesAsStrings() (lines []string, err error) {
	rw.lineScanner.Reset()
	var line []byte
	for {
		line, err = rw.readDotLine()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		lines = append(lines, string(line))
	}
}

func (rw *NntpReaderWriter) readDotLinesAsReader(done func()) (io.ReadCloser, error) {
	rw.lineScanner.Reset()
	return &dotLineReader{
		rw:   rw,
		done: done,
	}, nil
}

func (rw *NntpReaderWriter) readDotLinesAsBytesWithCallback(callback func([]byte) error) error {
	rw.lineScanner.Reset()
	var line []byte
	var err error
	for {
		line, err = rw.readDotLine()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if line != nil {
			if err := callback(line); err != nil {
				return err
			}
		}
	}
}

func (rw *NntpReaderWriter) readDotLinesAsStringsWithCallback(callback func(string) error) error {
	rw.lineScanner.Reset()
	var line []byte
	var err error
	for {
		line, err = rw.readDotLine()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		if line != nil {
			if err := callback(string(line)); err != nil {
				return err
			}
		}
	}
}

func (rw *NntpReaderWriter) readDotLine() (line []byte, err error) {
	line, err = rw.readRawLine()
	if err != nil {
		return
	}
	if len(line) == 1 && line[0] == '.' {
		return nil, io.EOF
	}
	if len(line) >= 2 && line[0] == '.' && line[1] == '.' {
		line = line[1:]
	}
	return
}

func (rw *NntpReaderWriter) readRawLine() ([]byte, error) {
	return rw.lineScanner.ScanLine(rw.readBytes())
}

func (rw *NntpReaderWriter) readBytes() io.Reader {
	if rw.readTimeout > 0 {
		rw.conn.SetReadDeadline(time.Now().Add(rw.readTimeout))
	}
	return rw.conn
}
