package buffer

import (
	"sync"
	"unsafe"
)

// Size of buffer (64) is chosen to fit any float64 in 'g' format + JSON wrapper `{"result":<value>}`
const Size = 64

type buffer [Size]byte

type bufferPool struct {
	pool sync.Pool
}

func (bp *bufferPool) Get() []byte {
	return unsafe.Slice(bp.pool.Get().(*byte), Size)[:0]
}

func (p *bufferPool) Put(b []byte) {
	if uintptr(cap(b)) != Size { // friend/foe check
		return
	}
	p.pool.Put(unsafe.SliceData(b))
}

var bufPool = &bufferPool{
	pool: sync.Pool{
		New: func() any {
			var b buffer
			return &b[0]
		},
	},
}

func Get() []byte  { return bufPool.Get() }
func Put(b []byte) { bufPool.Put(b) }
