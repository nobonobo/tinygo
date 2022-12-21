package joystick

import (
	"machine"
	"machine/debug"
	"machine/usb"
)

const (
	jsEndpointOut = usb.HID_ENDPOINT_OUT // from PC
	jsEndpointIn  = usb.HID_ENDPOINT_IN  // to PC
)

func ApplyGain(value int16, gain uint8) int32 {
	return int32(value) * int32(gain) / 255
}

func ApplyEnvelope(effect TEffectState, value int32) int32 {
	magnitude := ApplyGain(effect.Magnitude, effect.Gain)
	attackLevel := ApplyGain(effect.AttackLevel, effect.Gain)
	fadeLevel := ApplyGain(effect.FadeLevel, effect.Gain)
	newValue := magnitude
	attackTime := int32(effect.AttackTime)
	fadeTime := int32(effect.FadeTime)
	elapsedTime := int32(effect.ElapsedTime)
	duration := int32(effect.Duration)

	if elapsedTime < attackTime {
		newValue = (magnitude - attackLevel) * elapsedTime / attackTime
		newValue = newValue + attackLevel
	}
	if elapsedTime > duration-fadeTime {
		newValue = (magnitude - fadeLevel) * (duration - elapsedTime)
		newValue = newValue / fadeTime
		newValue = newValue + fadeLevel
	}
	newValue = newValue * value / magnitude
	return newValue
}

type PIDHandler struct {
	buf          *RingBuffer
	waitTxc      bool
	effectStates [MAX_EFFECTS]TEffectState
	enabled      bool
	paused       bool
	gain         uint8
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
	case ReportSetEffect:
		m.SetEffect(b)
	case ReportSetEnvelope:
		m.SetEnvelope(b)
	case ReportSetCondition:
		m.SetCondition(b)
	case ReportSetPeriodic:
		m.SetPeriodic(b)
	case ReportSetConstantForce:
		m.SetConstantForce(b)
	case ReportSetRampForce:
		m.SetRampForce(b)
	case ReportSetCustomForceData:
		m.SetCustomForceData(b)
	case ReportSetDownloadForceSample:
		m.SetDownloadForceSample(b)
	case ReportEffectOperation:
		m.EffectOperation(b)
	case ReportBlockFree:
		m.BlockFree(b)
	case ReportDeviceControl:
		m.DeviceControl(b)
	case ReportDeviceGain:
		m.DeviceGain(b)
	case ReportSetCustomForce:
		m.SetCustomForce(b)
	default:
		debug.Debug("unknown rx", debug.Hex(b))
		return
	}
	debug.Debug("rx", debug.Hex(b))
}

func (m *PIDHandler) GetNextFreeEffect() uint8 {
	for i, v := range m.effectStates {
		if v.State == MEFFECTSTATE_FREE {
			v.State = MEFFECTSTATE_ALLOCATED
			return uint8(i)
		}
	}
	return 0
}

func (m *PIDHandler) StopAllEffects() {
	for id := uint8(0); id < MAX_EFFECTS; id++ {
		m.StopEffect(id)
	}
}

func (m *PIDHandler) StartEffect(id uint8) {
	if id >= MAX_EFFECTS {
		debug.Debug("StartEffect", "effect index out of range", id)
		return
	}
	effect := TEffectState{}
	effect.State = MEFFECTSTATE_PLAYING
	effect.ElapsedTime = 0
	effect.StartTime = 0
	m.effectStates[id] = effect
}

func (m *PIDHandler) StopEffect(id uint8) {
	if id >= MAX_EFFECTS {
		debug.Debug("StopEffect", "effect index out of range", id)
		return
	}
	effect := m.effectStates[id]
	effect.State &= ^MEFFECTSTATE_PLAYING
	m.effectStates[id] = effect
}

func (m *PIDHandler) FreeAllEffects() {
	for id := uint8(0); id < MAX_EFFECTS; id++ {
		m.FreeEffect(id)
	}
}

func (m *PIDHandler) FreeEffect(id uint8) {
	if id >= MAX_EFFECTS {
		debug.Debug("FreeEffect", "effect index out of range", id)
		return
	}
	state := m.effectStates[id]
	state.State = MEFFECTSTATE_FREE
	m.effectStates[id] = state
}

// SetEffect reportId == 0x01
func (m *PIDHandler) SetEffect(b []byte) {
	var v SetEffectOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// SetEnvelope reportId == 0x02
func (m *PIDHandler) SetEnvelope(b []byte) {
	var v SetEnvelopeOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// SetCondition reportId == 0x03
func (m *PIDHandler) SetCondition(b []byte) {
	var v SetConditionOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// SetPeriodic reportId == 0x04
func (m *PIDHandler) SetPeriodic(b []byte) {
	var v SetPeriodicOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// SetConstantForce reportId == 0x05
func (m *PIDHandler) SetConstantForce(b []byte) {
	var v SetConstantForceOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// SetRampForce reportId == 0x06
func (m *PIDHandler) SetRampForce(b []byte) {
	var v SetRampForceOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// SetCustomForceData reportId == 0x07
func (m *PIDHandler) SetCustomForceData(b []byte) {
	var v SetCustomForceDataOutputData
	_ = v.UnmarshalBinary(b)
	// TODO: implement
}

// SetDownloadForceSample reportId == 0x08
func (m *PIDHandler) SetDownloadForceSample(b []byte) {
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
