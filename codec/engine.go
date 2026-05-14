package codec

import (
	"encoding/binary"
	"fmt"
)

type EngineInfo struct {
	EngineStatus byte
	CrankSpeed   uint16
	FuelRate     uint16
}

func DecodeEngineData(data []byte) (interface{}, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("engine data too short: %d < 5", len(data))
	}
	e := &EngineInfo{}
	pos := 0
	e.EngineStatus = data[pos]
	pos++
	e.CrankSpeed = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	e.FuelRate = binary.BigEndian.Uint16(data[pos : pos+2])
	return e, nil
}
