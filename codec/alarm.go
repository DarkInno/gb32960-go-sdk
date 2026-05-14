package codec

import (
	"encoding/binary"
	"fmt"
)

type AlarmInfo struct {
	MaxLevel     byte
	AlarmByteLen uint32
	AlarmBytes   []byte
}

func DecodeAlarmData(data []byte) (interface{}, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("alarm data too short: %d < 5", len(data))
	}
	a := &AlarmInfo{}
	pos := 0
	a.MaxLevel = data[pos]
	pos++
	a.AlarmByteLen = binary.BigEndian.Uint32(data[pos : pos+4])
	pos += 4
	if int(a.AlarmByteLen) > len(data)-pos {
		a.AlarmByteLen = uint32(len(data) - pos)
	}
	a.AlarmBytes = make([]byte, a.AlarmByteLen)
	copy(a.AlarmBytes, data[pos:pos+int(a.AlarmByteLen)])
	return a, nil
}
