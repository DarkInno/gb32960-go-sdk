package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/darkinno/gb32960-go-sdk/codec"
	"github.com/darkinno/gb32960-go-sdk/constant"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:32960")
	if err != nil {
		log.Fatalf("connect failed: %v", err)
	}
	defer conn.Close()

	vin := "SIMUVIN1234567890"
	log.Printf("=== Simulated T-BOX ===")
	log.Printf("VIN: %s", vin)

	// 1. Login
	loginData := []byte{
		0x26, 0x05, 0x0E, 0x0E, 0x30, 0x00,
		0x00, 0x01,
		0x10,
		'S', 'I', 'M', 'C', 'A', 'R', 'D', '0', '0', '0', '0', '0', '0', '0', '0', '1',
		0x00,
	}
	sendPacket(conn, constant.CmdLogin, vin, loginData, "LOGIN")
	time.Sleep(300 * time.Millisecond)

	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	n, _ := conn.Read(buf)
	if n > 0 {
		log.Printf("Login resp: %d bytes | CMD=0x%02X RESP=0x%02X", n, buf[2], buf[3])
	}

	time.Sleep(1 * time.Second)

	// 2. Realtime data x3
	for i := 1; i <= 3; i++ {
		realtimeData := buildRealtimeData(i)
		sendPacket(conn, constant.CmdRealtime, vin, realtimeData, fmt.Sprintf("REALTIME #%d", i))
		time.Sleep(2 * time.Second)
	}

	// 3. Heartbeat x2
	for i := 1; i <= 2; i++ {
		sendPacket(conn, constant.CmdHeartbeat, vin, nil, fmt.Sprintf("HEARTBEAT #%d", i))
		time.Sleep(500 * time.Millisecond)
		conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
		n, _ := conn.Read(buf)
		if n > 0 {
			log.Printf("Heartbeat resp: %d bytes", n)
		}
		time.Sleep(2 * time.Second)
	}

	// 4. Logout
	logoutData := []byte{
		0x26, 0x05, 0x0E, 0x0E, 0x35, 0x00,
		0x00, 0x03,
	}
	sendPacket(conn, constant.CmdLogout, vin, logoutData, "LOGOUT")

	time.Sleep(500 * time.Millisecond)
	log.Printf("=== Done ===")
}

func buildRealtimeData(seq int) []byte {
	data := make([]byte, 0, 100)
	data = append(data, 0x26, 0x05, 0x0E, 0x0E, 0x30, byte(seq*10))

	data = append(data, codec.FieldVehicle)
	v := make([]byte, 20)
	v[0] = 0x01
	v[1] = 0x01
	v[2] = 0x01
	binary.BigEndian.PutUint16(v[3:5], uint16(400+seq*50))
	binary.BigEndian.PutUint32(v[5:9], uint32(123450+seq*100))
	binary.BigEndian.PutUint16(v[9:11], uint16(380+seq))
	binary.BigEndian.PutUint16(v[11:13], uint16(200+seq*10))
	v[13] = byte(80 + seq)
	v[14] = 0x01
	v[15] = 0x03
	binary.BigEndian.PutUint16(v[16:18], 9999)
	data = append(data, v...)

	data = append(data, codec.FieldPosition)
	p := make([]byte, 8)
	binary.BigEndian.PutUint32(p[0:4], uint32(121473708+uint32(seq*1000)))
	binary.BigEndian.PutUint32(p[4:8], uint32(31235538+uint32(seq*500)))
	data = append(data, p...)

	data = append(data, codec.FieldExtreme)
	e := make([]byte, 10)
	binary.BigEndian.PutUint16(e[0:2], uint16(4100+uint16(seq*10)))
	e[2] = 0x05
	binary.BigEndian.PutUint16(e[3:5], uint16(3700))
	e[5] = 0x10
	e[6] = byte(40 + seq)
	e[7] = 0x03
	e[8] = byte(15)
	e[9] = 0x1F
	data = append(data, e...)

	return data
}

func sendPacket(conn net.Conn, command byte, vin string, data []byte, label string) {
	pkt := buildPacket(command, vin, data)
	_, err := conn.Write(pkt)
	if err != nil {
		log.Printf("[%s] send failed: %v", label, err)
		return
	}
	log.Printf("[%s] sent %d bytes", label, len(pkt))
}

func buildPacket(command byte, vin string, data []byte) []byte {
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
