package gb32960

import (
	"encoding/binary"
	"testing"

	"github.com/darkinno/gb32960-go-sdk/codec"
	"github.com/darkinno/gb32960-go-sdk/constant"
)

func TestStandardFrameStructure(t *testing.T) {
	vin := "GB329601234567890"
	data := []byte{0x01, 0x02}
	pkt := makeMinimalPacket(constant.CmdLogin, vin, data)

	if pkt[0] != 0x23 || pkt[1] != 0x23 {
		t.Fatal("start marker ## (0x23 0x23) mismatch")
	}

	cmd := pkt[2]
	resp := pkt[3]
	enc := pkt[21]
	dLen := binary.BigEndian.Uint16(pkt[22:24])

	if cmd != constant.CmdLogin {
		t.Error("command byte position 3 mismatch")
	}
	if resp != 0x01 {
		t.Error("response byte position 4 mismatch")
	}
	if enc != constant.EncNone {
		t.Error("encrypt byte position 22 mismatch")
	}
	if dLen != 2 {
		t.Error("data length mismatch")
	}
	if len(pkt) != constant.HeaderSize+len(data)+1 {
		t.Errorf("total frame length mismatch: %d != %d", len(pkt), constant.HeaderSize+len(data)+1)
	}

	var bcc byte
	for i := 2; i < len(pkt)-1; i++ {
		bcc ^= pkt[i]
	}
	if bcc != pkt[len(pkt)-1] {
		t.Error("BCC checksum mismatch")
	}
}

func TestStandardCommandCodes(t *testing.T) {
	tests := []struct {
		code byte
		name string
		desc string
	}{
		{constant.CmdLogin, "LOGIN", "vehicle login"},
		{constant.CmdRealtime, "REALTIME", "realtime data"},
		{constant.CmdReissue, "REISSUE", "reissue data"},
		{constant.CmdLogout, "LOGOUT", "vehicle logout"},
		{constant.CmdHeartbeat, "HEARTBEAT", "heartbeat"},
		{constant.CmdTimeCalibr, "TIME_CALIBRATION", "time calibration"},
	}

	for _, tt := range tests {
		if name, ok := constant.CommandNames[tt.code]; !ok {
			t.Errorf("command 0x%02X (%s) not registered", tt.code, tt.desc)
		} else {
			t.Logf("  CMD 0x%02X = %s (%s)", tt.code, name, tt.desc)
		}
	}

	if constant.RespSuccess != 0xFE {
		t.Error("response flag should be 0xFE")
	}
}

func TestStandardEncryptionTypes(t *testing.T) {
	tests := []struct {
		code byte
		desc string
	}{
		{constant.EncNone, "none"},
		{constant.EncRSA, "RSA"},
		{constant.EncAES128, "AES128"},
		{constant.EncAbnormal, "abnormal"},
	}

	for _, tt := range tests {
		t.Logf("  ENC 0x%02X = %s", tt.code, tt.desc)
	}
}

func TestStandardVINValidation(t *testing.T) {
	if constant.VINLength != 17 {
		t.Error("VIN must be 17 bytes per GB/T 16735")
	}
	if !VerifyVIN("ABCDEFGH123456789") {
		t.Error("valid 17-char alphanumeric VIN rejected")
	}
	if VerifyVIN("SHORT") {
		t.Error("VIN < 17 chars accepted")
	}
}

func TestStandardLoginDataFormat(t *testing.T) {
	data := []byte{
		0x26, 0x05, 0x0E, 0x0E, 0x30, 0x00,
		0x00, 0x01,
		0x10,
		'S', 'I', 'M', 'C', 'A', 'R', 'D', '0', '0', '0', '0', '0', '0', '0', '0', '1',
		0x02,
		0x01, 0x01, 0xAA,
		0x02, 0x02, 0xBB, 0xCC,
	}

	login, err := DecodeLoginData(data)
	if err != nil {
		t.Fatal(err)
	}
	if login.Sequence != 1 {
		t.Error("sequence mismatch")
	}
	if login.ICCID != "SIMCARD000000001" {
		t.Errorf("ICCID: %s", login.ICCID)
	}
	if login.ConfigDataCnt != 2 {
		t.Error("config count mismatch")
	}
	if len(login.ConfigData) != 2 {
		t.Error("config data count mismatch")
	}
	if login.ConfigData[0].ID != 0x01 || login.ConfigData[0].Length != 0x01 {
		t.Error("config field 0 mismatch")
	}

	t.Logf("  DateTime=%s Seq=%d ICCID=%s Cfg=%d",
		login.LoginTime.Format("06-01-02 15:04:05"),
		login.Sequence, login.ICCID, login.ConfigDataCnt)
}

func TestStandardLoginResponse(t *testing.T) {
	tests := []struct {
		code byte
		desc string
	}{
		{constant.LoginSuccess, "success"},
		{constant.LoginCarNotFound, "VIN not found"},
		{constant.LoginTerminalNotFound, "terminal not found"},
		{constant.LoginResultAbnormal, "abnormal"},
	}
	for _, tt := range tests {
		if name, ok := constant.LoginResultNames[tt.code]; ok {
			t.Logf("  LOGIN_RESULT 0x%02X = %s (%s)", tt.code, name, tt.desc)
		}
	}

	resp := &LoginResponse{
		Sequence: 1,
		Result:   constant.LoginSuccess,
		Token:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}
	data, err := EncodeLoginResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 9 {
		t.Error("login response too short")
	}
}

func TestStandardRealtimeDataFormat(t *testing.T) {
	data := make([]byte, 0, 27)
	data = append(data, 0x26, 0x05, 0x0E, 0x0E, 0x30, 0x00)
	data = append(data, byte(codec.FieldVehicle))
	v := make([]byte, 20)
	binary.BigEndian.PutUint16(v[3:5], 600)
	binary.BigEndian.PutUint32(v[5:9], 123456)
	binary.BigEndian.PutUint16(v[9:11], 400)
	binary.BigEndian.PutUint16(v[11:13], 500)
	v[13] = 85
	data = append(data, v...)

	msg, err := DecodeRealtimeData(data)
	if err != nil {
		t.Fatal(err)
	}
	if msg.VehicleData == nil {
		t.Fatal("vehicle data not parsed")
	}
	if msg.VehicleData.Speed != 600 {
		t.Error("speed mismatch")
	}
	if msg.VehicleData.SOC != 85 {
		t.Error("SOC mismatch")
	}

	t.Logf("  Time=%s Speed=%.1fkm/h SOC=%d%%",
		msg.InfoTime.Format("15:04:05"),
		float64(msg.VehicleData.Speed)/10,
		msg.VehicleData.SOC)
}

func TestStandardAllDataFields(t *testing.T) {
	fields := []struct {
		id   byte
		name string
	}{
		{codec.FieldVehicle, "vehicle"},
		{codec.FieldMotor, "motor"},
		{codec.FieldFuelCell, "fuel cell"},
		{codec.FieldEngine, "engine"},
		{codec.FieldPosition, "position"},
		{codec.FieldExtreme, "extreme"},
		{codec.FieldAlarm, "alarm"},
		{codec.FieldVoltage, "voltage"},
		{codec.FieldTemperature, "temperature"},
	}

	for _, f := range fields {
		if !codec.HasDecoder(f.id) {
			t.Errorf("Field 0x%02X (%s): missing decoder", f.id, f.name)
		} else {
			t.Logf("  FIELD 0x%02X = %s OK", f.id, f.name)
		}
	}
}

func TestStandardReissueDataFormat(t *testing.T) {
	item := make([]byte, 0, 27)
	item = append(item, 0x26, 0x05, 0x0E, 0x0E, 0x30, 0x00)
	item = append(item, byte(codec.FieldVehicle))
	v := make([]byte, 20)
	binary.BigEndian.PutUint16(v[3:5], 500)
	binary.BigEndian.PutUint32(v[5:9], 100000)
	v[13] = 80
	item = append(item, v...)

	data := make([]byte, 0)
	data = append(data, 0x26, 0x05, 0x0E, 0x0E, 0x30, 0x00)
	data = append(data, 0x00, 0x02)
	data = append(data, item...)
	data = append(data, item...)

	msg, err := DecodeReissueData(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(msg.DataItems) != 2 {
		t.Errorf("expected 2 items, got %d", len(msg.DataItems))
	}
}

func TestStandardLogoutDataFormat(t *testing.T) {
	data := []byte{
		0x26, 0x05, 0x0E, 0x0E, 0x35, 0x00,
		0x00, 0x03,
	}

	lo, err := DecodeLogoutData(data)
	if err != nil {
		t.Fatal(err)
	}
	if lo.Sequence != 3 {
		t.Error("sequence mismatch")
	}
}

func TestStandardHeartbeat(t *testing.T) {
	pkt := makeMinimalPacket(constant.CmdHeartbeat, "TESTVIN1234567890", nil)
	if len(pkt) != constant.HeaderSize+0+1 {
		t.Error("heartbeat should have no data")
	}
	d := NewDecoder()
	defer d.Close()
	d.Feed(pkt)
	decoded, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if decoded == nil {
		t.Fatal("heartbeat not decoded")
	}
	if decoded.Command != constant.CmdHeartbeat {
		t.Error("not heartbeat command")
	}
}

func TestStandardEncodeResponse(t *testing.T) {
	data, err := EncodeResponse(constant.CmdHeartbeat, "TESTVIN1234567890", constant.EncNone, nil)
	if err != nil {
		t.Fatal(err)
	}
	if data[0] != 0x23 || data[1] != 0x23 {
		t.Error("response start marker wrong")
	}
	if data[3] != constant.RespSuccess {
		t.Error("response flag not 0xFE")
	}
}
