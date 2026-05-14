package codec

import (
	"encoding/binary"
	"fmt"
)

type MotorInfo struct {
	MotorSeq       byte
	MotorStatus    byte
	ControllerTemp byte
	MotorSpeed     uint16
	MotorTorque    uint16
	MotorTemp      byte
	MotorVoltage   uint16
	MotorCurrent   uint16
}

func DecodeMotorData(data []byte) (interface{}, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("motor data too short")
	}
	pos := 0
	count := int(data[pos])
	pos++

	motors := make([]MotorInfo, count)
	for i := 0; i < count; i++ {
		if pos+12 > len(data) {
			break
		}
		m := MotorInfo{}
		m.MotorSeq = data[pos]
		pos++
		m.MotorStatus = data[pos]
		pos++
		m.ControllerTemp = data[pos]
		pos++
		m.MotorSpeed = binary.BigEndian.Uint16(data[pos : pos+2])
		pos += 2
		m.MotorTorque = binary.BigEndian.Uint16(data[pos : pos+2])
		pos += 2
		m.MotorTemp = data[pos]
		pos++
		m.MotorVoltage = binary.BigEndian.Uint16(data[pos : pos+2])
		pos += 2
		m.MotorCurrent = binary.BigEndian.Uint16(data[pos : pos+2])
		pos += 2
		motors[i] = m
	}
	return motors, nil
}
