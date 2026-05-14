package gb32960

import "sync"

const (
	readBufSize   = 4096
	packetBufSize = 65535 + 25
)

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

var packetBufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, packetBufSize)
		return &buf
	},
}

func getPacketBuffer() *[]byte {
	return packetBufferPool.Get().(*[]byte)
}

func putPacketBuffer(buf *[]byte) {
	packetBufferPool.Put(buf)
}
