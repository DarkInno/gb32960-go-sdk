package gb32960

import (
	"encoding/binary"
	"fmt"

	"github.com/darkinno/gb32960-go-sdk/codec"
)

func DecodeRealtimeData(data []byte) (*RealtimeMessage, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("realtime data too short: %d bytes", len(data))
	}

	msg := &RealtimeMessage{
		InfoTime: parseTime6(data[0:6]),
	}
	pos := 6

	for pos < len(data) {
		if pos+1 > len(data) {
			break
		}
		fieldID := data[pos]
		if !codec.HasDecoder(fieldID) {
			break
		}
		pos++

		fieldEnd := findFieldEnd(fieldID, data, pos)
		if fieldEnd < 0 || fieldEnd > len(data) {
			break
		}

		fieldData := data[pos:fieldEnd]
		pos = fieldEnd

		decoded, err := codec.DecodeField(fieldID, fieldData)
		if err != nil {
			continue
		}

		switch fieldID {
		case codec.FieldVehicle:
			if v, ok := decoded.(*codec.VehicleBaseInfo); ok {
				msg.VehicleData = &VehicleBaseInfo{
					VehicleStatus:  v.VehicleStatus,
					ChargingStatus: v.ChargingStatus,
					RunMode:        v.RunMode,
					Speed:          v.Speed,
					Odometer:       v.Odometer,
					TotalVoltage:   v.TotalVoltage,
					TotalCurrent:   v.TotalCurrent,
					SOC:            v.SOC,
					DCStatus:       v.DCStatus,
					Gear:           v.Gear,
					InsulationRes:  v.InsulationRes,
				}
			}
		case codec.FieldMotor:
			if motors, ok := decoded.([]codec.MotorInfo); ok {
				for _, m := range motors {
					msg.MotorData = append(msg.MotorData, MotorInfo{
						MotorSeq:       m.MotorSeq,
						MotorStatus:    m.MotorStatus,
						ControllerTemp: m.ControllerTemp,
						MotorSpeed:     m.MotorSpeed,
						MotorTorque:    m.MotorTorque,
						MotorTemp:      m.MotorTemp,
						MotorVoltage:   m.MotorVoltage,
						MotorCurrent:   m.MotorCurrent,
					})
				}
			}
		case codec.FieldFuelCell:
			if fc, ok := decoded.(*codec.FuelCellInfo); ok {
				msg.FuelCellData = &FuelCellInfo{
					CellVoltage:              fc.CellVoltage,
					CellCurrent:              fc.CellCurrent,
					FuelConsumption:          fc.FuelConsumption,
					ProbeCount:               fc.ProbeCount,
					ProbeTemps:               fc.ProbeTemps,
					H2MaxTemp:                fc.H2MaxTemp,
					H2MaxTempProbe:           fc.H2MaxTempProbe,
					H2MaxConcentration:       fc.H2MaxConcentration,
					H2MaxConcentrationSensor: fc.H2MaxConcentrationSensor,
					H2PressureMax:            fc.H2PressureMax,
					H2PressureMaxSensor:      fc.H2PressureMaxSensor,
					H2PressureMin:            fc.H2PressureMin,
					H2PressureMinSensor:      fc.H2PressureMinSensor,
					DCDCStatus:               fc.DCDCStatus,
				}
			}
		case codec.FieldEngine:
			if e, ok := decoded.(*codec.EngineInfo); ok {
				msg.EngineData = &EngineInfo{
					EngineStatus: e.EngineStatus,
					CrankSpeed:   e.CrankSpeed,
					FuelRate:     e.FuelRate,
				}
			}
		case codec.FieldPosition:
			if p, ok := decoded.(*codec.PositionInfo); ok {
				msg.PositionData = &PositionInfo{
					Longitude: p.Longitude,
					Latitude:  p.Latitude,
				}
			}
		case codec.FieldExtreme:
			if e, ok := decoded.(*codec.ExtremeInfo); ok {
				msg.ExtremeData = &ExtremeInfo{
					MaxBatteryVoltage:      e.MaxBatteryVoltage,
					MaxBatteryVoltageCode:  e.MaxBatteryVoltageCode,
					MinBatteryVoltage:      e.MinBatteryVoltage,
					MinBatteryVoltageCode:  e.MinBatteryVoltageCode,
					MaxTemp:                e.MaxTemp,
					MaxTempCode:            e.MaxTempCode,
					MinTemp:                e.MinTemp,
					MinTempCode:            e.MinTempCode,
				}
			}
		case codec.FieldAlarm:
			if a, ok := decoded.(*codec.AlarmInfo); ok {
				msg.AlarmData = &AlarmInfo{
					MaxLevel:     a.MaxLevel,
					AlarmByteLen: a.AlarmByteLen,
					AlarmBytes:   a.AlarmBytes,
				}
			}
		case codec.FieldVoltage:
			if v, ok := decoded.(*codec.VoltageInfo); ok {
				vi := &VoltageInfo{
					SubsysCount: v.SubsysCount,
				}
				for _, sub := range v.SubsystemVoltages {
					sv := SubsystemVoltage{
						SubsysNo:  sub.SubsysNo,
						Voltage:   sub.Voltage,
						Current:   sub.Current,
						CellCount: sub.CellCount,
					}
					for _, c := range sub.Cells {
						sv.Cells = append(sv.Cells, CellVoltage{
							CellNo: c.CellNo,
							CellInfo: CellInfo{
								Voltage: c.CellInfo.Voltage,
								Temp:    c.CellInfo.Temp,
							},
						})
					}
					vi.SubsystemVoltages = append(vi.SubsystemVoltages, sv)
				}
				msg.VoltageData = vi
			}
		}
	}

	return msg, nil
}

func findFieldEnd(fieldID byte, data []byte, pos int) int {
	switch fieldID {
	case codec.FieldVehicle:
		if pos+20 <= len(data) {
			return pos + 20
		}
	case codec.FieldMotor:
		if pos+1 <= len(data) {
			count := int(data[pos])
			return pos + 1 + count*12
		}
	case codec.FieldFuelCell:
		if pos+8 <= len(data) {
			probeCnt := int(binary.BigEndian.Uint16(data[pos+6 : pos+8]))
			return pos + 8 + probeCnt + 13
		}
	case codec.FieldEngine:
		if pos+5 <= len(data) {
			return pos + 5
		}
	case codec.FieldPosition:
		if pos+8 <= len(data) {
			return pos + 8
		}
	case codec.FieldExtreme:
		if pos+10 <= len(data) {
			return pos + 10
		}
	case codec.FieldAlarm:
		if pos+5 <= len(data) {
			alarmLen := int(binary.BigEndian.Uint32(data[1:5]))
			return pos + 5 + alarmLen
		}
	case codec.FieldVoltage:
		if pos+2 <= len(data) {
			subCnt := int(binary.BigEndian.Uint16(data[pos : pos+2]))
			end := pos + 2
			for i := 0; i < subCnt && end < len(data); i++ {
				if end+7 > len(data) {
					break
				}
				end += 7
				cellCnt := int(binary.BigEndian.Uint16(data[end-2 : end]))
				end += cellCnt * 3
			}
			return end
		}
	case codec.FieldTemperature:
		if pos+2 <= len(data) {
			subCnt := int(binary.BigEndian.Uint16(data[pos : pos+2]))
			end := pos + 2
			for i := 0; i < subCnt && end < len(data); i++ {
				if end+3 > len(data) {
					break
				}
				end += 3
				probeCnt := int(binary.BigEndian.Uint16(data[end-2 : end]))
				end += probeCnt
			}
			return end
		}
	}
	return -1
}

func DecodeReissueData(data []byte) (*ReissueMessage, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("reissue data too short: %d bytes", len(data))
	}

	pos := 0
	infoTime := parseTime6(data[pos : pos+6])
	pos += 6

	totalItems := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2

	msg := &ReissueMessage{
		InfoTime:  infoTime,
		DataItems: make([]RealtimeMessage, 0, totalItems),
	}

	for i := 0; i < totalItems && pos < len(data); i++ {
		item, consumed, err := decodeRealtimeItem(data[pos:])
		if err != nil {
			break
		}
		if item != nil {
			msg.DataItems = append(msg.DataItems, *item)
		}
		pos += consumed
	}

	return msg, nil
}

func decodeRealtimeItem(data []byte) (*RealtimeMessage, int, error) {
	if len(data) < 6 {
		return nil, 0, fmt.Errorf("realtime item too short: %d bytes", len(data))
	}

	msg := &RealtimeMessage{
		InfoTime: parseTime6(data[0:6]),
	}
	pos := 6

	for pos < len(data) {
		if pos+1 > len(data) {
			break
		}
		fieldID := data[pos]
		if !codec.HasDecoder(fieldID) {
			break
		}
		pos++

		fieldEnd := findFieldEnd(fieldID, data, pos)
		if fieldEnd < 0 || fieldEnd > len(data) {
			break
		}

		fieldData := data[pos:fieldEnd]
		pos = fieldEnd

		decoded, err := codec.DecodeField(fieldID, fieldData)
		if err != nil {
			continue
		}

		switch fieldID {
		case codec.FieldVehicle:
			if v, ok := decoded.(*codec.VehicleBaseInfo); ok {
				msg.VehicleData = &VehicleBaseInfo{
					VehicleStatus:  v.VehicleStatus,
					ChargingStatus: v.ChargingStatus,
					RunMode:        v.RunMode,
					Speed:          v.Speed,
					Odometer:       v.Odometer,
					TotalVoltage:   v.TotalVoltage,
					TotalCurrent:   v.TotalCurrent,
					SOC:            v.SOC,
					DCStatus:       v.DCStatus,
					Gear:           v.Gear,
					InsulationRes:  v.InsulationRes,
				}
			}
		case codec.FieldMotor:
			if motors, ok := decoded.([]codec.MotorInfo); ok {
				for _, m := range motors {
					msg.MotorData = append(msg.MotorData, MotorInfo{
						MotorSeq:       m.MotorSeq,
						MotorStatus:    m.MotorStatus,
						ControllerTemp: m.ControllerTemp,
						MotorSpeed:     m.MotorSpeed,
						MotorTorque:    m.MotorTorque,
						MotorTemp:      m.MotorTemp,
						MotorVoltage:   m.MotorVoltage,
						MotorCurrent:   m.MotorCurrent,
					})
				}
			}
		case codec.FieldFuelCell:
			if fc, ok := decoded.(*codec.FuelCellInfo); ok {
				msg.FuelCellData = &FuelCellInfo{
					CellVoltage:              fc.CellVoltage,
					CellCurrent:              fc.CellCurrent,
					FuelConsumption:          fc.FuelConsumption,
					ProbeCount:               fc.ProbeCount,
					ProbeTemps:               fc.ProbeTemps,
					H2MaxTemp:                fc.H2MaxTemp,
					H2MaxTempProbe:           fc.H2MaxTempProbe,
					H2MaxConcentration:       fc.H2MaxConcentration,
					H2MaxConcentrationSensor: fc.H2MaxConcentrationSensor,
					H2PressureMax:            fc.H2PressureMax,
					H2PressureMaxSensor:      fc.H2PressureMaxSensor,
					H2PressureMin:            fc.H2PressureMin,
					H2PressureMinSensor:      fc.H2PressureMinSensor,
					DCDCStatus:               fc.DCDCStatus,
				}
			}
		case codec.FieldEngine:
			if e, ok := decoded.(*codec.EngineInfo); ok {
				msg.EngineData = &EngineInfo{
					EngineStatus: e.EngineStatus,
					CrankSpeed:   e.CrankSpeed,
					FuelRate:     e.FuelRate,
				}
			}
		case codec.FieldPosition:
			if p, ok := decoded.(*codec.PositionInfo); ok {
				msg.PositionData = &PositionInfo{
					Longitude: p.Longitude,
					Latitude:  p.Latitude,
				}
			}
		case codec.FieldExtreme:
			if e, ok := decoded.(*codec.ExtremeInfo); ok {
				msg.ExtremeData = &ExtremeInfo{
					MaxBatteryVoltage:      e.MaxBatteryVoltage,
					MaxBatteryVoltageCode:  e.MaxBatteryVoltageCode,
					MinBatteryVoltage:      e.MinBatteryVoltage,
					MinBatteryVoltageCode:  e.MinBatteryVoltageCode,
					MaxTemp:                e.MaxTemp,
					MaxTempCode:            e.MaxTempCode,
					MinTemp:                e.MinTemp,
					MinTempCode:            e.MinTempCode,
				}
			}
		case codec.FieldAlarm:
			if a, ok := decoded.(*codec.AlarmInfo); ok {
				msg.AlarmData = &AlarmInfo{
					MaxLevel:     a.MaxLevel,
					AlarmByteLen: a.AlarmByteLen,
					AlarmBytes:   a.AlarmBytes,
				}
			}
		case codec.FieldVoltage:
			if v, ok := decoded.(*codec.VoltageInfo); ok {
				vi := &VoltageInfo{
					SubsysCount: v.SubsysCount,
				}
				for _, sub := range v.SubsystemVoltages {
					sv := SubsystemVoltage{
						SubsysNo:  sub.SubsysNo,
						Voltage:   sub.Voltage,
						Current:   sub.Current,
						CellCount: sub.CellCount,
					}
					for _, c := range sub.Cells {
						sv.Cells = append(sv.Cells, CellVoltage{
							CellNo: c.CellNo,
							CellInfo: CellInfo{
								Voltage: c.CellInfo.Voltage,
								Temp:    c.CellInfo.Temp,
							},
						})
					}
					vi.SubsystemVoltages = append(vi.SubsystemVoltages, sv)
				}
				msg.VoltageData = vi
			}
		}
	}

	return msg, pos, nil
}
