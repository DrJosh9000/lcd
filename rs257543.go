// Package lcd is a library for displaying on certain LCDs via GPIO pins (using
// periph.io).
package lcd // import "github.com/DrJosh9000/lcd"

import (
	"context"
	"time"

	"periph.io/x/conn/v3/gpio"
)

// Provides a basic translation from hex digits into segments.
var defaultRuneMap = map[rune]uint8{
	//     GFABCDE.
	' ': 0b00000000,
	'-': 0b10000000,
	'_': 0b00000100,
	'0': 0b01111110,
	'1': 0b00011000,
	'2': 0b10110110,
	'3': 0b10111100,
	'4': 0b11011000,
	'5': 0b11101100,
	'6': 0b11101110,
	'7': 0b00111000,
	'8': 0b11111110,
	'9': 0b11111100,
	'A': 0b11111010,
	'B': 0b11001110,
	'C': 0b01100110,
	'D': 0b10011110,
	'E': 0b11100110,
	'F': 0b11100010,
}

// RS257543 implements a driver for the RS 257-543 module (a discontinued
// 4-digit LCD module based on the Hughes 0438A chip). DEG, COL, and CUR are
// optional, and if provided, are assumed to each be connected via an XOR gate
// with BP as described in the data sheet. RuneMap is also optional and if
// provided, is used by Display to translate each rune into 8 segments.
type RS257543 struct {
	LD, CLK, DIN  gpio.PinIO // required, as described in the data sheet
	DEG, COL, CUR gpio.PinIO // optional

	RuneMap map[rune]uint8 // optional, uses a default map if nil
}

// Clear resets all segments on the display.
func (r *RS257543) Clear() {
	r.RawDisplay(0)
	// Reset colon, cursor, degree symbols
	if r.CUR != nil {
		r.CUR.Out(gpio.Low)
	}
	if r.COL != nil {
		r.COL.Out(gpio.Low)
	}
	if r.DEG != nil {
		r.DEG.Out(gpio.Low)
	}
}

// CycleDigits animates a simple test pattern consisting of hex digits.
func (r *RS257543) CycleDigits(ctx context.Context) {
	hex := "0123456789ABCDEF012"
	off := 0
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		r.Display(hex[off : off+4])
		select {
		case <-t.C:
		case <-ctx.Done():
			return
		}
		off++
		off %= 16
	}
}

// Display displays a string on the module. Each rune in the string is
// translated to a digit, except for '.', which is translated into the DP
// ("decimal point") segment of the following digit (the DP segment is to the
// left of each digit).
func (r *RS257543) Display(s string) {
	r.RawDisplay(r.toBits(s))
}

func (r *RS257543) toBits(s string) uint32 {
	runes := r.RuneMap
	if runes == nil {
		runes = defaultRuneMap
	}
	var x uint32
	rp := false
	for _, r := range s {
		if r == '.' {
			rp = true
			continue
		}
		x <<= 8
		x += uint32(runes[r])
		if rp {
			x++
			rp = false
		}
	}
	return x
}

// RawDisplay loads bits into the display directly.
func (r *RS257543) RawDisplay(bits uint32) {
	// Load bits into the shift register:
	// - bit is read on falling edge of CLK
	// - maximum clock frequency is 1.5 MHz - this ticker causes CLK to run at
	//   0.5MHz.
	t := time.NewTicker(time.Microsecond)
	defer t.Stop()
	for i := 0; i < 32; i++ {
		r.DIN.Out(bits&1 == 1)
		bits >>= 1

		r.CLK.Out(gpio.High)
		<-t.C
		r.CLK.Out(gpio.Low)
		<-t.C
	}
	// Then load from shift register into latches.
	r.LD.Out(gpio.High)
	<-t.C
	r.LD.Out(gpio.Low)
}
