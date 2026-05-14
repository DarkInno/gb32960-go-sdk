package codec

import (
	"encoding/binary"
	"fmt"
)

type ExtremeInfo struct {
	MaxBatteryVoltage     uint16
	MaxBatteryVoltageCode byte
	MinBatteryVoltage     uint16
	MinBatteryVoltageCode byte
	MaxTemp               byte
	MaxTempCode           byte
	MinTemp               byte
	MinTempCode           byte
}

func DecodeExtremeData(data []byte) (interface{}, error) {
	if len(data) < 10 {
		return nil, fmt.Errorf("extreme data too short: %d < 10", len(data))
	}
	e := &ExtremeInfo{}
	pos := 0
	e.MaxBatteryVoltage = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	e.MaxBatteryVoltageCode = data[pos]
	pos++
	e.MinBatteryVoltage = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	e.MinBatteryVoltageCode = data[pos]
	pos++
	e.MaxTemp = data[pos]
	pos++
	e.MaxTempCode = data[pos]
	pos++
	e.MinTemp = data[pos]
	pos++
	e.MinTempCode = data[pos]
	return e, nil
}
