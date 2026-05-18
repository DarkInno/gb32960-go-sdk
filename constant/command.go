package constant

const (
	StartMarker1 = 0x23
	StartMarker2 = 0x23

	CmdLogin      = 0x01
	CmdRealtime   = 0x02
	CmdReissue    = 0x03
	CmdLogout     = 0x04
	CmdPlatLogin  = 0x05
	CmdPlatLogout = 0x06
	CmdHeartbeat  = 0x07
	CmdTimeCalibr = 0x08

	CmdParamQuery     = 0x80
	CmdParamQueryResp = 0x81
	CmdParamSetting   = 0x82

	RespSuccess = 0xFE

	EncNone   = 0x01
	EncRSA    = 0x02
	EncAES128 = 0x03
	EncAbnormal = 0xFE

	VINLength = 17

	LoginSuccess   = 0x01
	LoginCarNotFound    = 0x02
	LoginTerminalNotFound = 0x03
	LoginResultAbnormal = 0xFE

	HeaderSize = 24
	MinPacketSize = 25
	MaxPacketSize = 65535 + 25
)

var CommandNames = map[byte]string{
	CmdLogin:          "LOGIN",
	CmdRealtime:       "REALTIME",
	CmdReissue:        "REISSUE",
	CmdLogout:         "LOGOUT",
	CmdPlatLogin:      "PLAT_LOGIN",
	CmdPlatLogout:     "PLAT_LOGOUT",
	CmdHeartbeat:      "HEARTBEAT",
	CmdTimeCalibr:     "TIME_CALIBRATION",
	CmdParamQuery:     "PARAM_QUERY",
	CmdParamQueryResp: "PARAM_QUERY_RESP",
	CmdParamSetting:   "PARAM_SETTING",
}

var LoginResultNames = map[byte]string{
	LoginSuccess:    "SUCCESS",
	LoginCarNotFound:     "CAR_NOT_FOUND",
	LoginTerminalNotFound: "TERMINAL_NOT_FOUND",
	LoginResultAbnormal:  "ABNORMAL",
}
