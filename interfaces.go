package gb32960

import "context"

type Authenticator interface {
	Authenticate(ctx context.Context, vin string) (bool, error)
}

type Forwarder interface {
	Forward(ctx context.Context, msg interface{}) error
	Close() error
}
