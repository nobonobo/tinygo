package joystick

import (
	"fmt"
	"machine"
	"machine/debug"
	"machine/usb"
	"machine/usb/hid"
	"time"
)

const (
	jsEndpointOut = usb.HID_ENDPOINT_OUT // from PC
	jsEndpointIn  = usb.HID_ENDPOINT_IN  // to PC
)

type PIDHandler struct {
	buf          *RingBuffer
	waitTxc      bool
	effectStates [MAX_EFFECTS]*TEffectState
	nextEID      uint8
	enabled      bool
	paused       bool
	gain         uint8
	pidBlockLoad PIDBlockLoadFeatureData
}

func NewPIDHandler() *PIDHandler {
	effects := [MAX_EFFECTS]*TEffectState{}
	for i := range effects[:] {
		effects[i] = &TEffectState{}
	}
	return &PIDHandler{
		buf:          NewRingBuffer(),
		effectStates: effects,
	}
}

// sendUSBPacket sends a JoystickPacket.
func (m *PIDHandler) sendUSBPacket(b []byte) {
	machine.SendUSBInPacket(jsEndpointIn, b)
}

// from InterruptIn
func (m *PIDHandler) Handler() bool {
	m.waitTxc = false
	if b, ok := m.buf.Get(); ok {
		m.waitTxc = true
		m.sendUSBPacket(b)
		return true
	}
	return false
}

func (m *PIDHandler) tx(b []byte) {
	if m.waitTxc {
		m.buf.Put(b)
	} else {
		m.waitTxc = true
		m.sendUSBPacket(b)
	}
}

// to InterruptOut
func (m *PIDHandler) SendReport(reportID byte, b []byte) {
	m.tx(append([]byte{reportID}, b...))
}

func (m *PIDHandler) SendState(state State) {
	b, _ := state.MarshalBinary()
	m.SendReport(1, b)
}

// from InterruptOut
func (m *PIDHandler) RxHandler(b []byte) {
	if len(b) == 0 {
		return
	}
	reportId := ReportID(b[0])
	switch reportId {
	case ReportSetEffect: // 0x01
		m.SetEffect(b)
	case ReportSetEnvelope: // 0x02
		m.SetEnvelope(b)
	case ReportSetCondition: // 0x03
		m.SetCondition(b)
	case ReportSetPeriodic: // 0x04
		m.SetPeriodic(b)
	case ReportSetConstantForce: // 0x05
		m.SetConstantForce(b)
	case ReportSetRampForce: // 0x06
		m.SetRampForce(b)
	case ReportSetCustomForceData: // 0x07
		m.SetCustomForceData(b)
	case ReportSetDownloadForceSample: // 0x08
		m.SetDownloadForceSample(b)
	case ReportEffectOperation: // 0x0a
		m.EffectOperation(b)
	case ReportBlockFree: // 0x0b
		m.BlockFree(b)
	case ReportDeviceControl: // 0x0c
		m.DeviceControl(b)
	case ReportDeviceGain: // 0x0d
		m.DeviceGain(b)
	case ReportSetCustomForce: // 0x0e
		m.SetCustomForce(b)
	default:
		debug.Debug("unknown rx", debug.Hex(b))
		return
	}
	// debug.Debug("rx", debug.Hex(b))
}

func (m *PIDHandler) CreateNewEffect(data *CreateNewEffectFeatureData) error {
	m.pidBlockLoad.ReportID = 6
	m.pidBlockLoad.EffectBlockIndex = m.GetNextFreeEffect()
	if m.pidBlockLoad.EffectBlockIndex == 0 {
		m.pidBlockLoad.LoadStatus = 2 // 1=Success,2=Full,3=Error
		return fmt.Errorf("effect not allocated")
	}
	m.pidBlockLoad.LoadStatus = 1 // 1=Success,2=Full,3=Error
	effect := TEffectState{}
	effect.State = MEFFECTSTATE_ALLOCATED
	m.pidBlockLoad.RamPoolAvailable -= SIZE_EFFECT
	*m.effectStates[m.pidBlockLoad.EffectBlockIndex] = effect
	return nil
}

func (m *PIDHandler) GetReport(setup usb.Setup) bool {
	reportId := setup.WValueL
	switch setup.WValueH {
	case hid.REPORT_TYPE_INPUT:
	case hid.REPORT_TYPE_OUTPUT:
	case hid.REPORT_TYPE_FEATURE:
		switch reportId {
		case 6:
			b, _ := m.pidBlockLoad.MarshalBinary()
			debug.Debug("GetReport 6 response", debug.Hex(b))
			machine.SendUSBInPacket(0, b)
			return true
		case 7:
			b, _ := PIDPoolFeatureData{
				ReportID:               7,
				RamPoolSize:            MEMORY_SIZE,
				MaxSimultaneousEffects: MAX_EFFECTS,
				MemoryManagement:       3,
			}.MarshalBinary()
			debug.Debug("GetReport 7 response", debug.Hex(b))
			machine.SendUSBInPacket(0, b)
			return true
		}
	}
	debug.Debug("GetReport unknown setup:", setup)
	return false
}

func (m *PIDHandler) GetIdle(setup usb.Setup) bool {
	machine.SendUSBInPacket(0, []byte{0})
	return true
}

func (m *PIDHandler) GetProtocol(setup usb.Setup) bool {
	machine.SendUSBInPacket(0, []byte{0})
	return true
}

func (m *PIDHandler) SetReport(setup usb.Setup) bool {
	reportId := setup.WValueL
	switch setup.WValueH {
	case hid.REPORT_TYPE_INPUT:
		machine.SendZlp()
		return true
	case hid.REPORT_TYPE_OUTPUT:
		machine.SendZlp()
		return true
	case hid.REPORT_TYPE_FEATURE:
		if setup.WLength == 0 {
			machine.ReceiveUSBControlPacket()
			machine.SendZlp()
			return true
		}
		if reportId == 5 {
			b, err := machine.ReceiveUSBControlPacket()
			v := &CreateNewEffectFeatureData{}
			v.UnmarshalBinary(b[:])
			if err := m.CreateNewEffect(v); err != nil {
				debug.Debug("SetReport", err)
			}
			debug.Debug("SetReport", debug.Hex(b[:]), err)
			machine.SendZlp()
			return true
		}
	}
	debug.Debug("SetReport unknown setup:", setup)
	return false
}

func (m *PIDHandler) SetIdle(setup usb.Setup) bool {
	machine.SendZlp()
	return true
}

func (m *PIDHandler) SetProtocol(setup usb.Setup) bool {
	machine.SendZlp()
	return true
}

func (m *PIDHandler) SetupHandler(setup usb.Setup) (ok bool) {
	switch setup.BmRequestType {
	case usb.REQUEST_DEVICETOHOST_CLASS_INTERFACE:
		switch setup.BRequest {
		case usb.GET_REPORT:
			return m.GetReport(setup)
		case usb.GET_IDLE:
			return m.GetIdle(setup)
		case usb.GET_PROTOCOL:
			return m.GetProtocol(setup)
		default:
			debug.Debug("setup: D2H unknown request")
		}
	case usb.REQUEST_HOSTTODEVICE_CLASS_INTERFACE:
		switch setup.BRequest {
		case usb.SET_REPORT:
			return m.SetReport(setup)
		case usb.SET_IDLE:
			return m.SetIdle(setup)
		case usb.SET_PROTOCOL:
			return m.SetProtocol(setup)
		default:
			debug.Debug("setup: H2D unknown request")
		}
	}
	return false
}

func (m *PIDHandler) GetNextFreeEffect() uint8 {
	if m.nextEID == MAX_EFFECTS {
		return 0
	}
	id := m.nextEID
	m.nextEID++

	for m.effectStates[m.nextEID].State != MEFFECTSTATE_FREE {
		if m.nextEID >= MAX_EFFECTS {
			break
		}
		m.nextEID++
	}
	effect := m.effectStates[id]
	effect.State = MEFFECTSTATE_ALLOCATED
	return id
}

func (m *PIDHandler) StopAllEffects() {
	//debug.Debug("StopAllEffects")
	for id := uint8(0); id < MAX_EFFECTS; id++ {
		m.StopEffect(id)
	}
}

func (m *PIDHandler) StartEffect(id uint8) {
	if id >= MAX_EFFECTS {
		debug.Debug("StartEffect", "effect index out of range", id)
		return
	}
	//debug.Debug("StartEffect", id)
	effect := m.effectStates[id]
	effect.State = MEFFECTSTATE_PLAYING
	effect.ElapsedTime = 0
	effect.StartTime = uint64(time.Now().UnixMilli())
}

func (m *PIDHandler) StopEffect(id uint8) {
	if id >= MAX_EFFECTS {
		debug.Debug("StopEffect", "effect index out of range", id)
		return
	}
	//debug.Debug("StopEffect", id)
	effect := m.effectStates[id]
	effect.State &= ^MEFFECTSTATE_PLAYING
	m.pidBlockLoad.RamPoolAvailable += SIZE_EFFECT
}

func (m *PIDHandler) FreeAllEffects() {
	//debug.Debug("FreeAllEffects")
	m.nextEID = 1
	for id := uint8(0); id < MAX_EFFECTS; id++ {
		*m.effectStates[id] = TEffectState{}
	}
	m.pidBlockLoad.RamPoolAvailable = MEMORY_SIZE
}

func (m *PIDHandler) FreeEffect(id uint8) {
	if id >= MAX_EFFECTS {
		debug.Debug("FreeEffect", "effect index out of range", id)
		return
	}
	//debug.Debug("FreeEffect", id)
	state := m.effectStates[id]
	state.State = MEFFECTSTATE_FREE
	if id < m.nextEID {
		m.nextEID = id
	}
}

// SetEffect reportId == 0x01
func (m *PIDHandler) SetEffect(b []byte) {
	//debug.Debug("SetEffect", debug.Hex(b))
	var v SetEffectOutputData
	_ = v.UnmarshalBinary(b)
	effect := m.effectStates[v.EffectBlockIndex]
	effect.Duration = v.Duration
	effect.DirectionX = v.DirectionX
	effect.DirectionY = v.DirectionY
	effect.EffectType = v.EffectType
	effect.Gain = v.Gain
	effect.EnableAxis = v.EnableAxis
}

// SetEnvelope reportId == 0x02
func (m *PIDHandler) SetEnvelope(b []byte) {
	debug.Debug("SetEnvelope", debug.Hex(b))
	var v SetEnvelopeOutputData
	_ = v.UnmarshalBinary(b)
	effect := m.effectStates[v.EffectBlockIndex]
	effect.AttackLevel = int16(v.AttackLevel)
	effect.FadeLevel = v.FadeLevel
	effect.AttackTime = uint16(v.AttackTime)
	effect.FadeTime = uint16(v.FadeTime)
}

// SetCondition reportId == 0x03
func (m *PIDHandler) SetCondition(b []byte) {
	//debug.Debug("SetCondition", debug.Hex(b))
	var v SetConditionOutputData
	_ = v.UnmarshalBinary(b)
	axis := v.ParameterBlockOffset
	effect := m.effectStates[v.EffectBlockIndex]
	condition := effect.Conditions[axis]
	condition.CpOffset = v.CpOffset
	condition.PositiveCoefficient = v.PositiveCoefficient
	condition.NegativeCoefficient = v.NegativeCoefficient
	condition.PositiveSaturation = v.PositiveSaturation
	condition.NegativeSaturation = v.NegativeSaturation
	condition.DeadBand = v.DeadBand
	effect.Conditions[axis] = condition
	if effect.ConditionBlocksCount < axis {
		effect.ConditionBlocksCount++
	}
}

// SetPeriodic reportId == 0x04
func (m *PIDHandler) SetPeriodic(b []byte) {
	//debug.Debug("SetPeriodic", debug.Hex(b))
	var v SetPeriodicOutputData
	_ = v.UnmarshalBinary(b)
	effect := m.effectStates[v.EffectBlockIndex]
	effect.Magnitude = v.Magnitude
	effect.Offset = v.Offset
	effect.Phase = v.Phase
	effect.Period = uint16(v.Period)
}

// SetConstantForce reportId == 0x05
func (m *PIDHandler) SetConstantForce(b []byte) {
	//debug.Debug("SetConstantForce", debug.Hex(b))
	var v SetConstantForceOutputData
	_ = v.UnmarshalBinary(b)
	effect := m.effectStates[v.EffectBlockIndex]
	effect.Magnitude = v.Magnitude
}

// SetRampForce reportId == 0x06
func (m *PIDHandler) SetRampForce(b []byte) {
	debug.Debug("SetRampForce", debug.Hex(b))
	var v SetRampForceOutputData
	_ = v.UnmarshalBinary(b)
	effect := m.effectStates[v.EffectBlockIndex]
	effect.StartMagnitude = v.StartMagnitude
	effect.EndMagnitude = v.EndMagnitude
}

// SetCustomForceData reportId == 0x07
func (m *PIDHandler) SetCustomForceData(b []byte) {
	debug.Debug("SetCustomForceData", debug.Hex(b))
	var v SetCustomForceDataOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// SetDownloadForceSample reportId == 0x08
func (m *PIDHandler) SetDownloadForceSample(b []byte) {
	debug.Debug("SetDownloadForceSample", debug.Hex(b))
	var v SetDownloadForceSampleOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// EffectOperation reportId == 0x0a
func (m *PIDHandler) EffectOperation(b []byte) {
	var v EffectOperationOutputData
	_ = v.UnmarshalBinary(b)
	switch v.Operation {
	case EOStart:
		effect := m.effectStates[v.EffectBlockIndex]
		switch v.LoopCount {
		case 0xff:
			effect.Duration = USB_DURATION_INFINITE
		default:
			effect.Duration *= uint16(v.LoopCount)
		}
		m.effectStates[v.EffectBlockIndex] = effect
		m.StartEffect(v.EffectBlockIndex)
	case EOStartSolo:
		m.StopAllEffects()
		m.StartEffect(v.EffectBlockIndex)
	case EOStop:
		m.StopEffect(v.EffectBlockIndex)
	default:
		debug.Debug("EffectOperation", "unknown operation", v.Operation)
	}
}

// BlockFree reportId == 0x0b
func (m *PIDHandler) BlockFree(b []byte) {
	var v BlockFreeOutputData
	_ = v.UnmarshalBinary(b)
	if v.EffectBlockIndex == 0xff {
		m.FreeAllEffects()
		return
	}
	m.FreeEffect(v.EffectBlockIndex)
}

// DeviceControl reportId == 0x0c
func (m *PIDHandler) DeviceControl(b []byte) {
	var v DeviceControlOutputData
	_ = v.UnmarshalBinary(b)
	switch v.Control {
	case ControlEnableActuators:
		m.enabled = true
	case ControlDisableActuators:
		m.enabled = false
	case ControlStopAllEffects:
		m.StopAllEffects()
	case ControlReset:
		m.FreeAllEffects()
	case ControlPause:
		m.paused = true
	case ControlContinue:
		m.paused = false
	default:
		debug.Debug("DeviceControl", "unknown code", v.Control)
	}
}

// DeviceGain reportId == 0x0d
func (m *PIDHandler) DeviceGain(b []byte) {
	var v DeviceGainOutputData
	_ = v.UnmarshalBinary(b)
	m.gain = v.Gain
}

// SetCustomForce reportId == 0x0e
func (m *PIDHandler) SetCustomForce(b []byte) {
	var v SetCustomForceOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}
