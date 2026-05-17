package nntpReaderWriter

import (
	"bytes"
	"io"
	"strings"
	"time"
)

func (rw *NntpReaderWriter) writeDotLinesFromStrings(lines []string) error {
	for _, line := range lines {
		if err := rw.writeDotLineFromString(line); err != nil {
			return err
		}
	}
	return rw.writeLineFromString(".")
}

func (rw *NntpReaderWriter) writeDotLinesFromReader(r io.Reader) error {
	rw.lineScanner.Reset()
	for {
		line, err := rw.lineScanner.ScanLine(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := rw.writeDotLineFromBytes(line); err != nil {
			return err
		}
	}
	return rw.writeLineFromString(".")
}

func (rw *NntpReaderWriter) writeDotLinesFromChan(ch chan string, errChan chan error) error {
	var err error
	for line := range ch {
		err = rw.writeDotLineFromString(line)
		if err != nil {
			errChan <- err
		}
	}
	err = rw.writeLineFromString(".")
	if err != nil {
		errChan <- err
	}
	return err
}

func (rw *NntpReaderWriter) writeDotLineFromString(line string) error {
	if err := checkStringLineForInvalidChars(line); err != nil {
		return err
	}
	prefix := ""
	if len(line) > 0 && line[0] == '.' {
		prefix = "."
	}
	return rw.writeLineFromString(line, prefix)
}

func (rw *NntpReaderWriter) writeDotLineFromBytes(line []byte) error {
	if err := checkBytesLineForInvalidChars(line); err != nil {
		return err
	}
	prefix := ""
	if len(line) > 0 && line[0] == '.' {
		prefix = "."
	}
	return rw.writeLineFromBytes(line, prefix)
}

func (rw *NntpReaderWriter) writeLineFromBytes(line []byte, prefixes ...string) error {
	rw.wbuf = rw.wbuf[:0]
	for _, p := range prefixes {
		rw.wbuf = append(rw.wbuf, p...)
	}
	rw.wbuf = append(rw.wbuf, line...)
	rw.wbuf = append(rw.wbuf, '\r', '\n')
	_, err := rw.writeBytes(rw.wbuf)
	return err
}

func (rw *NntpReaderWriter) writeLineFromString(line string, prefixes ...string) error {
	rw.wbuf = rw.wbuf[:0]
	for _, p := range prefixes {
		rw.wbuf = append(rw.wbuf, p...)
	}
	rw.wbuf = append(rw.wbuf, line...)
	rw.wbuf = append(rw.wbuf, '\r', '\n')
	_, err := rw.writeBytes(rw.wbuf)
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
		return ErrInvalidWriteline("\\n")
	}
	if strings.Contains(line, "\r") {
		return ErrInvalidWriteline("\\r")
	}
	if strings.Contains(line, "\x00") {
		return ErrInvalidWriteline("\\x00")
	}
	return nil
}

func checkBytesLineForInvalidChars(line []byte) error {
	if bytes.Contains(line, []byte{'\n'}) {
		return ErrInvalidWriteline("\\n")
	}
	if bytes.Contains(line, []byte{'\r'}) {
		return ErrInvalidWriteline("\\r")
	}
	if bytes.Contains(line, []byte{'\x00'}) {
		return ErrInvalidWriteline("\\x00")
	}
	return nil
}
