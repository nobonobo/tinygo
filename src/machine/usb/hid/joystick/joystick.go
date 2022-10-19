package joystick

import (
	"machine"
	"machine/usb"
	"machine/usb/hid"

	"machine/debug"
)

const (
	jsEndpointOut = usb.HID_ENDPOINT_OUT // from PC
	jsEndpointIn  = usb.HID_ENDPOINT_IN  // to PC
)

var Joystick *js

type js struct {
	State
	msg     [4]byte
	buf     *RingBuffer
	waitTxc bool
}

func init() {
	if Joystick == nil {
		Joystick = newJoystick()
		hid.SetHandler(Joystick)
	}
}

// New returns the USB Joystick port.
// Deprecated, better to just use Port()
func New(def Definitions) *js {
	m := Port()
	m.State = def.NewState()
	return m
}

// Port returns the USB js port.
func Port() *js {
	return Joystick
}

func newJoystick() *js {
	m := &js{
		buf: NewRingBuffer(),
	}
	return m
}

// sendUSBPacket sends a JoystickPacket.
func (m *js) sendUSBPacket(b []byte) {
	machine.SendUSBInPacket(jsEndpointIn, b)
}

// from BulkIn
func (m *js) Handler() bool {
	m.waitTxc = false
	if b, ok := m.buf.Get(); ok {
		m.waitTxc = true
		m.sendUSBPacket(b)
		return true
	}
	return false
}

func (m *js) tx(b []byte) {
	if m.waitTxc {
		m.buf.Put(b)
	} else {
		m.waitTxc = true
		m.sendUSBPacket(b)
	}
}

// from InterruptOut
func (m *js) RxHandler(b []byte) {
	debug.DebugHex("rx:", b)
	machine.SendZlp()
}

// to InterruptIn
func (m *js) SendReport(reportID byte, b []byte) {
	m.tx(append([]byte{reportID}, b...))
}

func (m *js) SendState() {
	b, _ := m.State.MarshalBinary()
	m.SendReport(1, b)
}
