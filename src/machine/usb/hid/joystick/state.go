package joystick

import "encoding/binary"

const (
	FORCE_FEEDBACK_MAXGAIN = 100
)

type HatDirection uint8

const (
	HatUp HatDirection = iota
	HatRightUp
	HatRight
	HatRightDown
	HatDown
	HatLeftDown
	HatLeft
	HatLeftUp
	HatCenter
)

type Gains struct {
	TotalGain        uint8
	ConstantGain     uint8
	RampGain         uint8
	SquareGain       uint8
	SineGain         uint8
	TriangleGain     uint8
	SawtoothdownGain uint8
	SawtoothupGain   uint8
	SpringGain       uint8
	DamperGain       uint8
	InertiaGain      uint8
	FrictionGain     uint8
	CustomGain       uint8
}

type Constraint struct {
	MinIn  int
	MaxIn  int
	MinOut int16
	MaxOut int16
}

type AxisValue struct {
	constraint Constraint
	Value      int
}

func fit(x, in_min, in_max int, out_min, out_max int16) int16 {
	return int16((x-in_min)*(int(out_max)-int(out_min))/(in_max-in_min) + int(out_min))
}

func limit(v, max int) int {
	if v > max {
		v = max
	} else if v < -max {
		v = -max
	}
	return v
}

type Definitions struct {
	ButtonCnt    int
	HatSwitchCnt int
	AxisDefs     []Constraint
}

func (c Definitions) NewState() State {
	axises := make([]*AxisValue, 0, len(c.AxisDefs))
	for _, v := range c.AxisDefs {

		axises = append(axises, &AxisValue{
			constraint: v,
			Value:      0,
		})
	}
	btnSize := (c.ButtonCnt + 7) / 8
	bufSize := btnSize
	if c.HatSwitchCnt > 0 {
		bufSize++
	}
	bufSize += len(axises) * 2
	return State{
		buf:         make([]byte, bufSize),
		Buttons:     make([]byte, btnSize),
		HatSwitches: make([]HatDirection, c.HatSwitchCnt),
		Axises:      axises,
	}
}

type State struct {
	buf         []byte
	Buttons     []byte
	HatSwitches []HatDirection
	Axises      []*AxisValue
}

func (s State) MarshalBinary() ([]byte, error) {
	s.buf = s.buf[0:0]
	s.buf = append(s.buf, s.Buttons...)
	if len(s.HatSwitches) > 0 {
		hat := byte(0)
		for _, v := range s.HatSwitches {
			hat <<= 4
			hat |= byte(v & 0xf)
		}
		s.buf = append(s.buf, hat)
	}
	for _, v := range s.Axises {
		c := v.constraint
		val := fit(v.Value, c.MinIn, c.MaxIn, c.MinOut, c.MaxOut)
		s.buf = binary.LittleEndian.AppendUint16(s.buf, uint16(val))
	}
	return s.buf, nil
}

func (s State) Button(index int) bool {
	idx := index / 8
	bit := uint8(1 << (index % 8))
	return s.Buttons[idx]&bit > 0
}

func (s State) SetButton(index int, push bool) {
	idx := index / 8
	bit := uint8(1 << (index % 8))
	b := s.Buttons[idx]
	b &= ^bit
	if push {
		b |= bit
	}
	s.Buttons[idx] = b
}

func (s State) Hat(index int) HatDirection {
	return s.HatSwitches[index]
}

func (s State) SetHat(index int, dir HatDirection) {
	s.HatSwitches[index] = dir
}

func (s State) Axis(index int) int {
	return s.Axises[index].Value
}

func (s State) SetAxis(index int, v int) {
	s.Axises[index].Value = v
}
