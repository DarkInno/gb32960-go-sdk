package gb32960

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/darkinno/gb32960-go-sdk/constant"
)

func makeMinimalPacket(command byte, vin string, data []byte) []byte {
	vinBytes := make([]byte, constant.VINLength)
	copy(vinBytes, []byte(vin))

	totalLen := constant.HeaderSize + len(data) + 1
	pkt := make([]byte, totalLen)
	pos := 0

	pkt[pos] = constant.StartMarker1
	pos++
	pkt[pos] = constant.StartMarker2
	pos++
	pkt[pos] = command
	pos++
	pkt[pos] = 0x01
	pos++
	copy(pkt[pos:pos+constant.VINLength], vinBytes)
	pos += constant.VINLength
	pkt[pos] = constant.EncNone
	pos++
	binary.BigEndian.PutUint16(pkt[pos:pos+2], uint16(len(data)))
	pos += 2
	copy(pkt[pos:pos+len(data)], data)
	pos += len(data)

	var bcc byte
	for i := 2; i < pos; i++ {
		bcc ^= pkt[i]
	}
	pkt[pos] = bcc
	return pkt
}

func TestDecoderSinglePacket(t *testing.T) {
	vin := "TESTVIN1234567890"
	data := []byte{0x01, 0x02, 0x03}
	raw := makeMinimalPacket(constant.CmdHeartbeat, vin, data)

	d := NewDecoder()
	defer d.Close()

	d.Feed(raw)
	pkt, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if pkt == nil {
		t.Fatal("pkt is nil")
	}

	if pkt.Command != constant.CmdHeartbeat {
		t.Errorf("Command: want %x got %x", constant.CmdHeartbeat, pkt.Command)
	}
	if pkt.VIN != vin {
		t.Errorf("VIN: want %s got %s", vin, pkt.VIN)
	}
	if pkt.EncryptType != constant.EncNone {
		t.Errorf("EncryptType: want %x got %x", constant.EncNone, pkt.EncryptType)
	}
	if pkt.Length != 3 {
		t.Errorf("Length: want 3 got %d", pkt.Length)
	}
}

func TestDecoderTwoPackets(t *testing.T) {
	vin := "TESTVIN1234567890"
	raw1 := makeMinimalPacket(constant.CmdHeartbeat, vin, []byte{0x01, 0x02})
	raw2 := makeMinimalPacket(constant.CmdLogin, vin, []byte{0x03, 0x04, 0x05})
	combined := append(raw1, raw2...)

	d := NewDecoder()
	defer d.Close()
	d.Feed(combined)

	pkt, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if pkt == nil {
		t.Fatal("first pkt is nil")
	}
	if pkt.Command != constant.CmdHeartbeat {
		t.Errorf("first Command: want %x got %x", constant.CmdHeartbeat, pkt.Command)
	}

	pkt, err = d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if pkt == nil {
		t.Fatal("second pkt is nil")
	}
	if pkt.Command != constant.CmdLogin {
		t.Errorf("second Command: want %x got %x", constant.CmdLogin, pkt.Command)
	}

	pkt, err = d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if pkt != nil {
		t.Error("third decode should return nil")
	}
}

func TestDecoderFragmented(t *testing.T) {
	vin := "TESTVIN1234567890"
	data := []byte{0x0A, 0x0B, 0x0C, 0x0D, 0x0E}
	raw := makeMinimalPacket(constant.CmdRealtime, vin, data)
	splitPoint := len(raw) / 2

	d := NewDecoder()
	defer d.Close()

	d.Feed(raw[:splitPoint])
	pkt, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if pkt != nil {
		t.Error("should be nil with partial data")
	}

	d.Feed(raw[splitPoint:])
	pkt, err = d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if pkt == nil {
		t.Fatal("pkt is nil after full data")
	}
	if pkt.Command != constant.CmdRealtime {
		t.Errorf("Command: want %x got %x", constant.CmdRealtime, pkt.Command)
	}
}

func TestDecoderBadChecksumSkip(t *testing.T) {
	vin := "TESTVIN1234567890"
	data := []byte{0x01, 0x02}

	raw := makeMinimalPacket(constant.CmdHeartbeat, vin, data)
	raw[len(raw)-1] ^= 0xFF

	raw2 := makeMinimalPacket(constant.CmdHeartbeat, vin, data)
	combined := append(raw, raw2...)

	d := NewDecoder()
	defer d.Close()
	d.Feed(combined)

	pkt, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if pkt == nil {
		t.Fatal("pkt is nil")
	}
	if pkt.Command != constant.CmdHeartbeat {
		t.Errorf("Command: want %x got %x", constant.CmdHeartbeat, pkt.Command)
	}
}

func TestEncodeResponse(t *testing.T) {
	vin := "TESTVIN1234567890"
	pkt, err := EncodeResponse(constant.CmdLogin, vin, constant.EncNone, []byte{0x01, 0x02})
	if err != nil {
		t.Fatal(err)
	}
	if pkt[0] != constant.StartMarker1 {
		t.Error("bad start marker 1")
	}
	if pkt[1] != constant.StartMarker2 {
		t.Error("bad start marker 2")
	}
	if pkt[2] != constant.CmdLogin {
		t.Error("bad command")
	}
	if pkt[3] != constant.RespSuccess {
		t.Error("bad response flag")
	}
	if pkt[21] != constant.EncNone {
		t.Error("bad encrypt type")
	}
	if binary.BigEndian.Uint16(pkt[22:24]) != 2 {
		t.Error("bad data length")
	}
}

func TestDecodeLoginData(t *testing.T) {
	data := []byte{
		0x16, 0x01, 0x01, 0x0C, 0x00, 0x00,
		0x00, 0x01,
		0x0A,
		'1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
		0x02,
		0x01, 0x01, 0xAA,
		0x02, 0x02, 0xBB, 0xCC,
	}

	loginData, err := DecodeLoginData(data)
	if err != nil {
		t.Fatal(err)
	}
	if loginData.LoginTime.Year() != 2022 {
		t.Error("bad year")
	}
	if loginData.LoginTime.Month() != time.January {
		t.Error("bad month")
	}
	if loginData.LoginTime.Day() != 1 {
		t.Error("bad day")
	}
	if loginData.LoginTime.Hour() != 12 {
		t.Error("bad hour")
	}
	if loginData.Sequence != 1 {
		t.Error("bad sequence")
	}
	if loginData.ICCID != "1234567890" {
		t.Error("bad ICCID")
	}
	if loginData.ConfigDataCnt != 2 {
		t.Error("bad config count")
	}
	if len(loginData.ConfigData) != 2 {
		t.Error("bad config data len")
	}
}

func TestDecodeLogoutData(t *testing.T) {
	data := []byte{
		0x16, 0x06, 0x0F, 0x12, 0x30, 0x00,
		0x00, 0x05,
	}

	logoutData, err := DecodeLogoutData(data)
	if err != nil {
		t.Fatal(err)
	}
	if logoutData.LogoutTime.Year() != 2022 {
		t.Error("bad year")
	}
	if logoutData.LogoutTime.Month() != time.June {
		t.Error("bad month")
	}
	if logoutData.Sequence != 5 {
		t.Error("bad sequence")
	}
}

func TestVerifyVIN(t *testing.T) {
	if !VerifyVIN("TESTVIN1234567890") {
		t.Error("valid VIN rejected")
	}
	if VerifyVIN("SHORT") {
		t.Error("short VIN accepted")
	}
	if VerifyVIN("TESTVIN123456789012345") {
		t.Error("long VIN accepted")
	}
	if VerifyVIN("TESTVIN123456789@") {
		t.Error("invalid char VIN accepted")
	}
}
