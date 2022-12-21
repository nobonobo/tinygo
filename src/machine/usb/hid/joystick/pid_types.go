package joystick

import "encoding/binary"

const (
	MAX_EFFECTS        = 16
	MAX_FFB_AXIS_COUNT = 0x02
)

type ReportID uint8
type ControlType uint8
type EffectType uint8
type EffectState uint8
type EffectOperation uint8

const (
	ReportPIDStatusInputData ReportID = 0x02

	ReportSetEffect              ReportID = 0x01
	ReportSetEnvelope            ReportID = 0x02
	ReportSetCondition           ReportID = 0x03
	ReportSetPeriodic            ReportID = 0x04
	ReportSetConstantForce       ReportID = 0x05
	ReportSetRampForce           ReportID = 0x06
	ReportSetCustomForceData     ReportID = 0x07
	ReportSetDownloadForceSample ReportID = 0x08
	ReportEffectOperation        ReportID = 0x0a
	ReportBlockFree              ReportID = 0x0b
	ReportDeviceControl          ReportID = 0x0c
	ReportDeviceGain             ReportID = 0x0d
	ReportSetCustomForce         ReportID = 0x0e
	//Report ReportID = 0x08

	ControlEnableActuators  ControlType = 0x01
	ControlDisableActuators ControlType = 0x02
	ControlStopAllEffects   ControlType = 0x03
	ControlReset            ControlType = 0x04
	ControlPause            ControlType = 0x05
	ControlContinue         ControlType = 0x06

	USB_EFFECT_CONSTANT     EffectType = 0x01
	USB_EFFECT_RAMP         EffectType = 0x02
	USB_EFFECT_SQUARE       EffectType = 0x03
	USB_EFFECT_SINE         EffectType = 0x04
	USB_EFFECT_TRIANGLE     EffectType = 0x05
	USB_EFFECT_SAWTOOTHDOWN EffectType = 0x06
	USB_EFFECT_SAWTOOTHUP   EffectType = 0x07
	USB_EFFECT_SPRING       EffectType = 0x08
	USB_EFFECT_DAMPER       EffectType = 0x09
	USB_EFFECT_INERTIA      EffectType = 0x0A
	USB_EFFECT_FRICTION     EffectType = 0x0B
	USB_EFFECT_CUSTOM       EffectType = 0x0C

	MEFFECTSTATE_FREE      EffectState = 0x00
	MEFFECTSTATE_ALLOCATED EffectState = 0x01
	MEFFECTSTATE_PLAYING   EffectState = 0x02

	EOStart     EffectOperation = 1
	EOStartSolo EffectOperation = 2
	EOStop      EffectOperation = 3

	X_AXIS_ENABLE     = 0x01
	Y_AXIS_ENABLE     = 0x02
	DIRECTION_ENABLE  = 0x04
	INERTIA_FORCE     = 0xFF
	FRICTION_FORCE    = 0xFF
	INERTIA_DEADBAND  = 0x30
	FRICTION_DEADBAND = 0x30
)

func TO_LT_END_16(x uint16) uint16 { return ((x << 8) & 0xFF00) | ((x >> 8) & 0x00FF) }

type PIDStatusInputData struct {
	ReportID         ReportID //2
	Status           uint8    // Bits: 0=Device Paused,1=Actuators Enabled,2=Safety Switch,3=Actuator Override Switch,4=Actuator Power
	EffectBlockIndex uint8    // Bit7=Effect Playing, Bit0..7=EffectId (1..40)
}

type SetEffectOutputData struct {
	ReportID              ReportID // =1
	EffectBlockIndex      uint8    // 1..40
	EffectType            uint8    // 1..12 (effect usages: 26,27,30,31,32,33,34,40,41,42,43,28)
	Duration              uint16   // 0..32767 ms
	TriggerRepeatInterval uint16   // 0..32767 ms
	SamplePeriod          uint16   // 0..32767 ms
	Gain                  uint8    // 0..255	 (physical 0..10000)
	TriggerButton         uint8    // button ID (0..8)
	EnableAxis            uint8    // bits: 0=X, 1=Y, 2=DirectionEnable
	DirectionX            uint8    // angle (0=0 .. 255=360deg)
	DirectionY            uint8    // angle (0=0 .. 255=360deg)
	StartDelay            uint16   // 0..32767 ms
}

func (s *SetEffectOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[0]
	s.EffectType = b[0]
	s.Duration = binary.LittleEndian.Uint16(b[3:5])
	s.TriggerRepeatInterval = binary.LittleEndian.Uint16(b[5:7])
	s.SamplePeriod = binary.LittleEndian.Uint16(b[7:9])
	s.Gain = b[9]
	s.TriggerButton = b[10]
	s.EnableAxis = b[11]
	s.DirectionX = b[12]
	s.DirectionY = b[13]
	s.StartDelay = binary.LittleEndian.Uint16(b[14:16])
	return nil
}

type SetEnvelopeOutputData struct {
	ReportID         ReportID // =2
	EffectBlockIndex uint8    // 1..40
	AttackLevel      uint16
	FadeLevel        uint16
	AttackTime       uint32 // ms
	FadeTime         uint32 // ms
}

func (s *SetEnvelopeOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.AttackLevel = binary.LittleEndian.Uint16(b[2:4])
	s.FadeLevel = binary.LittleEndian.Uint16(b[4:6])
	s.AttackTime = binary.LittleEndian.Uint32(b[6:10])
	s.FadeTime = binary.LittleEndian.Uint32(b[10:14])
	return nil
}

type SetConditionOutputData struct {
	ReportID             ReportID // =3
	EffectBlockIndex     uint8    // 1..40
	ParameterBlockOffset uint8    // bits: 0..3=parameterBlockOffset, 4..5=instance1, 6..7=instance2
	CpOffset             int16    // 0..255
	PositiveCoefficient  int16    // -128..127
	NegativeCoefficient  int16    // -128..127
	PositiveSaturation   uint16   // -	128..127
	NegativeSaturation   uint16   // -128..127
	DeadBand             uint16   // 0..255
}

func (s *SetConditionOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	return nil
}

type SetPeriodicOutputData struct {
	ReportID         ReportID // =4
	EffectBlockIndex uint8    // 1..40
	Magnitude        uint16
	Offset           int16
	Phase            uint16 // 0..255 (=0..359, exp-2)
	Period           uint32 // 0..32767 ms
}

func (s *SetPeriodicOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	return nil
}

type SetConstantForceOutputData struct {
	ReportID         ReportID // =5
	EffectBlockIndex uint8    // 1..40
	Magnitude        int16    // -255..255
}

func (s *SetConstantForceOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	return nil
}

type SetRampForceOutputData struct {
	ReportID         ReportID // =6
	EffectBlockIndex uint8    // 1..40
	StartMagnitude   int16
}

func (s *SetRampForceOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	return nil
}

type SetCustomForceDataOutputData struct {
	ReportID         ReportID // =7
	EffectBlockIndex uint8    // 1..40
	DataOffset       uint16
	Data             [12]int8
}

func (s *SetCustomForceDataOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	return nil
}

type SetDownloadForceSampleOutputData struct {
	ReportID ReportID // =8
	X        int8
	Y        int8
}

func (s *SetDownloadForceSampleOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.X = int8(b[1])
	s.Y = int8(b[2])
	return nil
}

type EffectOperationOutputData struct {
	ReportID         ReportID        // =10
	EffectBlockIndex uint8           // 1..40
	Operation        EffectOperation // 1=Start, 2=StartSolo, 3=Stop
	LoopCount        uint8
}

func (s *EffectOperationOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.Operation = EffectOperation(b[2])
	s.LoopCount = b[3]
	return nil
}

type BlockFreeOutputData struct {
	ReportID         ReportID // =11
	EffectBlockIndex uint8    // 1..40
}

func (s *BlockFreeOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	return nil
}

type DeviceControlOutputData struct {
	ReportID ReportID // =12
	// 1=Enable Actuators, 2=Disable Actuators, 3=Stop All Effects, 4=Reset, 5=Pause, 6=Continue
	// 1=Enable Actuators, 2=Disable Actuators, 4=Stop All Effects, 8=Reset, 16=Pause, 32=Continue
	Control ControlType
}

func (s *DeviceControlOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.Control = ControlType(b[1])
	return nil
}

type DeviceGainOutputData struct {
	ReportID ReportID // =13
	Gain     uint8
}

func (s *DeviceGainOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.Gain = b[1]
	return nil
}

type SetCustomForceOutputData struct {
	ReportID         ReportID // =14
	EffectBlockIndex uint8    // 1..40
	SampleCount      uint8
	SamplePeriod     uint16 // 0..32767 ms
}

func (s *SetCustomForceOutputData) UnmarshalBinary(b []byte) error {
	s.ReportID = ReportID(b[0])
	s.EffectBlockIndex = b[1]
	s.SampleCount = b[2]
	return nil
}

type CreateNewEffectFeatureData struct {
	ReportID   ReportID   //5
	EffectType EffectType // Enum (1..12): ET 26,27,30,31,32,33,34,40,41,42,43,28
	ByteCount  uint16     // 0..511
}

type PIDBlockLoadFeatureData struct {
	ReportID         ReportID // =6
	EffectBlockIndex uint8    // 1..40
	LoadStatus       uint8    // 1=Success,2=Full,3=Error
	RamPoolAvailable uint16   // =0 or 0xFFFF?
}

type PIDPoolFeatureData struct {
	ReportID               ReportID // =7
	RamPoolSize            uint16   // ?
	MaxSimultaneousEffects uint8    // ?? 40?
	MemoryManagement       uint8    // Bits: 0=DeviceManagedPool, 1=SharedParameterBlocks
}

type TEffectCondition struct {
	CpOffset            int16  // -128..127
	PositiveCoefficient int16  // -128..127
	NegativeCoefficient int16  // -128..127
	PositiveSaturation  uint16 // -128..127
	NegativeSaturation  uint16 // -128..127
	DeadBand            uint16 // 0..255
}

type TEffectState struct {
	State      EffectState // see constants <MEffectState_*>
	EffectType EffectType
	Offset     int16
	Gain       uint8
	// envelope
	AttackLevel int16
	FadeLevel   int16
	FadeTime    uint16
	AttackTime  uint16

	Magnitude int16
	// direction
	EnableAxis uint8 // bits: 0=X, 1=Y, 2=DirectionEnable
	DirectionX uint8 // angle (0=0 .. 255=360deg)
	DirectionY uint8 // angle (0=0 .. 255=360deg)
	// condition
	ConditionBlocksCount uint8
	Conditions           [MAX_FFB_AXIS_COUNT]TEffectCondition
	// periodic
	Phase          uint16 // 0..255 (=0..359, exp-2)
	StartMagnitude int16
	EndMagnitude   int16
	Period         uint16 // 0..32767 ms
	Duration       uint16
	ElapsedTime    uint16
	StartTime      uint64
}
