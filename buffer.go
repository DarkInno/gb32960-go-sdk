package gb32960

import "sync"

const readBufSize = 4096

var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, readBufSize)
		return &buf
	},
}

func getBuffer() *[]byte {
	return bufferPool.Get().(*[]byte)
}

func putBuffer(buf *[]byte) {
	bufferPool.Put(buf)
}
