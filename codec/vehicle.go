package codec

import (
	"encoding/binary"
	"fmt"
)

type VehicleBaseInfo struct {
	VehicleStatus  byte
	ChargingStatus byte
	RunMode        byte
	Speed          uint16
	Odometer       uint32
	TotalVoltage   uint16
	TotalCurrent   uint16
	SOC            byte
	DCStatus       byte
	Gear           byte
	InsulationRes  uint16
	Reserved       uint16
}

func DecodeVehicleData(data []byte) (interface{}, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("vehicle data too short: %d < 20", len(data))
	}
	v := &VehicleBaseInfo{}
	pos := 0
	v.VehicleStatus = data[pos]
	pos++
	v.ChargingStatus = data[pos]
	pos++
	v.RunMode = data[pos]
	pos++
	v.Speed = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	v.Odometer = binary.BigEndian.Uint32(data[pos : pos+4])
	pos += 4
	v.TotalVoltage = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	v.TotalCurrent = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	v.SOC = data[pos]
	pos++
	v.DCStatus = data[pos]
	pos++
	v.Gear = data[pos]
	pos++
	v.InsulationRes = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	v.Reserved = binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	return v, nil
}
