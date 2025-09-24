package api

import (
	"sync"
	"unsafe"
)

// buffer size (64) is chosen to fit any float64 in 'g' format + JSON wrapper `{"result":<value>}`
type buffer [64]byte

type bufferPool struct {
	pool sync.Pool
}

func (bp *bufferPool) Get() []byte {
	p := bp.pool.Get().(*byte)
	return unsafe.Slice(p, unsafe.Sizeof(buffer{}))[:0]
}

func (p *bufferPool) Put(b []byte) {
	if uintptr(cap(b)) != unsafe.Sizeof(buffer{}) { // friend/foe check
		return
	}
	p.pool.Put(&b[0])
}

var bufPool = &bufferPool{
	pool: sync.Pool{
		New: func() any {
			var b buffer
			return &b[0]
		},
	},
}
