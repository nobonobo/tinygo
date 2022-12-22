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

	REPORT_TYPE_INPUT   = 1
	REPORT_TYPE_OUTPUT  = 2
	REPORT_TYPE_FEATURE = 3
)

type hidDevicer interface {
	Handler() bool
}

type customHandler interface {
	RxHandler(b []byte)
	SetupHandler(setup usb.Setup) (ok bool)
}

var devices [5]hidDevicer
var size int

// SetHandler sets the handler. Only the first time it is called, it
// calls machine.EnableHID for USB configuration
func SetHandler(d hidDevicer) {
	if size == 0 {
		if r, ok := d.(customHandler); ok {
			machine.EnableHID(handler, r.RxHandler, r.SetupHandler)
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
	return ok
}

// SendUSBPacket sends a HIDPacket.
func SendUSBPacket(b []byte) {
	machine.SendUSBInPacket(hidEndpointIn, b)
}
