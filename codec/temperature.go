package codec

import (
	"encoding/binary"
	"fmt"
)

type TemperatureInfo struct {
	SubsysCount           uint16
	SubsystemTemperatures []SubsystemTemperature
}

type SubsystemTemperature struct {
	SubsysNo       byte
	ProbeCount     uint16
	ProbeTemperatures []byte
}

func DecodeTemperatureData(data []byte) (interface{}, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("temperature data too short: %d < 2", len(data))
	}
	t := &TemperatureInfo{}
	pos := 0
	t.SubsysCount = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2

	for i := uint16(0); i < t.SubsysCount; i++ {
		if pos+3 > len(data) {
			break
		}
		sub := SubsystemTemperature{}
		sub.SubsysNo = data[pos]
		pos++
		sub.ProbeCount = binary.BigEndian.Uint16(data[pos : pos+2])
		pos += 2

		probeCnt := int(sub.ProbeCount)
		if pos+probeCnt > len(data) {
			probeCnt = len(data) - pos
		}
		sub.ProbeTemperatures = make([]byte, probeCnt)
		copy(sub.ProbeTemperatures, data[pos:pos+probeCnt])
		pos += probeCnt

		t.SubsystemTemperatures = append(t.SubsystemTemperatures, sub)
	}
	return t, nil
}
