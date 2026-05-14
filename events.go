package gb32960

import (
	"context"
	"time"
)

type Handler interface {
	OnVehicleLogin(ctx context.Context, conn *Connection, msg *VehicleLoginData) (*LoginResponse, error)

	OnVehicleLogout(ctx context.Context, conn *Connection, msg *VehicleLogoutData) error

	OnRealtimeData(ctx context.Context, conn *Connection, msg *RealtimeMessage) error

	OnReissueData(ctx context.Context, conn *Connection, msg *ReissueMessage) error

	OnHeartbeat(ctx context.Context, conn *Connection, msg *HeartbeatData) error
}

type TimeProvider interface {
	OnTimeCalibration(ctx context.Context, conn *Connection) (time.Time, error)
}
