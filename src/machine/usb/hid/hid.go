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

func setupHandler(setup usb.Setup) bool {
	ok := false
	if setup.BmRequestType == usb_SET_REPORT_TYPE && setup.BRequest == usb_SET_IDLE {
		machine.SendZlp()
		ok = true
	}
	if setup.BmRequestType == 0xa1 && setup.BRequest == 0x01 {
		machine.SendZlp()
		ok = true
	}
	return ok
}

// SendUSBPacket sends a HIDPacket.
func SendUSBPacket(b []byte) {
	machine.SendUSBInPacket(hidEndpointIn, b)
}
