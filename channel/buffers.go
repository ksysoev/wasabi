package channel

import (
	"bytes"
	"sync"
)

type bufferPool struct {
	pool *sync.Pool
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (p *bufferPool) get() *bytes.Buffer {
	b, _ := p.pool.Get().(*bytes.Buffer)
	return b
}

func (p *bufferPool) put(b *bytes.Buffer) {
	b.Reset()
	p.pool.Put(b)
}
