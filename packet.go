package gb32960

import "time"

type Packet struct {
	Command     byte
	Response    byte
	VIN         string
	EncryptType byte
	Data        []byte
	Length      uint16
}

type ForwardMsg struct {
	Type string      `json:"type"`
	VIN  string      `json:"vin"`
	Data interface{} `json:"data"`
}

func newForwardMsg(msgType, vin string, data interface{}) *ForwardMsg {
	return &ForwardMsg{Type: msgType, VIN: vin, Data: data}
}

type VehicleLoginData struct {
	LoginTime      time.Time
	Sequence       uint16
	ICCID          string
	ConfigDataCnt  byte
	ConfigData     []ConfigField
}

type ConfigField struct {
	ID     byte
	Length byte
	Value  []byte
}

type VehicleLogoutData struct {
	LogoutTime time.Time
	Sequence   uint16
}

type HeartbeatData struct{}

type TimeCalibrationData struct{}

type LoginResponse struct {
	LoginTime time.Time
	Sequence  uint16
	Result    byte
	Token     []byte
}

type VehicleBaseInfo struct {
	VehicleStatus  byte
	ChargingStatus byte
	RunMode        byte
	Speed          uint16
	Odometer       uint32
	TotalVoltage   uint16
	TotalCurrent   uint16
	SOC            byte
	DCStatus       byte
	Gear           byte
	InsulationRes  uint16
}

type MotorInfo struct {
	MotorSeq        byte
	MotorStatus     byte
	ControllerTemp  byte
	MotorSpeed      uint16
	MotorTorque     uint16
	MotorTemp       byte
	MotorVoltage    uint16
	MotorCurrent    uint16
}

type FuelCellInfo struct {
	CellVoltage     uint16
	CellCurrent     uint16
	FuelConsumption uint16
	ProbeCount      uint16
	ProbeTemps      []byte
	H2MaxTemp       uint16
	H2MaxTempProbe  byte
	H2MaxConcentration uint16
	H2MaxConcentrationSensor byte
	H2PressureMax   uint16
	H2PressureMaxSensor     byte
	H2PressureMin   uint16
	H2PressureMinSensor     byte
	DCDCStatus      byte
}

type EngineInfo struct {
	EngineStatus    byte
	CrankSpeed      uint16
	FuelRate        uint16
}

type PositionInfo struct {
	Longitude  uint32
	Latitude   uint32
}

type ExtremeInfo struct {
	MaxBatteryVoltage      uint16
	MaxBatteryVoltageCode  byte
	MinBatteryVoltage      uint16
	MinBatteryVoltageCode  byte
	MaxTemp                byte
	MaxTempCode            byte
	MinTemp                byte
	MinTempCode            byte
}

type AlarmInfo struct {
	MaxLevel            byte
	AlarmByteLen        uint32
	AlarmBytes          []byte
}

type VoltageInfo struct {
	SubsysCount     uint16
	SubsystemVoltages []SubsystemVoltage
}

type SubsystemVoltage struct {
	SubsysNo        byte
	Voltage         uint16
	Current         uint16
	CellCount       uint16
	Cells           []CellVoltage
}

type CellVoltage struct {
	CellNo   byte
	CellInfo CellInfo
}

type CellInfo struct {
	Voltage   uint16
	Temp      byte
}

type RealtimeMessage struct {
	InfoTime     time.Time

	VehicleData  *VehicleBaseInfo
	MotorData    []MotorInfo
	FuelCellData *FuelCellInfo
	EngineData   *EngineInfo
	PositionData *PositionInfo
	ExtremeData  *ExtremeInfo
	AlarmData    *AlarmInfo
	VoltageData  *VoltageInfo
}

type ReissueMessage struct {
	InfoTime     time.Time
	DataItems    []RealtimeMessage
}

type PlatformLoginData struct {
	LoginTime time.Time
	Sequence  uint16
	Username  string
	Password  string
}

type PlatformLogoutData struct {
	LogoutTime time.Time
	Sequence   uint16
}

type ParamQueryData struct {
	QueryTime time.Time
	Count     byte
	ParamIDs  []uint32
}

type ParamQueryResponse struct {
	Count   byte
	Params  []ParamItem
}

type ParamItem struct {
	ID    uint32
	Value []byte
}

type ParamSettingData struct {
	SettingTime time.Time
	Count       byte
	Params      []ParamItem
}

