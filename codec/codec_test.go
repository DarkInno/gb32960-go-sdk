package codec

import (
	"encoding/binary"
	"testing"
)

func assertEqual(t *testing.T, expected, actual interface{}, msg ...string) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v %v", expected, actual, msg)
	}
}

func assertNotNil(t *testing.T, v interface{}, msg ...string) {
	t.Helper()
	if v == nil {
		t.Errorf("expected non-nil %v", msg)
	}
}

func assertNoError(t *testing.T, err error, msg ...string) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v %v", err, msg)
	}
}

func TestDecodeVehicleData(t *testing.T) {
	data := make([]byte, 20)
	data[0] = 0x01
	data[1] = 0x02
	data[2] = 0x01
	binary.BigEndian.PutUint16(data[3:5], 600)
	binary.BigEndian.PutUint32(data[5:9], 123456)
	binary.BigEndian.PutUint16(data[9:11], 400)
	binary.BigEndian.PutUint16(data[11:13], 500)
	data[13] = 85
	data[14] = 0x01
	data[15] = 0x03
	binary.BigEndian.PutUint16(data[16:18], 9999)
	binary.BigEndian.PutUint16(data[18:20], 0)

	result, err := DecodeVehicleData(data)
	assertNoError(t, err)
	assertNotNil(t, result)

	v, ok := result.(*VehicleBaseInfo)
	if !ok {
		t.Fatal("expected *VehicleBaseInfo")
	}
	assertEqual(t, byte(0x01), v.VehicleStatus)
	assertEqual(t, byte(0x02), v.ChargingStatus)
	assertEqual(t, uint16(600), v.Speed)
	assertEqual(t, uint32(123456), v.Odometer)
	assertEqual(t, byte(85), v.SOC)
	assertEqual(t, byte(0x03), v.Gear)
}

func TestDecodeMotorData(t *testing.T) {
	data := []byte{
		0x02,
		0x01, 0x02, 0x28,
		0x13, 0x88, 0x03, 0xE8,
		0x32, 0x01, 0x90, 0x00, 0x64,
		0x02, 0x01, 0x1E,
		0x0B, 0xB8, 0x02, 0x58,
		0x23, 0x01, 0x90, 0x00, 0x50,
	}

	result, err := DecodeMotorData(data)
	assertNoError(t, err)

	motors, ok := result.([]MotorInfo)
	if !ok {
		t.Fatal("expected []MotorInfo")
	}
	assertEqual(t, 2, len(motors))
	assertEqual(t, byte(0x01), motors[0].MotorSeq)
	assertEqual(t, uint16(5000), motors[0].MotorSpeed)
	assertEqual(t, byte(0x02), motors[1].MotorSeq)
}

func TestDecodePositionData(t *testing.T) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], 121473708)
	binary.BigEndian.PutUint32(data[4:8], 31235538)

	result, err := DecodePositionData(data)
	assertNoError(t, err)

	p, ok := result.(*PositionInfo)
	if !ok {
		t.Fatal("expected *PositionInfo")
	}
	assertEqual(t, uint32(121473708), p.Longitude)
	assertEqual(t, uint32(31235538), p.Latitude)
}

func TestDecodeExtremeData(t *testing.T) {
	data := make([]byte, 10)
	binary.BigEndian.PutUint16(data[0:2], 4200)
	data[2] = 0x05
	binary.BigEndian.PutUint16(data[3:5], 3800)
	data[5] = 0x10
	data[6] = 45
	data[7] = 0x03
	data[8] = 15
	data[9] = 0x1F

	result, err := DecodeExtremeData(data)
	assertNoError(t, err)

	e, ok := result.(*ExtremeInfo)
	if !ok {
		t.Fatal("expected *ExtremeInfo")
	}
	assertEqual(t, uint16(4200), e.MaxBatteryVoltage)
	assertEqual(t, uint16(3800), e.MinBatteryVoltage)
	assertEqual(t, byte(45), e.MaxTemp)
	assertEqual(t, byte(15), e.MinTemp)
}

func TestDecodeEngineData(t *testing.T) {
	data := make([]byte, 5)
	data[0] = 0x01
	binary.BigEndian.PutUint16(data[1:3], 2500)
	binary.BigEndian.PutUint16(data[3:5], 150)

	result, err := DecodeEngineData(data)
	assertNoError(t, err)

	e, ok := result.(*EngineInfo)
	if !ok {
		t.Fatal("expected *EngineInfo")
	}
	assertEqual(t, uint16(2500), e.CrankSpeed)
	assertEqual(t, uint16(150), e.FuelRate)
}

func TestDecodeAlarmData(t *testing.T) {
	data := []byte{
		0x01,
		0x00, 0x00, 0x00, 0x03,
		0xAA, 0xBB, 0xCC,
	}

	result, err := DecodeAlarmData(data)
	assertNoError(t, err)

	a, ok := result.(*AlarmInfo)
	if !ok {
		t.Fatal("expected *AlarmInfo")
	}
	assertEqual(t, byte(1), a.MaxLevel)
	assertEqual(t, uint32(3), a.AlarmByteLen)
	assertEqual(t, 3, len(a.AlarmBytes))
}

func TestDecodeVoltageData(t *testing.T) {
	data := []byte{
		0x00, 0x01,
		0x02,
		0x01, 0x90,
		0x00, 0x64,
		0x00, 0x02,
		0x01, 0x00, 0x64,
		0x02, 0x00, 0x65,
	}

	result, err := DecodeVoltageData(data)
	assertNoError(t, err)

	v, ok := result.(*VoltageInfo)
	if !ok {
		t.Fatal("expected *VoltageInfo")
	}
	assertEqual(t, uint16(1), v.SubsysCount)
	assertEqual(t, 1, len(v.SubsystemVoltages))
	assertEqual(t, byte(0x02), v.SubsystemVoltages[0].SubsysNo)
	assertEqual(t, 2, len(v.SubsystemVoltages[0].Cells))
	assertEqual(t, byte(0x01), v.SubsystemVoltages[0].Cells[0].CellNo)
	assertEqual(t, uint16(100), v.SubsystemVoltages[0].Cells[0].CellInfo.Voltage)
}

func TestDecodeFuelCellData(t *testing.T) {
	data := make([]byte, 21)
	binary.BigEndian.PutUint16(data[0:2], 300)
	binary.BigEndian.PutUint16(data[2:4], 150)
	binary.BigEndian.PutUint16(data[4:6], 0)
	binary.BigEndian.PutUint16(data[6:8], 0)

	result, err := DecodeFuelCellData(data)
	assertNoError(t, err)

	f, ok := result.(*FuelCellInfo)
	if !ok {
		t.Fatal("expected *FuelCellInfo")
	}
	assertEqual(t, uint16(300), f.CellVoltage)
	assertEqual(t, uint16(150), f.CellCurrent)
}

func TestDecodeTemperatureData(t *testing.T) {
	data := []byte{
		0x00, 0x01,
		0x01,
		0x00, 0x03,
		0x28, 0x2A, 0x26,
	}

	result, err := DecodeTemperatureData(data)
	assertNoError(t, err)

	ti, ok := result.(*TemperatureInfo)
	if !ok {
		t.Fatal("expected *TemperatureInfo")
	}
	assertEqual(t, uint16(1), ti.SubsysCount)
	assertEqual(t, 1, len(ti.SubsystemTemperatures))
	assertEqual(t, byte(0x01), ti.SubsystemTemperatures[0].SubsysNo)
	assertEqual(t, uint16(3), ti.SubsystemTemperatures[0].ProbeCount)
	assertEqual(t, 3, len(ti.SubsystemTemperatures[0].ProbeTemperatures))
	assertEqual(t, byte(0x28), ti.SubsystemTemperatures[0].ProbeTemperatures[0])
}
