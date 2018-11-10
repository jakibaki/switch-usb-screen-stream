package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/google/gousb"
)

func CaptureScreen(c *xgb.Conn) (*image.RGBA, error) {
	screen := xproto.Setup(c).DefaultScreen(c)
	x := screen.WidthInPixels
	y := screen.HeightInPixels
	rect := image.Rect(0, 0, int(x), int(y))

	xImg, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(screen.Root), int16(rect.Min.X), int16(rect.Min.Y), uint16(x), uint16(y), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}

	data := xImg.Data

	for i := 0; i < len(data); i += 4 {
		data[i], data[i+2], data[i+3] = data[i+2], data[i], 255
	}

	img := &image.RGBA{data, 4 * int(x), image.Rect(0, 0, int(x), int(y))}
	return img, nil
}

func main() {
	exec.Command("xrandr", "--output", "eDP1", "--mode", "640x360").Output()
	defer exec.Command("xrandr", "--output", "eDP1", "--auto").Output()

	c, err := xgb.NewConn()
	if err != nil {
		log.Fatalf("Could not get xgb-conn: %v", err)
	}
	defer c.Close()

	ctx := gousb.NewContext()
	defer ctx.Close()

	dev, err := ctx.OpenDeviceWithVIDPID(0x057e, 0x3000)
	if err != nil {
		log.Fatalf("Could not open a device: %v", err)
	}
	defer dev.Close()

	intf, done, err := dev.DefaultInterface()
	if err != nil {
		log.Fatalf("%s.DefaultInterface(): %v", dev, err)
	}
	defer done()

	// Open an OUT endpoint.
	ep, err := intf.OutEndpoint(0x1)
	if err != nil {
		log.Fatalf("%s.OutEndpoint(1): %v", intf, err)
	}

	inp, err := intf.InEndpoint(0x81)
	if err != nil {
		log.Fatalf("%s.InEndpoint(1): %v", intf, err)

	}
	go inputHandler(inp)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	for {
		select {
		case <-ch:
			fmt.Println("Shutting down...")
			return
		default:

		}
		img, err := CaptureScreen(c)

		if err != nil {
			panic(err)
		}
		ep.Write(img.Pix)
	}
}
