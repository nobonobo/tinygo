package joystick

import (
	"machine/usb/hid"
)

var js *Joystick

type Joystick struct {
	State
	handler *PIDHandler
	gains   Gains
	params  EffectParams
}

func init() {
	if js == nil {
		js = &Joystick{
			handler: NewPIDHandler(),
			gains: Gains{
				TotalGain:    255,
				ConstantGain: 255,
			},
			params: EffectParams{},
		}
		js.handler.FreeAllEffects()
		hid.SetHandler(js.handler)
	}
}

// New returns the USB js port.
// Deprecated, better to just use Port()
func New(def Definitions) *Joystick {
	m := Port()
	m.State = def.NewState()
	return m
}

// Port returns the USB Joystick port.
func Port() *Joystick {
	return js
}

func (m *Joystick) SendState() {
	m.handler.SendState(m.State)
}

func (m *Joystick) CalcForces() []int32 {
	forces := []int32{0, 0}
	for _, ef := range m.handler.effectStates {
		if ef.State == MEFFECTSTATE_PLAYING &&
			(ef.Duration == USB_DURATION_INFINITE ||
				ef.ElapsedTime <= ef.Duration) &&
			!m.handler.paused {
			//log.Printf("%d: %v", idx, ef)
			forces[0] += ef.Force(m.gains, m.params, 0)
			forces[1] += ef.Force(m.gains, m.params, 1)
		}
	}
	return forces
}

func (m *Joystick) GetCurrentEffect() *TEffectState {
	return m.handler.effectStates[m.handler.pidBlockLoad.EffectBlockIndex]
}
