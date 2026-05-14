package gb32960

import (
	"context"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	addr         string
	handler      Handler
	auth         Authenticator
	forwarders   []Forwarder
	forwardMu    sync.RWMutex
	timeProvider TimeProvider

	maxConns    int
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration

	listener    net.Listener
	connections sync.Map
	connCount   atomic.Int64
	vinRegistry *vinRegistry

	logger *slog.Logger
	logMu  sync.RWMutex

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
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
		vinRegistry: newVinRegistry(),
	}

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
	s.logMu.Lock()
	defer s.logMu.Unlock()
	s.logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
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
				s.logMu.RLock()
				logger := s.logger
				s.logMu.RUnlock()
				logger.Error("forward error", "error", err)
			}
		}(f)
	}
}
