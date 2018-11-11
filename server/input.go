package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"time"

	uinput "github.com/ynsta/uinput"

	"github.com/google/gousb"
)

const (
	keyA      uint64 = 1
	keyB      uint64 = 1 << 1
	keyX      uint64 = 1 << 2
	keyY      uint64 = 1 << 3
	keyLStick uint64 = 1 << 4
	keyRStick uint64 = 1 << 5
	keyL      uint64 = 1 << 6
	keyR      uint64 = 1 << 7
	keyZL     uint64 = 1 << 8
	keyZR     uint64 = 1 << 9
	keyPlus   uint64 = 1 << 10
	keyMinus  uint64 = 1 << 11
	keyLeft   uint64 = 1 << 12
	keyUp     uint64 = 1 << 13
	keyRight  uint64 = 1 << 14
	keyDown   uint64 = 1 << 15
)

type inputPackage struct {
	HeldKeys uint64
	LJoyX    int32
	LJoyY    int32
	RJoyX    int32
	RJoyY    int32
}

func keyHeld(key uint64, pkg inputPackage) uinput.EventValue {
	if (pkg.HeldKeys & key) > 0 {
		return uinput.EventValue(1)
	}
	return uinput.EventValue(0)
}

func inputHandler(ep *gousb.InEndpoint) {
	var wd uinput.WriteDevice
	var ui uinput.UInput
	wd.Open()
	if err := ui.Init(
		&wd,
		"Microsoft X-Box 360 pad",
		0x045e,
		0x028e,
		0x110,
		[]uinput.EventCode{uinput.BTN_START,
			uinput.BTN_MODE,
			uinput.BTN_SELECT,
			uinput.BTN_A,
			uinput.BTN_B,
			uinput.BTN_X,
			uinput.BTN_Y,
			uinput.BTN_TL,
			uinput.BTN_TR,
			uinput.BTN_THUMBL,
			uinput.BTN_THUMBR,
		},
		[]uinput.EventCode{},
		[]uinput.AxisSetup{
			uinput.AxisSetup{Code: uinput.ABS_X, Min: -32768, Max: 32767, Fuzz: 16, Flat: 128},
			uinput.AxisSetup{Code: uinput.ABS_Y, Min: -32768, Max: 32767, Fuzz: 16, Flat: 128},
			uinput.AxisSetup{Code: uinput.ABS_RX, Min: -32768, Max: 32767, Fuzz: 16, Flat: 128},
			uinput.AxisSetup{Code: uinput.ABS_RY, Min: -32768, Max: 32767, Fuzz: 16, Flat: 128},
			uinput.AxisSetup{Code: uinput.ABS_Z, Min: 0, Max: 255, Fuzz: 0, Flat: 0},
			uinput.AxisSetup{Code: uinput.ABS_RZ, Min: 0, Max: 255, Fuzz: 0, Flat: 0},
			uinput.AxisSetup{Code: uinput.ABS_HAT0X, Min: -1, Max: 1, Fuzz: 0, Flat: 0},
			uinput.AxisSetup{Code: uinput.ABS_HAT0Y, Min: -1, Max: 1, Fuzz: 0, Flat: 0},
		},
		false,
	); err != nil {
		log.Fatal(err)
	}

	pkg := inputPackage{}
	by := make([]byte, 24)

	for {
		time.Sleep(time.Millisecond * 20)
		read, err := ep.Read(by)
		if read != 24 || err != nil {
			continue
		}
		buffer := bytes.NewBuffer(by)
		binary.Read(buffer, binary.LittleEndian, &pkg)

		// Keys
		ui.KeyEvent(uinput.BTN_B, keyHeld(keyA, pkg))
		ui.KeyEvent(uinput.BTN_A, keyHeld(keyB, pkg))
		ui.KeyEvent(uinput.BTN_Y, keyHeld(keyX, pkg))
		ui.KeyEvent(uinput.BTN_X, keyHeld(keyY, pkg))
		ui.KeyEvent(uinput.BTN_START, keyHeld(keyPlus, pkg))
		ui.KeyEvent(uinput.BTN_SELECT, keyHeld(keyMinus, pkg))
		ui.KeyEvent(uinput.BTN_TL, keyHeld(keyL, pkg))
		ui.KeyEvent(uinput.BTN_TR, keyHeld(keyR, pkg))
		ui.KeyEvent(uinput.BTN_THUMBL, keyHeld(keyLStick, pkg))
		ui.KeyEvent(uinput.BTN_THUMBR, keyHeld(keyRStick, pkg))

		ui.AbsEvent(uinput.ABS_X, uinput.EventValue(pkg.LJoyX))
		ui.AbsEvent(uinput.ABS_Y, uinput.EventValue(-pkg.LJoyY))
		ui.AbsEvent(uinput.ABS_RX, uinput.EventValue(pkg.RJoyX))
		ui.AbsEvent(uinput.ABS_RY, uinput.EventValue(-pkg.RJoyY))

		// Dpad
		if (pkg.HeldKeys & keyLeft) > 0 {
			ui.AbsEvent(uinput.ABS_HAT0X, uinput.EventValue(-1))
		} else if (pkg.HeldKeys & keyRight) > 0 {
			ui.AbsEvent(uinput.ABS_HAT0X, uinput.EventValue(1))
		} else {
			ui.AbsEvent(uinput.ABS_HAT0X, uinput.EventValue(0))
		}

		if (pkg.HeldKeys & keyDown) > 0 {
			ui.AbsEvent(uinput.ABS_HAT0Y, uinput.EventValue(1))
		} else if (pkg.HeldKeys & keyUp) > 0 {
			ui.AbsEvent(uinput.ABS_HAT0Y, uinput.EventValue(-1))
		} else {
			ui.AbsEvent(uinput.ABS_HAT0Y, uinput.EventValue(0))
		}

		// Back-Triggers (?)
		if (pkg.HeldKeys & keyZL) > 0 {
			ui.AbsEvent(uinput.ABS_Z, uinput.EventValue(255))
		} else {
			ui.AbsEvent(uinput.ABS_Z, uinput.EventValue(0))
		}

		if (pkg.HeldKeys & keyZR) > 0 {
			ui.AbsEvent(uinput.ABS_RZ, uinput.EventValue(255))
		} else {
			ui.AbsEvent(uinput.ABS_RZ, uinput.EventValue(0))
		}

		ui.SynEvent()
	}
}
