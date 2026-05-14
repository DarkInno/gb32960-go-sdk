package gb32960

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/darkinno/gb32960-go-sdk/constant"
)

var (
	ErrInvalidStart    = errors.New("gb32960: invalid start marker")
	ErrInvalidVIN      = errors.New("gb32960: invalid VIN")
	ErrInvalidLength   = errors.New("gb32960: invalid data length")
	ErrInvalidChecksum = errors.New("gb32960: checksum mismatch")
	ErrBufferTooSmall  = errors.New("gb32960: buffer too small")
)

type Decoder struct {
	buf       []byte
	pos       int
	limit     int
	packetBuf *[]byte
}

func NewDecoder() *Decoder {
	return &Decoder{
		buf:       nil,
		packetBuf: getPacketBuffer(),
	}
}

func (d *Decoder) Reset() {
	d.buf = nil
	d.pos = 0
	d.limit = 0
}

func (d *Decoder) Feed(data []byte) {
	if d.buf == nil {
		d.buf = make([]byte, len(data))
		copy(d.buf, data)
		d.pos = 0
		d.limit = len(d.buf)
		return
	}

	remaining := d.limit - d.pos
	if remaining > 0 {
		newBuf := make([]byte, remaining+len(data))
		copy(newBuf, d.buf[d.pos:])
		copy(newBuf[remaining:], data)
		d.buf = newBuf
		d.pos = 0
		d.limit = len(newBuf)
	} else {
		d.buf = make([]byte, len(data))
		copy(d.buf, data)
		d.pos = 0
		d.limit = len(d.buf)
	}
}

func (d *Decoder) Decode() (*Packet, error) {
	for {
		startIdx := d.findStartMarker()
		if startIdx < 0 {
			d.Reset()
			return nil, nil
		}

		d.pos = startIdx

		if d.limit-d.pos < constant.HeaderSize {
			return nil, nil
		}

		header := d.buf[d.pos : d.pos+constant.HeaderSize]
		command := header[2]
		response := header[3]
		vin := string(header[4:21])
		encryptType := header[21]
		dataLength := binary.BigEndian.Uint16(header[22:24])

		d.pos += constant.HeaderSize

		if int(dataLength) > d.limit-d.pos {
			d.pos = startIdx
			return nil, nil
		}

		data := make([]byte, dataLength)
		copy(data, d.buf[d.pos:d.pos+int(dataLength)])
		d.pos += int(dataLength)

		bcc := d.buf[d.pos]
		d.pos++

		calcBCC := calculateBCC(header[2:])
		for _, b := range data {
			calcBCC ^= b
		}
		if calcBCC != bcc {
			d.pos = startIdx + 2
			continue
		}

		pkt := &Packet{
			Command:     command,
			Response:    response,
			VIN:         vin,
			EncryptType: encryptType,
			Data:        data,
			Length:      dataLength,
		}

		return pkt, nil
	}
}

func (d *Decoder) findStartMarker() int {
	for i := d.pos; i < d.limit-1; i++ {
		if d.buf[i] == constant.StartMarker1 && d.buf[i+1] == constant.StartMarker2 {
			return i
		}
	}
	return -1
}

func calculateBCC(data []byte) byte {
	var bcc byte
	for _, b := range data {
		bcc ^= b
	}
	return bcc
}

func (d *Decoder) Close() {
	if d.packetBuf != nil {
		putPacketBuffer(d.packetBuf)
		d.packetBuf = nil
	}
}

func EncodeResponse(command byte, vin string, encryptType byte, data []byte) ([]byte, error) {
	if len(vin) != constant.VINLength {
		return nil, ErrInvalidVIN
	}
	if encryptType == 0 {
		encryptType = constant.EncNone
	}

	dataLen := len(data)
	totalLen := constant.HeaderSize + dataLen + 1

	pkt := make([]byte, totalLen)
	pos := 0

	pkt[pos] = constant.StartMarker1
	pos++
	pkt[pos] = constant.StartMarker2
	pos++

	pkt[pos] = command
	pos++

	pkt[pos] = constant.RespSuccess
	pos++

	copy(pkt[pos:pos+constant.VINLength], []byte(vin))
	pos += constant.VINLength

	pkt[pos] = encryptType
	pos++

	binary.BigEndian.PutUint16(pkt[pos:pos+2], uint16(dataLen))
	pos += 2

	if dataLen > 0 {
		copy(pkt[pos:pos+dataLen], data)
		pos += dataLen
	}

	bcc := calculateBCC(pkt[2:])
	pkt[pos] = bcc

	return pkt, nil
}

func VerifyVIN(vin string) bool {
	if len(vin) != constant.VINLength {
		return false
	}
	for _, c := range vin {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			return false
		}
	}
	return true
}

func DecodeLoginData(data []byte) (*VehicleLoginData, error) {
	if len(data) < 9 {
		return nil, fmt.Errorf("login data too short: %d bytes", len(data))
	}

	pos := 0
	loginTime := parseTime6(data[pos : pos+6])
	pos += 6
	seq := binary.BigEndian.Uint16(data[pos : pos+2])
	pos += 2
	iccidLen := int(data[pos])
	pos++

	if pos+iccidLen > len(data) {
		return nil, fmt.Errorf("ICCID length exceeds data: %d > %d", pos+iccidLen, len(data))
	}
	iccid := string(data[pos : pos+iccidLen])
	pos += iccidLen

	var configCnt byte
	var configData []ConfigField
	if pos < len(data) {
		configCnt = data[pos]
		pos++
		for i := byte(0); i < configCnt && pos+2 <= len(data); i++ {
			cf := ConfigField{
				ID:     data[pos],
				Length: data[pos+1],
			}
			pos += 2
			if pos+int(cf.Length) > len(data) {
				break
			}
			cf.Value = make([]byte, cf.Length)
			copy(cf.Value, data[pos:pos+int(cf.Length)])
			pos += int(cf.Length)
			configData = append(configData, cf)
		}
	}

	return &VehicleLoginData{
		LoginTime:     loginTime,
		Sequence:      seq,
		ICCID:         iccid,
		ConfigDataCnt: configCnt,
		ConfigData:    configData,
	}, nil
}

func EncodeLoginResponse(data *LoginResponse) ([]byte, error) {
	pkt := make([]byte, 0, 64)
	pkt = append(pkt, encodeTime6(data.LoginTime)...)
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, data.Sequence)
	pkt = append(pkt, buf...)
	pkt = append(pkt, data.Result)
	pkt = append(pkt, data.Token...)
	return pkt, nil
}

func DecodeLogoutData(data []byte) (*VehicleLogoutData, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("logout data too short: %d bytes", len(data))
	}
	return &VehicleLogoutData{
		LogoutTime: parseTime6(data[0:6]),
		Sequence:   binary.BigEndian.Uint16(data[6:8]),
	}, nil
}

func parseTime6(b []byte) time.Time {
	return time.Date(
		2000+int(b[0]), time.Month(b[1]), int(b[2]),
		int(b[3]), int(b[4]), int(b[5]),
		0, time.UTC,
	)
}

func encodeTime6(t time.Time) []byte {
	return []byte{
		byte(t.Year() % 100),
		byte(t.Month()),
		byte(t.Day()),
		byte(t.Hour()),
		byte(t.Minute()),
		byte(t.Second()),
	}
}


