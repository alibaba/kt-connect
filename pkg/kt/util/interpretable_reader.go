package util

import "io"

type InterpretableReader struct {
	r         io.Reader
	interrupt chan int
}

func NewInterpretableReader(r io.Reader) InterpretableReader {
	return InterpretableReader{
		r,
		make(chan int),
	}
}

func (r InterpretableReader) Read(p []byte) (n int, err error) {
	if r.r == nil {
		return 0, io.EOF
	}
	select {
	case <-r.interrupt:
		r.r = nil
		return 0, io.EOF
	default:
		return r.r.Read(p)
	}
}

func (r InterpretableReader) Cancel() {
	r.interrupt <- 0
}
