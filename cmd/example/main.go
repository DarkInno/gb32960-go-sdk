package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	gb32960 "github.com/darkinno/gb32960-go-sdk"
	"github.com/darkinno/gb32960-go-sdk/auth"
)

type MyHandler struct{}

func (h *MyHandler) OnVehicleLogin(ctx context.Context, conn *gb32960.Connection, msg *gb32960.VehicleLoginData) (*gb32960.LoginResponse, error) {
	log.Printf("[LOGIN] VIN=%s ICCID=%s Seq=%d Time=%s",
		conn.VIN(), msg.ICCID, msg.Sequence, msg.LoginTime.Format(time.RFC3339))

	token := make([]byte, 8)
	copy(token, []byte("TOKEN123"))

	return &gb32960.LoginResponse{
		LoginTime: time.Now().UTC(),
		Sequence:  msg.Sequence,
		Result:    0x01,
		Token:     token,
	}, nil
}

func (h *MyHandler) OnVehicleLogout(ctx context.Context, conn *gb32960.Connection, msg *gb32960.VehicleLogoutData) error {
	log.Printf("[LOGOUT] VIN=%s Time=%s", conn.VIN(), msg.LogoutTime.Format(time.RFC3339))
	return nil
}

func (h *MyHandler) OnRealtimeData(ctx context.Context, conn *gb32960.Connection, msg *gb32960.RealtimeMessage) error {
	vin := conn.VIN()
	log.Printf("[REALTIME] VIN=%s Time=%s", vin, msg.InfoTime.Format(time.RFC3339))

	if msg.VehicleData != nil {
		d := msg.VehicleData
		log.Printf("  Vehicle: Speed=%.1fkm/h Odometer=%.1fkm SOC=%d%% Voltage=%.1fV",
			float64(d.Speed)/10.0, float64(d.Odometer)/10.0, d.SOC, float64(d.TotalVoltage))
	}

	if msg.MotorData != nil {
		log.Printf("  Motors: %d", len(msg.MotorData))
		for _, m := range msg.MotorData {
			log.Printf("    Motor#%d: Speed=%drpm Torque=%.1fNm Temp=%dC",
				m.MotorSeq, m.MotorSpeed, float64(m.MotorTorque)/10.0, m.MotorTemp)
		}
	}

	if msg.PositionData != nil {
		lat := float64(msg.PositionData.Latitude) / 1e6
		lng := float64(msg.PositionData.Longitude) / 1e6
		log.Printf("  Position: Lat=%.6f Lng=%.6f", lat, lng)
	}

	if msg.ExtremeData != nil {
		e := msg.ExtremeData
		log.Printf("  Battery: Vmax=%.3fV Vmin=%.3fV Tmax=%dC Tmin=%dC",
			float64(e.MaxBatteryVoltage)/1000.0, float64(e.MinBatteryVoltage)/1000.0,
			e.MaxTemp, e.MinTemp)
	}

	if msg.AlarmData != nil {
		log.Printf("  Alarm: Level=%d Bytes=%d", msg.AlarmData.MaxLevel, msg.AlarmData.AlarmByteLen)
	}

	return nil
}

func (h *MyHandler) OnReissueData(ctx context.Context, conn *gb32960.Connection, msg *gb32960.ReissueMessage) error {
	log.Printf("[REISSUE] VIN=%s Items=%d", conn.VIN(), len(msg.DataItems))
	return nil
}

func (h *MyHandler) OnHeartbeat(ctx context.Context, conn *gb32960.Connection, msg *gb32960.HeartbeatData) error {
	log.Printf("[HEARTBEAT] VIN=%s", conn.VIN())
	return nil
}

func main() {
	handler := &MyHandler{}

	server := gb32960.NewServer(
		gb32960.WithListenAddr(":32960"),
		gb32960.WithMaxConnections(50000),
		gb32960.WithHandler(handler),
		gb32960.WithAuthenticator(&auth.AllowAll{}),
		gb32960.WithReadTimeout(5*time.Minute),
		gb32960.WithWriteTimeout(10*time.Second),
		gb32960.WithIdleTimeout(10*time.Minute),
	)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down...")
		server.Stop()
		os.Exit(0)
	}()

	log.Println("GB32960 SDK server listening on :32960")
	log.Println("  Max connections: 50000")
	log.Println("  Auth: AllowAll")

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			log.Printf("[STATS] Active connections: %d", server.ConnCount())
		}
	}()

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
