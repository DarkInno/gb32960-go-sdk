package codec

import (
	"encoding/binary"
	"fmt"
)

type FuelCellInfo struct {
	CellVoltage              uint16
	CellCurrent              uint16
	FuelConsumption          uint16
	ProbeCount               uint16
	ProbeTemps               []byte
	H2MaxTemp                uint16
	H2MaxTempProbe           byte
	H2MaxConcentration       uint16
	H2MaxConcentrationSensor byte
	H2PressureMax            uint16
	H2PressureMaxSensor      byte
	H2PressureMin            uint16
	H2PressureMinSensor      byte
	DCDCStatus               byte
}

func DecodeFuelCellData(data []byte) (interface{}, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("fuel cell data too short: %d < 8", len(data))
	}
	f := &FuelCellInfo{}
	pos := 0
	f.CellVoltage = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	f.CellCurrent = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	f.FuelConsumption = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	f.ProbeCount = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2

	probeCnt := int(f.ProbeCount)
	expectedLen := pos + probeCnt + 13
	if len(data) < expectedLen {
		return nil, fmt.Errorf("fuel cell data truncated: need %d bytes, have %d", expectedLen, len(data))
	}
	f.ProbeTemps = make([]byte, probeCnt)
	copy(f.ProbeTemps, data[pos:pos+probeCnt])
	pos += probeCnt

	f.H2MaxTemp = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	f.H2MaxTempProbe = data[pos]
	pos++
	f.H2MaxConcentration = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	f.H2MaxConcentrationSensor = data[pos]
	pos++
	f.H2PressureMax = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	f.H2PressureMaxSensor = data[pos]
	pos++
	f.H2PressureMin = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	f.H2PressureMinSensor = data[pos]
	pos++
	f.DCDCStatus = data[pos]
	return f, nil
}
