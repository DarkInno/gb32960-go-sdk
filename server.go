package gb32960

import (
	"context"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/darkinno/gb32960-go-sdk/constant"
)

type Server struct {
	addr         string
	handler      Handler
	auth         Authenticator
	forwarders   []Forwarder
	forwardMu    sync.RWMutex
	timeProvider TimeProvider
	cryptoProvider  CryptoProvider
	platformHandler PlatformHandler
	paramHandler    ParamHandler

	maxConns    int
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration

	listener    net.Listener
	connections sync.Map
	connCount   atomic.Int64
	vinRegistry *vinRegistry

	logger   *slog.Logger
	logLevel slog.LevelVar

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		addr:        ":32960",
		maxConns:    10000,
		readTimeout: 5 * time.Minute,
		writeTimeout: 10 * time.Second,
		idleTimeout: 10 * time.Minute,
		vinRegistry: newVinRegistry(),
	}
	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: &s.logLevel,
	}))
	s.logLevel.Set(slog.LevelInfo)

	for _, o := range opts {
		o(s)
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.logger.Info("server started", "addr", s.addr)

	if s.idleTimeout > 0 {
		s.wg.Add(1)
		go s.idleCheckLoop()
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
				s.logger.Error("accept error", "error", err)
				continue
			}
		}

		if s.connCount.Load() >= int64(s.maxConns) {
			conn.Close()
			s.logger.Warn("connection limit reached", "limit", s.maxConns)
			continue
		}

		c := newConnection(generateConnID(), conn, s)
		s.connections.Store(c.id, c)
		s.connCount.Add(1)
		s.wg.Add(1)

		go func() {
			defer s.wg.Done()
			c.run()
		}()
	}
}

func (s *Server) Stop() {
	s.logger.Info("shutting down server")
	s.cancel()

	if s.listener != nil {
		s.listener.Close()
	}

	s.connections.Range(func(key, value interface{}) bool {
		if c, ok := value.(*Connection); ok {
			c.Close()
		}
		return true
	})

	s.wg.Wait()

	s.forwardMu.RLock()
	for _, f := range s.forwarders {
		f.Close()
	}
	s.forwardMu.RUnlock()

	s.logger.Info("server stopped")
}

func (s *Server) ConnCount() int64 {
	return s.connCount.Load()
}

func (s *Server) Connections() []*Connection {
	list := make([]*Connection, 0)
	s.connections.Range(func(key, value interface{}) bool {
		if c, ok := value.(*Connection); ok {
			list = append(list, c)
		}
		return true
	})
	return list
}

func (s *Server) GetConnectionByVIN(vin string) *Connection {
	return s.vinRegistry.get(vin)
}

func (s *Server) GetConnection(id string) *Connection {
	v, ok := s.connections.Load(id)
	if !ok {
		return nil
	}
	return v.(*Connection)
}

func (s *Server) SetLogLevel(level slog.Level) {
	s.logLevel.Set(level)
}

func (s *Server) unregister(c *Connection) {
	s.connections.Delete(c.id)
	s.connCount.Add(-1)
	if c.vin != "" {
		s.vinRegistry.remove(c.vin, c)
	}
}

func (s *Server) forward(ctx context.Context, msg interface{}) {
	s.forwardMu.RLock()
	defer s.forwardMu.RUnlock()
	for _, f := range s.forwarders {
		go func(fw Forwarder) {
			if err := fw.Forward(ctx, msg); err != nil {
				s.logger.Error("forward error", "error", err)
			}
		}(f)
	}
}

func (s *Server) idleCheckLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
		}

		s.connections.Range(func(key, value interface{}) bool {
			c, ok := value.(*Connection)
			if !ok || c.State() == ConnClosed {
				return true
			}

			if time.Since(c.LastSeen()) <= s.idleTimeout {
				return true
			}

			if err := c.Send(constant.CmdHeartbeat, nil); err != nil {
				s.logger.Warn("idle probe failed, closing", "conn_id", c.id, "error", err)
				c.Close()
				return true
			}

			time.Sleep(3 * time.Second)

			if c.State() != ConnClosed && time.Since(c.LastSeen()) > s.idleTimeout+3*time.Second {
				s.logger.Warn("idle connection closing", "conn_id", c.id, "last_seen", c.LastSeen())
				c.Close()
			}

			return true
		})
	}
}
