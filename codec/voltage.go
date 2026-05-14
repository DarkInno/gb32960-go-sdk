package codec

import (
	"encoding/binary"
	"fmt"
)

type VoltageInfo struct {
	SubsysCount        uint16
	SubsystemVoltages  []SubsystemVoltage
}

type SubsystemVoltage struct {
	SubsysNo  byte
	Voltage   uint16
	Current   uint16
	CellCount uint16
	Cells     []CellVoltage
}

type CellVoltage struct {
	CellNo   byte
	CellInfo CellInfo
}

type CellInfo struct {
	Voltage uint16
	Temp    byte
}

func DecodeVoltageData(data []byte) (interface{}, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("voltage data too short: %d < 2", len(data))
	}
	v := &VoltageInfo{}
	pos := 0
	v.SubsysCount = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2

	for i := uint16(0); i < v.SubsysCount; i++ {
		if pos+7 > len(data) {
			break
		}
		sub := SubsystemVoltage{}
		sub.SubsysNo = data[pos]
		pos++
		sub.Voltage = binary.BigEndian.Uint16(data[pos : pos+2])
		pos += 2
		sub.Current = binary.BigEndian.Uint16(data[pos : pos+2])
		pos += 2
		sub.CellCount = binary.BigEndian.Uint16(data[pos : pos+2])
		pos += 2

		for j := uint16(0); j < sub.CellCount; j++ {
			if pos+3 > len(data) {
				break
			}
			cell := CellVoltage{}
			cell.CellNo = data[pos]
			pos++
			cell.CellInfo.Voltage = binary.BigEndian.Uint16(data[pos : pos+2])
			pos += 2
			sub.Cells = append(sub.Cells, cell)
		}

		v.SubsystemVoltages = append(v.SubsystemVoltages, sub)
	}
	return v, nil
}
