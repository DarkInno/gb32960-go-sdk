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

type PlatformHandler interface {
	OnPlatLoginResponse(ctx context.Context, conn *Connection, msg *PlatformLoginData) error
	OnPlatLogoutResponse(ctx context.Context, conn *Connection, msg *PlatformLogoutData) error
}

type ParamHandler interface {
	OnParamQuery(ctx context.Context, conn *Connection, msg *ParamQueryData) (*ParamQueryResponse, error)
	OnParamQueryAck(ctx context.Context, conn *Connection, msg *ParamQueryData) error
	OnParamSettingAck(ctx context.Context, conn *Connection, msg *ParamSettingData) error
}
