package joystick

import "machine/usb/hid"

var js *Joystick

type Joystick struct {
	State
	handler *PIDHandler
}

func init() {
	if js == nil {
		js = &Joystick{
			handler: &PIDHandler{
				buf: NewRingBuffer(),
			},
		}
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
