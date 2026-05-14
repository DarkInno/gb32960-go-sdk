package gb32960

import (
	"encoding/binary"
	"testing"

	"github.com/darkinno/gb32960-go-sdk/codec"
	"github.com/darkinno/gb32960-go-sdk/constant"
)

func TestDecodeRealtimeVehicleOnly(t *testing.T) {
	data := make([]byte, 0)
	data = append(data, 0x16, 0x01, 0x01, 0x0C, 0x00, 0x00)

	data = append(data, codec.FieldVehicle)
	vehicleBytes := make([]byte, 20)
	vehicleBytes[0] = 0x01
	vehicleBytes[2] = 0x01
	binary.BigEndian.PutUint16(vehicleBytes[3:5], 600)
	binary.BigEndian.PutUint32(vehicleBytes[5:9], 123456)
	binary.BigEndian.PutUint16(vehicleBytes[9:11], 400)
	binary.BigEndian.PutUint16(vehicleBytes[11:13], 500)
	vehicleBytes[13] = 85
	binary.BigEndian.PutUint16(vehicleBytes[16:18], 9999)
	data = append(data, vehicleBytes...)

	msg, err := DecodeRealtimeData(data)
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("msg is nil")
	}
	if msg.VehicleData == nil {
		t.Fatal("vehicle data is nil")
	}
	if msg.VehicleData.Speed != 600 {
		t.Errorf("speed: %d", msg.VehicleData.Speed)
	}
	if msg.VehicleData.SOC != 85 {
		t.Errorf("SOC: %d", msg.VehicleData.SOC)
	}
}

func TestDecodeRealtimeWithPosition(t *testing.T) {
	data := make([]byte, 0)
	data = append(data, 0x16, 0x01, 0x01, 0x0C, 0x00, 0x00)

	data = append(data, codec.FieldPosition)
	posBytes := make([]byte, 8)
	binary.BigEndian.PutUint32(posBytes[0:4], 121473708)
	binary.BigEndian.PutUint32(posBytes[4:8], 31235538)
	data = append(data, posBytes...)

	msg, err := DecodeRealtimeData(data)
	if err != nil {
		t.Fatal(err)
	}
	if msg.PositionData == nil {
		t.Fatal("position data is nil")
	}
	if msg.PositionData.Longitude != 121473708 {
		t.Errorf("longitude: %d", msg.PositionData.Longitude)
	}
	if msg.PositionData.Latitude != 31235538 {
		t.Errorf("latitude: %d", msg.PositionData.Latitude)
	}
}

func TestDecodeRealtimeMotors(t *testing.T) {
	data := make([]byte, 0)
	data = append(data, 0x16, 0x01, 0x01, 0x0C, 0x00, 0x00)

	data = append(data, codec.FieldMotor)
	data = append(data, byte(1))
	motorBytes := []byte{
		0x01, 0x02, 0x28,
		0x13, 0x88, 0x03, 0xE8,
		0x32, 0x01, 0x90, 0x00, 0x64,
	}
	data = append(data, motorBytes...)

	msg, err := DecodeRealtimeData(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(msg.MotorData) != 1 {
		t.Fatalf("expected 1 motor, got %d", len(msg.MotorData))
	}
	if msg.MotorData[0].MotorSpeed != 5000 {
		t.Errorf("motor speed: %d", msg.MotorData[0].MotorSpeed)
	}
}

func TestEncodeLoginResponse(t *testing.T) {
	resp := &LoginResponse{
		LoginTime: parseTime6([]byte{0x16, 0x01, 0x01, 0x0C, 0x00, 0x00}),
		Sequence:  1,
		Result:    constant.LoginSuccess,
		Token:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}

	data, err := EncodeLoginResponse(resp)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := DecodeLoginData(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Sequence != 1 {
		t.Errorf("sequence: %d", decoded.Sequence)
	}
}

func TestRealtimeWithMultipleFields(t *testing.T) {
	data := make([]byte, 0)
	data = append(data, 0x16, 0x01, 0x01, 0x0C, 0x00, 0x00)

	data = append(data, codec.FieldVehicle)
	vehicleBytes := make([]byte, 20)
	binary.BigEndian.PutUint16(vehicleBytes[3:5], 500)
	binary.BigEndian.PutUint32(vehicleBytes[5:9], 100000)
	binary.BigEndian.PutUint16(vehicleBytes[9:11], 380)
	data = append(data, vehicleBytes...)

	data = append(data, codec.FieldEngine)
	engineBytes := make([]byte, 5)
	engineBytes[0] = 0x01
	binary.BigEndian.PutUint16(engineBytes[1:3], 2000)
	binary.BigEndian.PutUint16(engineBytes[3:5], 120)
	data = append(data, engineBytes...)

	data = append(data, codec.FieldExtreme)
	extremeBytes := make([]byte, 10)
	binary.BigEndian.PutUint16(extremeBytes[0:2], 4200)
	binary.BigEndian.PutUint16(extremeBytes[3:5], 3800)
	extremeBytes[6] = 45
	extremeBytes[8] = 15
	data = append(data, extremeBytes...)

	msg, err := DecodeRealtimeData(data)
	if err != nil {
		t.Fatal(err)
	}

	if msg.VehicleData == nil {
		t.Error("vehicle data nil")
	} else if msg.VehicleData.Speed != 500 {
		t.Errorf("vehicle speed: %d", msg.VehicleData.Speed)
	}

	if msg.EngineData == nil {
		t.Error("engine data nil")
	} else if msg.EngineData.CrankSpeed != 2000 {
		t.Errorf("crank speed: %d", msg.EngineData.CrankSpeed)
	}

	if msg.ExtremeData == nil {
		t.Error("extreme data nil")
	} else if msg.ExtremeData.MaxBatteryVoltage != 4200 {
		t.Errorf("max voltage: %d", msg.ExtremeData.MaxBatteryVoltage)
	}
}
