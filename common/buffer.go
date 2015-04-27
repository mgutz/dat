package common

import (
	"bytes"
	"sync"
)

// BufferWriter a common interface between bytes.Buffer and bufio.Writer
type BufferWriter interface {
	Write(p []byte) (nn int, err error)
	WriteRune(r rune) (n int, err error)
	WriteString(s string) (n int, err error)
}

// BufferPool is sync pool for *bytes.Buffer
type BufferPool struct {
	sync.Pool
}

// NewBufferPool creates a buffer pool.
func NewBufferPool() *BufferPool {
	return &BufferPool{
		Pool: sync.Pool{New: func() interface{} {
			b := bytes.NewBuffer(make([]byte, 128))
			b.Reset()
			return b
		}},
	}
}

// Get checks out a buffer which must be put back in.
func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.Pool.Get().(*bytes.Buffer)
}

// Put reurns a buffer which was previously checked out.
func (bp *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	bp.Pool.Put(b)
}
