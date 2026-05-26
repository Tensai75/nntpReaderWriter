package nntpReaderWriter

import (
	"io"
)

type dotLineReader struct {
	rw       *NntpReaderWriter
	done     func()
	pending  []byte
	err      error
	closed   bool
	finished bool
}

func (dlr *dotLineReader) finish() {
	if dlr.finished {
		return
	}
	dlr.finished = true
	if dlr.done != nil {
		dlr.done()
	}
}

func (dlr *dotLineReader) Read(p []byte) (n int, err error) {
	if dlr.closed {
		return 0, io.ErrClosedPipe
	}
	if len(dlr.pending) == 0 && dlr.err != nil {
		return 0, dlr.err
	}
	for n < len(p) {
		if len(dlr.pending) == 0 {
			dlr.err = dlr.rw.readDotLineIntoLineBuffer()
			if dlr.err != nil {
				if dlr.err == io.EOF {
					dlr.finish()
					if n > 0 {
						return n, nil
					}
					return 0, io.EOF
				}
				dlr.finish()
				if n > 0 {
					return n, nil
				}
				return 0, dlr.err
			}
			dlr.pending = append(dlr.rw.lineBuffer, '\n')
		}
		copied := copy(p[n:], dlr.pending)
		n += copied
		dlr.pending = dlr.pending[copied:]
		if n == len(p) {
			return n, nil
		}
	}
	return n, nil
}

func (dlr *dotLineReader) Close() error {
	if dlr.closed {
		return nil
	}
	dlr.closed = true

	if dlr.finished {
		return nil
	}
	for {
		err := dlr.rw.readDotLineIntoLineBuffer()
		if err == io.EOF {
			dlr.finish()
			return nil
		}
		if err != nil {
			dlr.finish()
			return err
		}
	}
}
