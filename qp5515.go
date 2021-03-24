package lcd

import (
	"time"

	"periph.io/x/conn/v3/gpio"
)

// QP5515 implements a driver for a QP-5515 or QP-5516 module. It assumes 8-bit
// mode (all 8 data pins are connected). The contrast adjust (pin 3 on mine)
// should be connected to the centre of a 10k trimpot between 0 and 5v.
type QP5515 struct {
	RS, RW, E gpio.PinIO    // register select, read/write, enable signal
	DB        [8]gpio.PinIO // data bits 0 - 7
}

// Display writes data to the display.
func (q *QP5515) Display(s string) {
	for _, r := range s {
		q.BusyWait()
		q.WriteData(uint8(r))
	}
}

// ReadBFAC reads the busy flag and address counter.
func (q *QP5515) ReadBFAC() uint8 {
	q.RS.Out(gpio.Low)
	return q.rawReadData()
}

// BusyWait waits until the busy flag is cleared.
func (q *QP5515) BusyWait() {
	// check it now
	bfac := q.ReadBFAC()
	if bfac&0b10000000 == 0 {
		return
	}
	// ok then, check every 40µs
	t := time.NewTicker(40 * time.Microsecond)
	defer t.Stop()
	for {
		<-t.C
		bfac := q.ReadBFAC()
		if bfac&0b10000000 == 0 {
			return
		}
	}
}

// ReadData reads a value from CG RAM or DD RAM.
func (q *QP5515) ReadData() uint8 {
	q.RS.Out(gpio.High)
	return q.rawReadData()
}

// RawFunction performs a function or sets an address for the next write.
func (q *QP5515) RawFunction(a uint8) {
	q.RS.Out(gpio.Low)
	q.rawWriteData(a)
}

// Clear clears the display and returns the cursor to the home position.
func (q *QP5515) Clear() {
	q.RawFunction(0b00000001)
}

// ReturnHome returns the cursor to the home position and resets the display
// shift.
func (q *QP5515) ReturnHome() {
	q.RawFunction(0b00000010)
}

// SetEntryMode sets the data entry direction and whether to also shift.
func (q *QP5515) SetEntryMode(increment, shift bool) {
	a := uint8(0b00000100)
	if increment {
		a += 0b00000010
	}
	if shift {
		a += 0b00000001
	}
	q.RawFunction(a)
}

// SetDisplayMode turns on/off the whole display, cursor, or cursor-blinking.
func (q *QP5515) SetDisplayMode(display, cursor, blink bool) {
	a := uint8(0b00001000)
	if display {
		a += 0b00000100
	}
	if cursor {
		a += 0b00000010
	}
	if blink {
		a += 0b00000001
	}
	q.RawFunction(a)
}

// SetDisplayShiftOrCursorMove sets display shift or cursor move, and direction.
func (q *QP5515) SetDisplayShiftOrCursorMove(shift, right bool) {
	a := uint8(0b00010000)
	if shift {
		a += 0b00001000
	}
	if right {
		a += 0b00000100
	}
	q.RawFunction(a)
}

// SetFunction sets the interface data length, number of display lines, and
// character font.
// eightbit = false means switch to 4-bit operation.
// twolines = false means use 1 display line.
// largefont = false mease use 5x7 font instead of 5x10 font.
func (q *QP5515) SetFunction(eightbit, twolines, largefont bool) {
	a := uint8(0b00100000)
	if eightbit {
		a += 0b00010000
	}
	if twolines {
		a += 0b00001000
	}
	if largefont {
		a += 0b00000100
	}
	q.RawFunction(a)
}

// SetCGAddress sets the CG RAM address (0 <= a < 64).
func (q *QP5515) SetCGAddress(a uint8) {
	a += 0b01000000
	q.RawFunction(a)
}

// SetDDAddress sets the DD RAM address (0 <= a < 128).
func (q *QP5515) SetDDAddress(a uint8) {
	a += 0b10000000
	q.RawFunction(a)
}

// WriteData writes a value to CG RAM or DD RAM.
func (q *QP5515) WriteData(b uint8) {
	q.RS.Out(gpio.High)
	q.rawWriteData(b)
}

func (q *QP5515) rawReadData() uint8 {
	// Ensure the data pins are inputs
	// TODO: handle errors
	for i := range q.DB {
		q.DB[i].In(gpio.PullNoChange, gpio.NoEdge)
	}

	q.RW.Out(gpio.High)
	time.Sleep(250 * time.Nanosecond) // tAS > 100ns
	q.E.Out(gpio.High)
	time.Sleep(250 * time.Nanosecond) // tDDR < 190ns

	var b uint8
	for i := range q.DB {
		if !q.DB[i].Read() {
			continue
		}
		b += 1 << i
	}

	q.E.Out(gpio.Low)
	time.Sleep(500 * time.Nanosecond) // (tCYCE > 1000ns, PWEH > 450ns)

	return b
}

func (q *QP5515) rawWriteData(b uint8) {
	// Ensure the data pins are outputs
	for i := range q.DB {
		q.DB[i].Out(gpio.Low)
	}

	q.RW.Out(gpio.Low)
	time.Sleep(250 * time.Nanosecond) // tAS > 100ns
	q.E.Out(gpio.High)
	time.Sleep(250 * time.Nanosecond) // tDSW > 100ns

	for i := range q.DB {
		q.DB[i].Out(b&(1<<i) != 0)
	}

	q.E.Out(gpio.Low)
	time.Sleep(500 * time.Nanosecond) // (tCYCE > 1000ns, PWEH > 450ns)
}
