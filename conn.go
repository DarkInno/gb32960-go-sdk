package gb32960

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/darkinno/gb32960-go-sdk/constant"
)

type ConnectionState int32

const (
	ConnNew ConnectionState = iota
	ConnLoggedIn
	ConnClosed
)

type Connection struct {
	id        string
	conn      net.Conn
	state     atomic.Int32
	vin       string
	iccid     string
	createdAt time.Time
	lastSeen  atomic.Int64

	writeMu   sync.Mutex

	decoder   *Decoder
	server    *Server

	logger    *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
}

func newConnection(id string, netConn net.Conn, server *Server) *Connection {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Connection{
		id:        id,
		conn:      netConn,
		createdAt: time.Now(),
		decoder:   NewDecoder(),
		server:    server,
		logger:    server.logger.With("conn_id", id, "remote", netConn.RemoteAddr().String()),
		ctx:       ctx,
		cancel:    cancel,
	}
	c.state.Store(int32(ConnNew))
	c.lastSeen.Store(time.Now().Unix())
	return c
}

func (c *Connection) ID() string {
	return c.id
}

func (c *Connection) VIN() string {
	return c.vin
}

func (c *Connection) ICCID() string {
	return c.iccid
}

func (c *Connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *Connection) State() ConnectionState {
	return ConnectionState(c.state.Load())
}

func (c *Connection) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Connection) LastSeen() time.Time {
	return time.Unix(c.lastSeen.Load(), 0)
}

func (c *Connection) Send(command byte, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.State() == ConnClosed {
		return net.ErrClosed
	}

	if c.server.writeTimeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.server.writeTimeout))
	}

	pkt, err := EncodeResponse(command, c.vin, constant.EncNone, data)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(pkt)
	return err
}

func (c *Connection) sendLoginResponse(resp *LoginResponse) error {
	data, err := EncodeLoginResponse(resp)
	if err != nil {
		return err
	}
	return c.Send(constant.CmdLogin, data)
}

func (c *Connection) Close() {
	if c.State() == ConnClosed {
		return
	}
	c.state.Store(int32(ConnClosed))
	c.cancel()
	c.decoder.Close()
	c.conn.Close()

	c.server.unregister(c)

	c.logger.Info("connection closed", "vin", c.vin)
}

func (c *Connection) run() {
	defer c.Close()

	bufPtr := getBuffer()
	buf := *bufPtr
	defer putBuffer(bufPtr)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		if c.server.readTimeout > 0 {
			c.conn.SetReadDeadline(time.Now().Add(c.server.readTimeout))
		}

		n, err := c.conn.Read(buf)
		if err != nil {
			if c.State() != ConnClosed {
				c.logger.Debug("read error", "error", err)
			}
			return
		}

		if n > 0 {
			c.lastSeen.Store(time.Now().Unix())
			c.decoder.Feed(buf[:n])
			c.processPackets()
		}
	}
}

func (c *Connection) processPackets() {
	for {
		pkt, err := c.decoder.Decode()
		if err != nil {
			c.logger.Debug("decode error", "error", err)
			return
		}
		if pkt == nil {
			return
		}
		c.handlePacket(pkt)
	}
}

func (c *Connection) handlePacket(pkt *Packet) {
	ctx := c.ctx
	h := c.server.handler

	switch pkt.Command {
	case constant.CmdLogin:
		if c.server.auth != nil {
			ok, err := c.server.auth.Authenticate(ctx, pkt.VIN)
			if err != nil || !ok {
				c.logger.Warn("auth failed", "vin", pkt.VIN, "error", err)
				c.Close()
				return
			}
		}

		loginData, err := DecodeLoginData(pkt.Data)
		if err != nil {
			c.logger.Error("login decode error", "error", err)
			return
		}

		c.vin = pkt.VIN
		c.iccid = loginData.ICCID
		c.state.Store(int32(ConnLoggedIn))

		if c.server.vinRegistry != nil {
			c.server.vinRegistry.add(pkt.VIN, c)
		}

		if h != nil {
			resp, err := h.OnVehicleLogin(ctx, c, loginData)
			if err != nil {
				c.logger.Error("login handler error", "error", err)
				return
			}
			if resp != nil {
				if err := c.sendLoginResponse(resp); err != nil {
					c.logger.Error("login response send error", "error", err)
				}
			}
		}

		c.server.forward(ctx, newForwardMsg("login", pkt.VIN, loginData))
		c.logger.Info("vehicle login", "vin", pkt.VIN, "iccid", loginData.ICCID)

	case constant.CmdLogout:
		logoutData, err := DecodeLogoutData(pkt.Data)
		if err != nil {
			c.logger.Error("logout decode error", "error", err)
			return
		}

		if h != nil {
			if err := h.OnVehicleLogout(ctx, c, logoutData); err != nil {
				c.logger.Error("logout handler error", "error", err)
			}
		}

		c.server.forward(ctx, newForwardMsg("logout", pkt.VIN, logoutData))
		c.logger.Info("vehicle logout", "vin", pkt.VIN)

	case constant.CmdRealtime:
		msg, err := DecodeRealtimeData(pkt.Data)
		if err != nil {
			c.logger.Error("realtime decode error", "error", err)
			return
		}

		if h != nil {
			if err := h.OnRealtimeData(ctx, c, msg); err != nil {
				c.logger.Error("realtime handler error", "error", err)
			}
		}

		c.server.forward(ctx, newForwardMsg("realtime", pkt.VIN, msg))

	case constant.CmdReissue:
		msg, err := DecodeReissueData(pkt.Data)
		if err != nil {
			c.logger.Error("reissue decode error", "error", err)
			return
		}

		if h != nil {
			if err := h.OnReissueData(ctx, c, msg); err != nil {
				c.logger.Error("reissue handler error", "error", err)
			}
		}

		c.server.forward(ctx, newForwardMsg("reissue", pkt.VIN, msg))

	case constant.CmdHeartbeat:
		if h != nil {
			if err := h.OnHeartbeat(ctx, c, &HeartbeatData{}); err != nil {
				c.logger.Error("heartbeat handler error", "error", err)
			}
		}

		c.Send(constant.CmdHeartbeat, nil)

	case constant.CmdTimeCalibr:
		var tp time.Time
		if c.server.timeProvider != nil {
			var err error
			tp, err = c.server.timeProvider.OnTimeCalibration(ctx, c)
			if err != nil {
				c.logger.Error("time calibration error", "error", err)
				tp = time.Now().UTC()
			}
		} else {
			tp = time.Now().UTC()
		}

		data := encodeTime6(tp)
		c.Send(constant.CmdTimeCalibr, data)

	default:
		c.logger.Debug("unknown command", "cmd", pkt.Command)
	}
}

func generateConnID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return hex.EncodeToString(b)
}
