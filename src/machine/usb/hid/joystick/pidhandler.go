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
	buf     *RingBuffer
	waitTxc bool
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
	// TODO: implement
	if len(b) == 0 {
		return
	}
	reportId := b[0]
	switch reportId {
	case 0x01:
		var v SetEffectOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("SetEffectOutputData:", err.Error())
			return
		}
		//debug.Debug("SetEffectOutputData:", v)
	case 0x02:
		var v SetEnvelopeOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("SetEnvelopeOutputData:", err.Error())
			return
		}
		//debug.Debug("SetEnvelopeOutputData:", v)
	case 0x03:
		var v SetConditionOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("SetConditionOutputData:", err.Error())
			return
		}
		//debug.Debug("SetConditionOutputData:", v)
	case 0x04:
		var v SetPeriodicOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("SetPeriodicOutputData:", err.Error())
			return
		}
		//debug.Debug("SetPeriodicOutputData:", v)
	case 0x05:
		var v SetConstantForceOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("SetConstantForceOutputData:", err.Error())
			return
		}
		//debug.Debug("SetConstantForceOutputData:", v)
	case 0x06:
		var v SetRampForceOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("SetRampForceOutputData:", err.Error())
			return
		}
		//debug.Debug("SetRampForceOutputData:", v)
	case 0x07:
		var v SetCustomForceDataOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("SetCustomForceDataOutputData:", err.Error())
			return
		}
		//debug.Debug("SetCustomForceDataOutputData:", v)
	case 0x08:
		var v SetDownloadForceSampleOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("SetDownloadForceSampleOutputData:", err.Error())
			return
		}
		//debug.Debug("SetDownloadForceSampleOutputData:", v)
	case 0x0a:
		var v EffectOperationOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("EffectOperationOutputData:", err.Error())
			return
		}
		//debug.Debug("EffectOperationOutputData:", v)
	case 0x0b:
		var v BlockFreeOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("BlockFreeOutputData:", err.Error())
			return
		}
		//debug.Debug("BlockFreeOutputData:", v)
	case 0x0c:
		var v DeviceControlOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("DeviceControlOutputData:", err.Error())
			return
		}
		//debug.Debug("DeviceControlOutputData:", v)
	case 0x0d:
		var v DeviceGainOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("DeviceGainOutputData:", err.Error())
			return
		}
		//debug.Debug("DeviceGainOutputData:", v)
	case 0x0e:
		var v SetCustomForceOutputData
		if err := v.UnmarshalBinary(b); err != nil {
			debug.Debug("SetCustomForceOutputData:", err.Error())
			return
		}
		//debug.Debug("SetCustomForceOutputData:", v)
	default:
		debug.Debug("unknown rx:", debug.Hex(b))
	}
}
