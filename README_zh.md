<p align="center">
  <i>"We are DarkInno. Like a stout beer, our best ideas are brewed slowly in the dark, away from the hype."</i>
</p>

---

# gb32960-go-sdk

> GB/T 32960 电动汽车远程服务与管理系统通信协议 Go SDK — by [DarkInno](https://github.com/darkinno)
>
> Go 标准库编写，零外部依赖。接收 T-BOX TCP 连接，解析协议数据，回调分发。

<p align="center">
  <a href="go.mod"><img src="https://img.shields.io/badge/Go-%3E%3D1.22-00ADD8?logo=go" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green" alt="License"></a>
  <a href="https://github.com/darkinno"><img src="https://img.shields.io/badge/DarkInno-open%20source-blue" alt="DarkInno"></a>
  <a href="README.md">English</a>
</p>

## 功能特性

- **协议完整**：车辆登录/登出、实时/补发数据上报、心跳、终端校时
- **高并发**：goroutine-per-connection 架构，实测 10k+ 并发连接
- **零依赖**：核心 SDK 仅用 Go 标准库
- **9 种数据字段**：整车/电机/燃料电池/发动机/位置/极值/报警/电压/温度
- **认证可插拔**：内置 VIN 白名单，可自定义认证逻辑
- **事件回调**：实现 `Handler` 接口即可接收数据，无需关心协议细节
- **消息转发**：`Forwarder` 接口可推送数据至 Kafka、MQTT 等

## 环境要求

- Go >= 1.22

## 安装

```bash
go get github.com/darkinno/gb32960-go-sdk@v1.0.0
```

## 快速开始

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
	log.Printf("车辆登录: VIN=%s ICCID=%s", conn.VIN(), msg.ICCID)
	return &gb32960.LoginResponse{
		LoginTime: time.Now().UTC(),
		Sequence:  msg.Sequence,
		Result:    0x01,
		Token:     []byte("my-token"),
	}, nil
}

func (h *AppHandler) OnVehicleLogout(ctx context.Context, conn *gb32960.Connection, msg *gb32960.VehicleLogoutData) error {
	log.Printf("车辆登出: VIN=%s", conn.VIN())
	return nil
}

func (h *AppHandler) OnRealtimeData(ctx context.Context, conn *gb32960.Connection, msg *gb32960.RealtimeMessage) error {
	if msg.VehicleData != nil {
		log.Printf("[%s] 速度=%.1fkm/h SOC=%d%% 里程=%.1fkm",
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
	log.Printf("补发数据: VIN=%s 条数=%d", conn.VIN(), len(msg.DataItems))
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

	log.Println("GB32960 SDK 监听 :32960")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
```

## 配置选项

| 选项 | 默认值 | 说明 |
|---|---|---|
| `WithListenAddr(addr)` | `:32960` | TCP 监听地址 |
| `WithMaxConnections(n)` | `10000` | 最大连接数 |
| `WithReadTimeout(d)` | `5m` | 读超时 |
| `WithWriteTimeout(d)` | `10s` | 写超时 |
| `WithIdleTimeout(d)` | `10m` | 空闲超时 |
| `WithHandler(h)` | nil | 事件处理器（必须设置以接收数据） |
| `WithAuthenticator(a)` | nil | 认证器，nil 则不认证 |
| `WithForwarder(f...)` | nil | 消息转发器，可设置多个 |
| `WithTimeProvider(tp)` | nil | 时间提供者，nil 则用系统 UTC |

## Handler 接口

```go
type Handler interface {
	OnVehicleLogin(ctx context.Context, conn *Connection, msg *VehicleLoginData) (*LoginResponse, error)
	OnVehicleLogout(ctx context.Context, conn *Connection, msg *VehicleLogoutData) error
	OnRealtimeData(ctx context.Context, conn *Connection, msg *RealtimeMessage) error
	OnReissueData(ctx context.Context, conn *Connection, msg *ReissueMessage) error
	OnHeartbeat(ctx context.Context, conn *Connection, msg *HeartbeatData) error
}
```

## 认证

**VIN 白名单：**

```go
whitelist := auth.NewVINWhitelist([]string{
	"VIN00100000000000",
	"VIN00200000000000",
})
server := gb32960.NewServer(gb32960.WithAuthenticator(whitelist))
```

**放行所有连接：**

```go
server := gb32960.NewServer(gb32960.WithAuthenticator(&auth.AllowAll{}))
```

**自定义认证** — 实现 `gb32960.Authenticator`：

```go
type Authenticator interface {
	Authenticate(ctx context.Context, vin string) (bool, error)
}
```

## 消息转发

将解析后数据推送到消息队列，通过编译标签启用：

```bash
go build -tags kafka ./cmd/example/   # Kafka 转发
go build -tags mqtt  ./cmd/example/   # MQTT 转发
```

**自定义** — 实现 `gb32960.Forwarder`：

```go
type Forwarder interface {
	Forward(ctx context.Context, msg interface{}) error
	Close() error
}
```

## 协议 — GB/T 32960.3-2016

### 命令

| 代码 | 说明 | 流向 | 状态 |
|---|---|---|---|
| 0x01 | 车辆登录 | T-BOX → 平台 | OK |
| 0x02 | 实时信息上报 | T-BOX → 平台 | OK |
| 0x03 | 补发信息上报 | T-BOX → 平台 | OK |
| 0x04 | 车辆登出 | T-BOX → 平台 | OK |
| 0x07 | 心跳 | T-BOX → 平台 | OK |
| 0x08 | 终端校时 | T-BOX → 平台 | OK |
| 0xFE | 平台应答 | 平台 → T-BOX | 自动 |

### 数据字段

| 字段 | 标准章节 | 内容 |
|---|---|---|
| 0x01 | 整车数据 | 状态/速度/里程/电压/电流/SOC/档位/绝缘电阻 (20 字节) |
| 0x02 | 驱动电机 | 转速/转矩/温度/电压/电流 |
| 0x03 | 燃料电池 | 电压/电流/氢浓度/氢压力 |
| 0x04 | 发动机 | 状态/转速/燃料消耗率 |
| 0x05 | 位置 | 经度/纬度 |
| 0x06 | 极值 | 电池电压与温度极值 |
| 0x07 | 报警 | 报警等级/内容 |
| 0x08 | 电压 | 子系统/单体电压 |
| 0x09 | 温度 | 子系统/探针温度 |

## 标准合规

| 条目 | GB 32960.3-2016 | 实现 |
|---|---|---|
| 帧结构 | `##` + 命令(1) + 应答(1) + VIN(17) + 加密(1) + 长度(2) + 数据(N) + BCC(1) | OK |
| 起始标识 | `0x23 0x23` | OK |
| 命令单元 | 0x01-0x08 (6 种) | OK |
| 应答标志 | 命令=0x01, 应答=0xFE | OK |
| 加密方式 | 0x01/0x02/0x03/0xFE | OK |
| VIN 码 | 17 字节 ASCII | OK |
| BCC 校验 | 异或 (命令单元→数据单元) | OK |
| 数据字段 | 0x01-0x09 (9 种) | OK |
| 时间格式 | 6 字节 BCD (YYMMDDHHMMSS) | OK |

## 高并发测试

测试环境：Windows x64, Go 1.25, localhost, JSON 日志

| 总连接 | Worker | 成功率 | 吞吐量 | 平均延迟 |
|---|---|---|---|---|
| 500 | 128 | 100% | 3,207 conns/s | 32 ms |
| 2,000 | 128 | 100% | 1,672 conns/s | 73 ms |
| 5,000 | 128 | 100% | 1,118 conns/s | 113 ms |
| 10,000 | 128 | 100% | 1,212 conns/s | 105 ms |

每连接执行：登录 → 实时×2 → 心跳 → 登出（5 次命令交互）。

## API 参考

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

## 开发

```bash
git clone https://github.com/darkinno/gb32960-go-sdk.git
cd gb32960-go-sdk

go build ./...          # 编译
go test ./... -v        # 测试
go vet ./...            # 静态分析

go build -o server ./cmd/example/   # 编译示例
```

## 项目结构

```
gb32960-go-sdk/
├── server.go          # 服务主体
├── options.go         # 函数式配置选项
├── events.go          # Handler 事件接口
├── interfaces.go      # Authenticator / Forwarder 接口
├── decoder.go         # 协议流式解码器（帧/BCC）
├── packet.go          # 数据包与消息类型
├── message.go         # 消息解码
├── conn.go            # 连接管理
├── buffer.go          # 缓冲区池
├── vin_registry.go    # VIN 索引
├── constant/
│   └── command.go     # 协议常量
├── codec/
│   ├── field.go       # 字段分发
│   ├── vehicle.go     # 整车数据（20 字节）
│   ├── motor.go       # 驱动电机
│   ├── battery.go     # 燃料电池
│   ├── engine.go      # 发动机
│   ├── position.go    # 位置
│   ├── extreme.go     # 极值
│   ├── alarm.go       # 报警
│   ├── voltage.go     # 电压
│   └── temperature.go # 温度
├── auth/
│   └── auth.go        # VIN 白名单 / AllowAll
├── forward/
│   └── config.go      # 转发器配置
├── cmd/
│   ├── example/       # 使用示例
│   ├── simclient/     # T-BOX 模拟器
│   └── stresstest/    # 并发压力测试
├── go.mod
├── LICENSE
├── README.md
└── README_zh.md
```

## 许可证

[MIT License](LICENSE) — [DarkInno](https://github.com/darkinno)

---

<p align="center">
  如果你觉得有用，请点亮一颗星<br>
  <a href="https://github.com/darkinno/gb32960-go-sdk">
    <img src="https://img.shields.io/github/stars/darkinno/gb32960-go-sdk?style=social" alt="GitHub stars">
  </a>
</p>
