package joystick

const (
	MAX_EFFECTS        = 14
	MAX_FFB_AXIS_COUNT = 0x02
)

type ReportID uint8
type EffectType uint8

const (
	ReportPIDStatusInputData ReportID = 0x02

	ReportSetEffectOutputData              ReportID = 0x01
	ReportSetEnvelopeOutputData            ReportID = 0x02
	ReportSetConditionOutputData           ReportID = 0x03
	ReportSetPeriodicOutputData            ReportID = 0x04
	ReportSetConstantForceOutputData       ReportID = 0x05
	ReportSetRampForceOutputData           ReportID = 0x06
	ReportSetCustomForceDataOutputData     ReportID = 0x07
	ReportSetDownloadForceSampleOutputData ReportID = 0x08
	ReportEffectOperationOutputData        ReportID = 0x0a
	ReportBlockFreeOutputData              ReportID = 0x0b
	ReportDeviceControlOutputData          ReportID = 0x0c
	ReportDeviceGainOutputData             ReportID = 0x0d
	ReportSetCustomForceOutputData         ReportID = 0x0e
	//Report ReportID = 0x08

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

	MEFFECTSTATE_FREE      = 0x00
	MEFFECTSTATE_ALLOCATED = 0x01
	MEFFECTSTATE_PLAYING   = 0x02

	X_AXIS_ENABLE     = 0x01
	Y_AXIS_ENABLE     = 0x02
	DIRECTION_ENABLE  = 0x04
	INERTIA_FORCE     = 0xFF
	FRICTION_FORCE    = 0xFF
	INERTIA_DEADBAND  = 0x30
	FRICTION_DEADBAND = 0x30
)

func TO_LT_END_16(x uint16) uint16 { return ((x << 8) & 0xFF00) | ((x >> 8) & 0x00FF) }

type TEffectState struct {
	State                 uint8
	EffectType            EffectType
	Offset                int16
	Gain                  uint8
	AttackLevel           int16
	FadeLevel             int16
	AttackTime            uint16
	FadeTime              uint16
	Magnitude             int16
	EnableAxis            uint8
	DirectionX            uint8 // angle (0=0 .. 255=360deg)
	DirectionY            uint8 // angle (0=0 .. 255=360deg)
	ConditionBlockCount   uint8
	Conditions            [MAX_FFB_AXIS_COUNT]TEffectCondition
	Phase                 uint16 // 0..255 (=0..359, exp-2)
	StartMagnitude        int16
	EndMagnitude          int16
	Period                uint16 // 0..32767 ms
	Duration, ElapsedTime uint16
	StartTime             uint64
}

type PIDStatusInputData struct {
	ReportId         uint8 //2
	Status           uint8 // Bits: 0=Device Paused,1=Actuators Enabled,2=Safety Switch,3=Actuator Override Switch,4=Actuator Power
	EffectBlockIndex uint8 // Bit7=Effect Playing, Bit0..7=EffectId (1..40)
}

type SetEffectOutputData struct {
	ReportId              uint8  // =1
	EffectBlockIndex      uint8  // 1..40
	EffectType            uint8  // 1..12 (effect usages: 26,27,30,31,32,33,34,40,41,42,43,28)
	Duration              uint16 // 0..32767 ms
	TriggerRepeatInterval uint16 // 0..32767 ms
	SamplePeriod          uint16 // 0..32767 ms
	Gain                  uint8  // 0..255	 (physical 0..10000)
	TriggerButton         uint8  // button ID (0..8)
	EnableAxis            uint8  // bits: 0=X, 1=Y, 2=DirectionEnable
	DirectionX            uint8  // angle (0=0 .. 255=360deg)
	DirectionY            uint8  // angle (0=0 .. 255=360deg)
	//	uint16	startDelay	// 0..32767 ms
}

type SetEnvelopeOutputData struct {
	ReportId         uint8 // =2
	EffectBlockIndex uint8 // 1..40
	AttackLevel      uint16
	FadeLevel        uint16
	AttackTime       uint32 // ms
	FadeTime         uint32 // ms
}

type SetConditionOutputData struct {
	ReportId             uint8  // =3
	EffectBlockIndex     uint8  // 1..40
	ParameterBlockOffset uint8  // bits: 0..3=parameterBlockOffset, 4..5=instance1, 6..7=instance2
	CpOffset             int16  // 0..255
	PositiveCoefficient  int16  // -128..127
	NegativeCoefficient  int16  // -128..127
	PositiveSaturation   uint16 // -	128..127
	NegativeSaturation   uint16 // -128..127
	DeadBand             uint16 // 0..255
}

type SetPeriodicOutputData struct {
	ReportId         uint8 // =4
	EffectBlockIndex uint8 // 1..40
	Magnitude        uint16
	Offset           int16
	Phase            uint16 // 0..255 (=0..359, exp-2)
	Period           uint32 // 0..32767 ms
}

type SetConstantForceOutputData struct {
	ReportId         uint8 // =5
	EffectBlockIndex uint8 // 1..40
	Magnitude        int16 // -255..255
}

type SetRampForceOutputData struct {
	ReportId         uint8 // =6
	EffectBlockIndex uint8 // 1..40
	StartMagnitude   int16
}

type SetCustomForceDataOutputData struct {
	ReportId         uint8 // =7
	EffectBlockIndex uint8 // 1..40
	DataOffset       uint16
	Data             [12]int8
}

type SetDownloadForceSampleOutputData struct {
	ReportId uint8 // =8
	X        int8
	Y        int8
}

type EffectOperationOutputData struct {
	ReportId         uint8 // =10
	EffectBlockIndex uint8 // 1..40
	Operation        uint8 // 1=Start, 2=StartSolo, 3=Stop
	LoopCount        uint8
}

type BlockFreeOutputData struct {
	ReportId         uint8 // =11
	EffectBlockIndex uint8 // 1..40
}

type DeviceControlOutputData struct {
	ReportId uint8 // =12
	Control  uint8 // 1=Enable Actuators, 2=Disable Actuators, 4=Stop All Effects, 8=Reset, 16=Pause, 32=Continue
}

type DeviceGainOutputData struct {
	ReportId uint8 // =13
	Gain     uint8
}

type SetCustomForceOutputData struct {
	ReportId         uint8 // =14
	EffectBlockIndex uint8 // 1..40
	SampleCount      uint8
	SamplePeriod     uint16 // 0..32767 ms
}

type CreateNewEffectFeatureData struct {
	ReportId   uint8  //5
	EffectType uint8  // Enum (1..12): ET 26,27,30,31,32,33,34,40,41,42,43,28
	ByteCount  uint16 // 0..511
}

type PIDBlockLoadFeatureData struct {
	ReportId         uint8  // =6
	EffectBlockIndex uint8  // 1..40
	LoadStatus       uint8  // 1=Success,2=Full,3=Error
	RamPoolAvailable uint16 // =0 or 0xFFFF?
}

type PIDPoolFeatureData struct {
	ReportId               uint8  // =7
	RamPoolSize            uint16 // ?
	MaxSimultaneousEffects uint8  // ?? 40?
	MemoryManagement       uint8  // Bits: 0=DeviceManagedPool, 1=SharedParameterBlocks
}

type TEffectCondition struct {
	CpOffset            int16  // -128..127
	PositiveCoefficient int16  // -128..127
	NegativeCoefficient int16  // -128..127
	PositiveSaturation  uint16 // -128..127
	NegativeSaturation  uint16 // -128..127
	DeadBand            uint16 // 0..255
}
