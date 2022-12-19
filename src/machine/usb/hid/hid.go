package hid

import (
	"errors"
	"machine"
	"machine/usb"
)

// from usb-hid.go
var (
	ErrHIDInvalidPort    = errors.New("invalid USB port")
	ErrHIDInvalidCore    = errors.New("invalid USB core")
	ErrHIDReportTransfer = errors.New("failed to transfer HID report")
)

const (
	hidEndpointIn  = usb.HID_ENDPOINT_IN
	hidEndpointOut = usb.HID_ENDPOINT_OUT

	usb_SET_REPORT_TYPE = 33
	usb_SET_IDLE        = 10
)

type hidDevicer interface {
	Handler() bool
}

type hidReceiver interface {
	RxHandler(b []byte)
}

var devices [5]hidDevicer
var size int

// SetHandler sets the handler. Only the first time it is called, it
// calls machine.EnableHID for USB configuration
func SetHandler(d hidDevicer) {
	if size == 0 {
		if r, ok := d.(hidReceiver); ok {
			machine.EnableHID(handler, r.RxHandler, setupHandler)
		} else {
			machine.EnableHID(handler, nil, setupHandler)
		}
	}

	devices[size] = d
	size++
}

func handler() {
	for _, d := range devices {
		if d == nil {
			continue
		}
		if done := d.Handler(); done {
			return
		}
	}
}

func setupHandler(setup usb.Setup) (ok bool) {
	switch setup.BmRequestType {
	case usb.REQUEST_DEVICETOHOST_CLASS_INTERFACE:
		switch setup.BRequest {
		case usb.GET_REPORT:
			b := []byte{0x06, 0x01, 0x01, 0xe4, 0x06}
			machine.SendUSBInPacket(0, b)
			return true
		case usb.GET_IDLE:
			// TODO: imprement Send8(idle)
		case usb.GET_PROTOCOL:
			// TODO: imprement Send8(protocol)
			machine.SendZlp()
			return true
		}
	case usb.REQUEST_HOSTTODEVICE_CLASS_INTERFACE:
		switch setup.BRequest {
		case usb.SET_IDLE:
			// TODO: imprement use setup.WValueL
			machine.SendZlp()
			return true
		case usb.SET_REPORT:
			// TODO: imprement use setup
			machine.SendZlp()
			return true
		case usb.SET_PROTOCOL:
			// TODO: imprement use setup.WValueL
			machine.SendZlp()
			return true
		}
	}
	return false
}

// SendUSBPacket sends a HIDPacket.
func SendUSBPacket(b []byte) {
	machine.SendUSBInPacket(hidEndpointIn, b)
}
