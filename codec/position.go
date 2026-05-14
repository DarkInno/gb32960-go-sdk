package codec

import (
	"encoding/binary"
	"fmt"
)

type PositionInfo struct {
	Longitude uint32
	Latitude  uint32
}

func DecodePositionData(data []byte) (interface{}, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("position data too short: %d < 8", len(data))
	}
	p := &PositionInfo{}
	p.Longitude = binary.BigEndian.Uint32(data[0:4])
	p.Latitude = binary.BigEndian.Uint32(data[4:8])
	return p, nil
}
