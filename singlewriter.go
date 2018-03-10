package singlewriter

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

type SingleWriter struct {
	b         bytes.Buffer
	hasWriter bool
	closed    bool
	readers   []*reader
	sync.Mutex
}

type reader struct {
	buf    *SingleWriter
	index  int
	notify chan int
}

func NewSingleWriter() *SingleWriter {
	w := &SingleWriter{}
	w.readers = make([]*reader, 0)
	return w
}

func (w *SingleWriter) notify() {
	for _, r := range w.readers {
		select {
		case r.notify <- 1:
		default:
		}
	}
}

func (w *SingleWriter) Write(p []byte) (int, error) {
	w.Lock()
	defer func() {
		w.Unlock()
	}()
	if w.closed {
		return 0, errors.New("buffer closed")
	}
	n, err := w.b.Write(p)
	w.notify()
	return n, err
}

func (w *SingleWriter) ReadFrom(p []byte, off int) (int, error) {
	w.Lock()
	defer func() {
		w.Unlock()
	}()
	buf := w.b.Bytes()
	return copy(p, buf[off:]), nil
}

func (w *SingleWriter) Bytes() []byte {
	w.Lock()
	defer w.Unlock()
	return w.b.Bytes()
}

func (w *SingleWriter) Len() int {
	w.Lock()
	defer w.Unlock()
	return w.b.Len()
}

func (w *SingleWriter) Close() error {
	w.Lock()
	w.closed = true
	w.notify()
	w.Unlock()
	return nil
}

func (w *SingleWriter) closeReader(a *reader) {
	w.Lock()
	for i, r := range w.readers {
		if r == a {
			w.readers = w.readers[:i+copy(w.readers[i:], w.readers[i+1:])]
		}
	}
	w.Unlock()
}

func (w *SingleWriter) isClosed() bool {
	w.Lock()
	defer func() {
		w.Unlock()
	}()
	return w.closed
}

func (w *SingleWriter) Open() (io.ReadCloser, error) {
	w.Lock()
	defer w.Unlock()
	a := &reader{buf: w, notify: make(chan int)}
	w.readers = append(w.readers, a)
	return a, nil
}

func (a *reader) Read(p []byte) (int, error) {
again:
	n, err := a.buf.ReadFrom(p, a.index)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		if a.buf.isClosed() {
			return 0, io.EOF
		}
		<-a.notify
		goto again
	}
	a.index += n
	return n, nil
}

func (a *reader) Close() error {
	a.buf.closeReader(a)
	return nil
}
