package main

import (
	"fmt"
	"math"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

func drawDot(surface *sdl.Surface, x, y int, radius int, color sdl.Color) {
	rect := sdl.Rect{
		X: int32(x - radius),
		Y: int32(y - radius),
		W: int32(2 * radius),
		H: int32(2 * radius),
	}
	pixel := sdl.MapRGBA(surface.Format, color.R, color.G, color.B, color.A)
	surface.FillRect(&rect, pixel)
}

func positionAtAngle(cx, cy, a, r0, aq, r1 int) (int, int) {
	rad := float64(a) / 180.0 * math.Pi
	rad_aq := float64(aq) / 180.0 * math.Pi

	x := cx + int(float64(r0)*math.Sin(rad)+float64(r1)*math.Sin(rad_aq))
	y := cy - int(float64(r0)*math.Cos(rad)+float64(r1)*math.Cos(rad_aq))

	return x, y
}

func drawDial(surface *sdl.Surface, color sdl.Color, min, dial int) {

	a0 := (min * 120 / 60) % 360
	h := min / 60
	s := h % 4

	q := (h / 3) * 90
	aq := q
	if a0 > 120 {
		aq += (a0 - 120) * 90 / 240
	}

	if dial == 0 {
		fmt.Printf("min %d , s = %d - a = %d, q = %d, aq = %d\n", min, s, a0, q, aq)
	}

	r0 := 200
	r1 := 60
	rd := 30

	b1 := dial % 3
	b2 := dial / 3

	a0 += 120 * b1
	aq += 90 * b2

	a0 -= 60
	a1 := a0 + aq

	x, y := positionAtAngle(300, 300, a0, r0, a1, r1)

	drawDot(surface, x, y, rd, color)
}

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		600, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	running := true

	go func() {
		min := 0

		color := sdl.Color{R: 255, G: 0, B: 255, A: 255}    // purple
		color1 := sdl.Color{R: 255, G: 255, B: 255, A: 255} // purple
		color2 := sdl.Color{R: 255, G: 0, B: 0, A: 255}     // red

		for {
			if !running {
				return
			}
			surface.FillRect(nil, 0)

			for j := range 12 {
				var c sdl.Color
				if j == 0 {
					c = color1
				} else if j < 3 {
					c = color2
				} else {
					c = color
				}
				drawDial(surface, c, min, j)
			}
			window.UpdateSurface()

			min += 1
			min %= 720

			<-time.After(100 * time.Millisecond)
		}
	}()

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			et := event.GetType()

			switch et {
			case sdl.QUIT:
				running = false
			}
		}
	}
}
