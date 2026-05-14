package codec

const (
	FieldVehicle     = 0x01
	FieldMotor       = 0x02
	FieldFuelCell    = 0x03
	FieldEngine      = 0x04
	FieldPosition    = 0x05
	FieldExtreme     = 0x06
	FieldAlarm       = 0x07
	FieldVoltage     = 0x08
	FieldTemperature = 0x09
)

type FieldDecoder func(data []byte) (interface{}, error)

var decoders = map[byte]FieldDecoder{
	FieldVehicle:     DecodeVehicleData,
	FieldMotor:       DecodeMotorData,
	FieldFuelCell:    DecodeFuelCellData,
	FieldEngine:      DecodeEngineData,
	FieldPosition:    DecodePositionData,
	FieldExtreme:     DecodeExtremeData,
	FieldAlarm:       DecodeAlarmData,
	FieldVoltage:     DecodeVoltageData,
	FieldTemperature: DecodeTemperatureData,
}

func DecodeField(fieldID byte, data []byte) (interface{}, error) {
	if d, ok := decoders[fieldID]; ok {
		return d(data)
	}
	return data, nil
}

func HasDecoder(fieldID byte) bool {
	_, ok := decoders[fieldID]
	return ok
}
