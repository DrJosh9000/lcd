package lcd

import (
	"os"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

func ExampleQP5515() {
	host.Init()
	q := &QP5515{
		RS: gpioreg.ByName("27"),
		RW: gpioreg.ByName("20"),
		E:  gpioreg.ByName("26"),
		DB: [8]gpio.PinIO{
			0: gpioreg.ByName("19"),
			1: gpioreg.ByName("25"),
			2: gpioreg.ByName("18"),
			3: gpioreg.ByName("24"),
			4: gpioreg.ByName("17"),
			5: gpioreg.ByName("23"),
			6: gpioreg.ByName("16"),
			7: gpioreg.ByName("22"),
		},
	}
	q.BusyWait()
	q.Clear()
	q.BusyWait()
	q.SetFunction(true, true, false) // 8-bit, 2-line, small font
	q.BusyWait()
	q.SetDisplayMode(true, false, false) // display on, no cursor or blink
	q.BusyWait()
	q.SetEntryMode(true, false) // increment address, no shift
	q.BusyWait()
	if len(os.Args) == 1 {
		q.Display("github.com/")
		q.BusyWait()
		q.SetDDAddress(0x40)
		q.BusyWait()
		q.Display("DrJosh9000/lcd")
		return
	}
	q.Display(os.Args[1])
	if len(os.Args[1]) > 16 {
		q.BusyWait()
		q.SetDDAddress(0x40)
		q.BusyWait()
		q.Display(os.Args[1][16:])
	}
}
