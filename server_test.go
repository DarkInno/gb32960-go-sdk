package gb32960

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/darkinno/gb32960-go-sdk/codec"
	"github.com/darkinno/gb32960-go-sdk/constant"
)

type testHandler struct {
	loginCalls    int
	logoutCalls   int
	realtimeCalls int
	heartbeatCalls int
	lastMessage   *RealtimeMessage
}

func (h *testHandler) OnVehicleLogin(ctx context.Context, conn *Connection, msg *VehicleLoginData) (*LoginResponse, error) {
	h.loginCalls++
	return &LoginResponse{
		LoginTime: time.Now().UTC(),
		Sequence:  msg.Sequence,
		Result:    constant.LoginSuccess,
		Token:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}, nil
}

func (h *testHandler) OnVehicleLogout(ctx context.Context, conn *Connection, msg *VehicleLogoutData) error {
	h.logoutCalls++
	return nil
}

func (h *testHandler) OnRealtimeData(ctx context.Context, conn *Connection, msg *RealtimeMessage) error {
	h.realtimeCalls++
	h.lastMessage = msg
	return nil
}

func (h *testHandler) OnReissueData(ctx context.Context, conn *Connection, msg *ReissueMessage) error {
	return nil
}

func (h *testHandler) OnHeartbeat(ctx context.Context, conn *Connection, msg *HeartbeatData) error {
	h.heartbeatCalls++
	return nil
}

func TestServerStartStop(t *testing.T) {
	s := NewServer(
		WithListenAddr(":33960"),
		WithMaxConnections(10),
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	// Verify server is listening
	conn, err := net.Dial("tcp", ":33960")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	conn.Close()

	time.Sleep(50 * time.Millisecond)

	if s.ConnCount() != 0 {
		t.Errorf("expected 0 connections after close, got %d", s.ConnCount())
	}

	s.Stop()
}

func TestServerConnectionAndLogin(t *testing.T) {
	h := &testHandler{}
	s := NewServer(
		WithListenAddr(":33961"),
		WithMaxConnections(10),
		WithHandler(h),
		WithReadTimeout(500*time.Millisecond),
		WithWriteTimeout(500*time.Millisecond),
	)

	go s.Start()
	time.Sleep(30 * time.Millisecond)
	defer s.Stop()

	// Connect via TCP
	netConn, err := net.Dial("tcp", ":33961")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer netConn.Close()

	time.Sleep(50 * time.Millisecond)

	if s.ConnCount() != 1 {
		t.Errorf("expected 1 connection, got %d", s.ConnCount())
	}

	// Send login packet
	loginData := []byte{
		0x16, 0x01, 0x01, 0x0C, 0x00, 0x00,
		0x00, 0x01,
		0x0A,
		'1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
		0x00,
	}
	loginPkt := makeMinimalPacket(constant.CmdLogin, "TESTVIN1234567890", loginData)
	_, err = netConn.Write(loginPkt)
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if h.loginCalls != 1 {
		t.Errorf("expected 1 login call, got %d", h.loginCalls)
	}

	// Read response
	respBuf := make([]byte, 1024)
	netConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	n, err := netConn.Read(respBuf)
	if err != nil {
		t.Logf("read response error (may be ok if server not responding): %v", err)
	}
	if n > 0 && respBuf[2] == constant.CmdLogin && respBuf[3] == constant.RespSuccess {
		t.Log("got login response from server")
	}
}

func TestServerLoginRejected(t *testing.T) {
	h := &testHandler{}
	whitelist := &mockAuth{allow: false}

	s := NewServer(
		WithListenAddr(":33962"),
		WithMaxConnections(10),
		WithHandler(h),
		WithAuthenticator(whitelist),
		WithReadTimeout(500*time.Millisecond),
	)

	go s.Start()
	time.Sleep(30 * time.Millisecond)
	defer s.Stop()

	netConn, err := net.Dial("tcp", ":33962")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer netConn.Close()

	time.Sleep(30 * time.Millisecond)

	loginPkt := makeMinimalPacket(constant.CmdLogin, "TESTVIN1234567890", []byte{
		0x16, 0x01, 0x01, 0x0C, 0x00, 0x00,
		0x00, 0x01,
		0x00,
	})
	netConn.Write(loginPkt)

	time.Sleep(200 * time.Millisecond)

	if h.loginCalls != 0 {
		t.Errorf("login should be rejected, got %d calls", h.loginCalls)
	}
}

func TestServerHeartbeat(t *testing.T) {
	h := &testHandler{}
	s := NewServer(
		WithListenAddr(":33963"),
		WithMaxConnections(10),
		WithHandler(h),
		WithReadTimeout(500*time.Millisecond),
	)

	go s.Start()
	time.Sleep(30 * time.Millisecond)
	defer s.Stop()

	netConn, err := net.Dial("tcp", ":33963")
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer netConn.Close()

	time.Sleep(30 * time.Millisecond)

	// Login first: datetime(6) + seq(2) + iccidLen(1) = 9 bytes minimum
	loginData := []byte{
		0x16, 0x01, 0x01, 0x0C, 0x00, 0x00,
		0x00, 0x01,
		0x00,
	}
	loginPkt := makeMinimalPacket(constant.CmdLogin, "TESTVIN1234567890", loginData)
	netConn.Write(loginPkt)
	time.Sleep(50 * time.Millisecond)

	// Send heartbeat
	heartbeatPkt := makeMinimalPacket(constant.CmdHeartbeat, "TESTVIN1234567890", nil)
	netConn.Write(heartbeatPkt)

	time.Sleep(100 * time.Millisecond)

	if h.heartbeatCalls != 1 {
		t.Errorf("expected 1 heartbeat call, got %d", h.heartbeatCalls)
	}
}

func TestDecodeReissueMultipleItems(t *testing.T) {
	// Build item 1: time(6) + field_vehicle(1) + vehicle_data(20) = 27 bytes
	item1 := make([]byte, 0, 27)
	item1 = append(item1, 0x16, 0x01, 0x01, 0x0C, 0x00, 0x00) // datetime
	item1 = append(item1, codec.FieldVehicle)
	v := make([]byte, 20)
	binary.BigEndian.PutUint16(v[3:5], 500)
	binary.BigEndian.PutUint32(v[5:9], 100000)
	v[13] = 80
	item1 = append(item1, v...)

	// Build item 2: time(6) + field_position(1) + position_data(8) = 15 bytes
	item2 := make([]byte, 0, 15)
	item2 = append(item2, 0x16, 0x02, 0x01, 0x0C, 0x30, 0x00) // datetime
	item2 = append(item2, codec.FieldPosition)
	p := make([]byte, 8)
	binary.BigEndian.PutUint32(p[0:4], 121000000)
	binary.BigEndian.PutUint32(p[4:8], 31000000)
	item2 = append(item2, p...)

	// Assemble reissue data: time(6) + count(2) + items
	data := make([]byte, 0)
	data = append(data, 0x16, 0x01, 0x01, 0x0C, 0x00, 0x00) // reissue time
	data = append(data, 0x00, 0x02)                          // count = 2
	data = append(data, item1...)
	data = append(data, item2...)

	t.Logf("data length: %d", len(data))

	msg, err := DecodeReissueData(data)
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("msg is nil")
	}

	t.Logf("items decoded: %d", len(msg.DataItems))

	if len(msg.DataItems) != 2 {
		t.Fatalf("expected 2 items, got %d", len(msg.DataItems))
	}

	if msg.DataItems[0].VehicleData == nil {
		t.Error("item 0: vehicle data nil")
	} else {
		t.Logf("item 0 vehicle: speed=%d SOC=%d", msg.DataItems[0].VehicleData.Speed, msg.DataItems[0].VehicleData.SOC)
	}

	if msg.DataItems[1].PositionData == nil {
		t.Error("item 1: position data nil")
	} else {
		t.Logf("item 1 position: lng=%d lat=%d", msg.DataItems[1].PositionData.Longitude, msg.DataItems[1].PositionData.Latitude)
	}
}

func TestForwardMsg(t *testing.T) {
	msg := newForwardMsg("login", "TESTVIN1234567890", "data")
	if msg.Type != "login" {
		t.Errorf("type: %s", msg.Type)
	}
	if msg.VIN != "TESTVIN1234567890" {
		t.Errorf("vin: %s", msg.VIN)
	}
	if msg.Data != "data" {
		t.Errorf("data: %v", msg.Data)
	}
}

type mockAuth struct {
	allow bool
}

func (m *mockAuth) Authenticate(ctx context.Context, vin string) (bool, error) {
	return m.allow, nil
}
