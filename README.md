<p align="center">
  <i>"We are DarkInno. Like a stout beer, our best ideas are brewed slowly in the dark, away from the hype."</i>
</p>

---

# gb32960-go-sdk

> A Go SDK for GB/T 32960 — China's national standard communication protocol for electric vehicle remote monitoring.  Built with Go's standard library. Zero external dependencies.

<p align="center">
  <a href="go.mod"><img src="https://img.shields.io/badge/Go-%3E%3D1.22-00ADD8?logo=go" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green" alt="License"></a>
  <a href="https://github.com/darkinno"><img src="https://img.shields.io/badge/DarkInno-open%20source-blue" alt="DarkInno"></a>
  <a href="README_zh.md">中文</a>
</p>

## Features

- **Protocol complete**: vehicle login/logout, realtime & reissue data, heartbeat, time calibration
- **High concurrency**: goroutine-per-connection architecture, tested to 10k+ concurrent connections
- **Zero dependencies**: core SDK uses only Go standard library
- **9 data fields**: vehicle, motor, fuel cell, engine, position, extreme, alarm, voltage, temperature
- **Pluggable auth**: built-in VIN whitelist, custom authenticator interface
- **Event callbacks**: implement the `Handler` interface and receive parsed data — no protocol knowledge needed
- **Message forwarding**: `Forwarder` interface to push data to Kafka, MQTT, or any queue

## Requirements

- Go >= 1.22

## Install

```bash
go get github.com/darkinno/gb32960-go-sdk@v1.0.0
```

## Quick Start

```go
package main

import (
	"context"
	"log"
	"time"

	gb32960 "github.com/darkinno/gb32960-go-sdk"
	"github.com/darkinno/gb32960-go-sdk/auth"
)

type AppHandler struct{}

func (h *AppHandler) OnVehicleLogin(ctx context.Context, conn *gb32960.Connection, msg *gb32960.VehicleLoginData) (*gb32960.LoginResponse, error) {
	log.Printf("Vehicle login: VIN=%s ICCID=%s", conn.VIN(), msg.ICCID)
	return &gb32960.LoginResponse{
		LoginTime: time.Now().UTC(),
		Sequence:  msg.Sequence,
		Result:    0x01,
		Token:     []byte("my-token"),
	}, nil
}

func (h *AppHandler) OnVehicleLogout(ctx context.Context, conn *gb32960.Connection, msg *gb32960.VehicleLogoutData) error {
	log.Printf("Vehicle logout: VIN=%s", conn.VIN())
	return nil
}

func (h *AppHandler) OnRealtimeData(ctx context.Context, conn *gb32960.Connection, msg *gb32960.RealtimeMessage) error {
	if msg.VehicleData != nil {
		log.Printf("[%s] Speed=%.1fkm/h SOC=%d%% Odometer=%.1fkm",
			conn.VIN(),
			float64(msg.VehicleData.Speed)/10,
			msg.VehicleData.SOC,
			float64(msg.VehicleData.Odometer)/10,
		)
	}
	if msg.PositionData != nil {
		log.Printf("[%s] Lat=%.6f Lng=%.6f",
			conn.VIN(),
			float64(msg.PositionData.Latitude)/1e6,
			float64(msg.PositionData.Longitude)/1e6,
		)
	}
	return nil
}

func (h *AppHandler) OnReissueData(ctx context.Context, conn *gb32960.Connection, msg *gb32960.ReissueMessage) error {
	log.Printf("Reissue data: VIN=%s items=%d", conn.VIN(), len(msg.DataItems))
	return nil
}

func (h *AppHandler) OnHeartbeat(ctx context.Context, conn *gb32960.Connection, msg *gb32960.HeartbeatData) error {
	return nil
}

func main() {
	server := gb32960.NewServer(
		gb32960.WithListenAddr(":32960"),
		gb32960.WithMaxConnections(50000),
		gb32960.WithHandler(&AppHandler{}),
		gb32960.WithAuthenticator(&auth.AllowAll{}),
		gb32960.WithReadTimeout(5*time.Minute),
	)

	log.Println("GB32960 SDK listening on :32960")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
```

## Configuration

| Option | Default | Description |
|---|---|---|
| `WithListenAddr(addr)` | `:32960` | TCP listen address |
| `WithMaxConnections(n)` | `10000` | Max connections |
| `WithReadTimeout(d)` | `5m` | Read timeout |
| `WithWriteTimeout(d)` | `10s` | Write timeout |
| `WithIdleTimeout(d)` | `10m` | Idle connection timeout |
| `WithHandler(h)` | nil | Event handler (required for data) |
| `WithAuthenticator(a)` | nil | Authenticator, nil = no auth |
| `WithForwarder(f...)` | nil | Message forwarders |
| `WithTimeProvider(tp)` | nil | Time provider, nil = system UTC |

## Handler Interface

```go
type Handler interface {
	OnVehicleLogin(ctx context.Context, conn *Connection, msg *VehicleLoginData) (*LoginResponse, error)
	OnVehicleLogout(ctx context.Context, conn *Connection, msg *VehicleLogoutData) error
	OnRealtimeData(ctx context.Context, conn *Connection, msg *RealtimeMessage) error
	OnReissueData(ctx context.Context, conn *Connection, msg *ReissueMessage) error
	OnHeartbeat(ctx context.Context, conn *Connection, msg *HeartbeatData) error
}
```

## Authentication

**VIN whitelist:**

```go
whitelist := auth.NewVINWhitelist([]string{
	"VIN00100000000000",
	"VIN00200000000000",
})
server := gb32960.NewServer(gb32960.WithAuthenticator(whitelist))
```

**Allow all:**

```go
server := gb32960.NewServer(gb32960.WithAuthenticator(&auth.AllowAll{}))
```

**Custom auth** — implement `gb32960.Authenticator`:

```go
type Authenticator interface {
	Authenticate(ctx context.Context, vin string) (bool, error)
}
```

## Message Forwarding

Forward parsed data to message queues via build tags:

```bash
go build -tags kafka ./cmd/example/   # Kafka
go build -tags mqtt  ./cmd/example/   # MQTT
```

**Custom** — implement `gb32960.Forwarder`:

```go
type Forwarder interface {
	Forward(ctx context.Context, msg interface{}) error
	Close() error
}
```

## Protocol — GB/T 32960.3-2016

### Commands

| Code | Description | Direction | Status |
|---|---|---|---|
| 0x01 | Vehicle login | T-BOX → Platform | OK |
| 0x02 | Realtime data | T-BOX → Platform | OK |
| 0x03 | Reissue data | T-BOX → Platform | OK |
| 0x04 | Vehicle logout | T-BOX → Platform | OK |
| 0x07 | Heartbeat | T-BOX → Platform | OK |
| 0x08 | Time calibration | T-BOX → Platform | OK |
| 0xFE | Platform response | Platform → T-BOX | Auto |

### Data Fields

| Field | Section | Content |
|---|---|---|
| 0x01 | Vehicle | status/speed/odometer/voltage/current/SOC/gear/insulation (20 bytes) |
| 0x02 | Motor | rpm/torque/temperature/voltage/current |
| 0x03 | Fuel cell | voltage/current/hydrogen concentration/pressure |
| 0x04 | Engine | status/crank speed/fuel rate |
| 0x05 | Position | longitude/latitude |
| 0x06 | Extreme | battery voltage & temperature extremes |
| 0x07 | Alarm | alarm level/content |
| 0x08 | Voltage | subsystem/cell voltages |
| 0x09 | Temperature | subsystem/probe temperatures |

## Standard Compliance

| Item | GB 32960.3-2016 | Implementation |
|---|---|---|
| Frame structure | `##` + CMD(1) + RESP(1) + VIN(17) + ENC(1) + LEN(2) + DATA(N) + BCC(1) | OK |
| Start marker | `0x23 0x23` | OK |
| Command codes | 0x01-0x08 (6 commands) | OK |
| Response flag | request=0x01, response=0xFE | OK |
| Encryption | 0x01/0x02/0x03/0xFE | OK |
| VIN | 17 bytes ASCII | OK |
| BCC | XOR (cmd unit through data unit) | OK |
| Data fields | 0x01-0x09 (9 fields) | OK |
| Time format | 6 bytes BCD (YYMMDDHHMMSS) | OK |

## Concurrency Benchmarks

Environment: Windows x64, Go 1.25, localhost, JSON logging enabled.

| Connections | Workers | Success | Throughput | Avg Latency |
|---|---|---|---|---|
| 500 | 128 | 100% | 3,207 conns/s | 32 ms |
| 2,000 | 128 | 100% | 1,672 conns/s | 73 ms |
| 5,000 | 128 | 100% | 1,118 conns/s | 113 ms |
| 10,000 | 128 | 100% | 1,212 conns/s | 105 ms |

Per connection: login → realtime×2 → heartbeat → logout (5 command exchanges).

## API Reference

### Server

```go
func NewServer(opts ...Option) *Server
func (s *Server) Start() error
func (s *Server) Stop()
func (s *Server) ConnCount() int64
func (s *Server) Connections() []*Connection
func (s *Server) GetConnectionByVIN(vin string) *Connection
func (s *Server) GetConnection(id string) *Connection
```

### Connection

```go
func (c *Connection) ID() string
func (c *Connection) VIN() string
func (c *Connection) ICCID() string
func (c *Connection) RemoteAddr() string
func (c *Connection) State() ConnectionState
func (c *Connection) LastSeen() time.Time
func (c *Connection) Send(cmd byte, data []byte) error
```

## Development

```bash
git clone https://github.com/darkinno/gb32960-go-sdk.git
cd gb32960-go-sdk

go build ./...          # compile
go test ./... -v        # run tests
go vet ./...            # static analysis

go build -o server ./cmd/example/   # build example
```

## Project Structure

```
gb32960-go-sdk/
├── server.go          # server core
├── options.go         # functional options
├── events.go          # handler interface
├── interfaces.go      # authenticator / forwarder
├── decoder.go         # stream decoder (framing/BCC)
├── packet.go          # packet & message types
├── message.go         # message decoding
├── conn.go            # connection management
├── buffer.go          # buffer pools
├── vin_registry.go    # VIN index
├── constant/
│   └── command.go     # protocol constants
├── codec/
│   ├── field.go       # field dispatcher
│   ├── vehicle.go     # vehicle data (20 bytes)
│   ├── motor.go       # motor data
│   ├── battery.go     # fuel cell data
│   ├── engine.go      # engine data
│   ├── position.go    # position data
│   ├── extreme.go     # extreme data
│   ├── alarm.go       # alarm data
│   ├── voltage.go     # voltage data
│   └── temperature.go # temperature data
├── auth/
│   └── auth.go        # VIN whitelist / AllowAll
├── forward/
│   └── config.go      # forwarder config
├── cmd/
│   ├── example/       # usage example
│   ├── simclient/     # T-BOX simulator
│   └── stresstest/    # concurrency stress test
├── go.mod
├── LICENSE
├── README.md
└── README_zh.md
```

## License

[MIT License](LICENSE) — [DarkInno](https://github.com/darkinno)

---

<p align="center">
  If you find this useful, please consider<br>
  <a href="https://github.com/darkinno/gb32960-go-sdk">
    <img src="https://img.shields.io/github/stars/darkinno/gb32960-go-sdk?style=social" alt="GitHub stars">
  </a>
</p>
