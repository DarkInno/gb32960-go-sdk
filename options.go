package gb32960

import "time"

type Option func(*Server)

func WithListenAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

func WithMaxConnections(n int) Option {
	return func(s *Server) {
		s.maxConns = n
	}
}

func WithReadTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = d
	}
}

func WithWriteTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = d
	}
}

func WithIdleTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.idleTimeout = d
	}
}

func WithHandler(h Handler) Option {
	return func(s *Server) {
		s.handler = h
	}
}

func WithAuthenticator(a Authenticator) Option {
	return func(s *Server) {
		s.auth = a
	}
}

func WithForwarder(f ...Forwarder) Option {
	return func(s *Server) {
		s.forwardMu.Lock()
		defer s.forwardMu.Unlock()
		s.forwarders = append(s.forwarders, f...)
	}
}

func WithTimeProvider(tp TimeProvider) Option {
	return func(s *Server) {
		s.timeProvider = tp
	}
}
