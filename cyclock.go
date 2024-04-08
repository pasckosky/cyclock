package main

import (
	"fmt"
	"math"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	fontPath = "./assets/font.ttf"
	fontSize = 32
)

var CircleSurface *sdl.Surface

func drawDot(surface *sdl.Surface, x, y int, radius int, dc sdl.Color) {
	rect := sdl.Rect{
		X: int32(x - radius),
		Y: int32(y - radius),
		W: int32(2 * radius),
		H: int32(2 * radius),
	}
	pixel := sdl.MapRGBA(surface.Format, dc.R, dc.G, dc.B, dc.A)

	/*
		if CircleSurface == nil {
			CircleSurface, err := sdl.CreateRGBSurface(0, int32(radius)*2, int32(radius)*2, 8, 0, 0, 0, 0)
			if err != nil {
				panic(err)
			}

			rdf := float64(radius)
			for a := range 360 {
				ar := float64(a) * 180.0 / math.Pi
				px := rdf + rdf*math.Cos(ar)
				py := rdf + rdf*math.Sin(ar)

				CircleSurface.Set(int(px), int(py), color.RGBA{R: dc.R, G: dc.G, B: dc.B, A: dc.A})
			}

		}
	*/
	//surface.
	surface.FillRect(&rect, pixel)
}

func positionAtAngle(cx, cy, a, r0, aq, r1 int) (int, int) {
	rad := float64(a) / 180.0 * math.Pi
	rad_aq := float64(aq) / 180.0 * math.Pi

	x := cx + int(float64(r0)*math.Sin(rad)+float64(r1)*math.Sin(rad_aq))
	y := cy - int(float64(r0)*math.Cos(rad)+float64(r1)*math.Cos(rad_aq))

	return x, y
}

func drawDial(surface *sdl.Surface, color, selcolor sdl.Color, min int, paddle *sdl.Surface, dial int) {

	hour := dial % 12
	min += 60 * hour

	a0 := (min * 120 / 60) % 360
	h := min / 60
	s := h % 4

	q := ((h / 3) * 90) % 360
	aq := q
	if a0 > 120 {
		aq += (a0 - 120) * (90 - 360*2) / 240
	}

	c := color
	if a0 >= 0 && a0 <= 120 && aq == 0 {
		c = selcolor
	}

	if dial == 0 {
		fmt.Printf("min %d , s = %d - a = %d, q = %d, aq = %d\n", min, s, a0, q, aq)
	}

	r0 := 200
	r1 := 60
	rd := 30

	a0 -= 60
	a1 := a0 + aq

	x, y := positionAtAngle(300, 300, a0, r0, a1, r1)
	drawDot(surface, x, y, rd, c)

	paddle.Blit(nil, surface, &sdl.Rect{
		X: int32(x) - (paddle.W / 2),
		Y: int32(y) - (paddle.H / 2),
		W: 0, H: 0})

}

func main() {
	if err := ttf.Init(); err != nil {
		panic(err)
	}
	defer ttf.Quit()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// Load the font for our text
	font, err := ttf.OpenFont(fontPath, fontSize)
	if err != nil {
		panic(err)
	}
	defer font.Close()

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

	// Create a red text with the font
	paddle := make([]*sdl.Surface, 12)
	tpaddle := []string{"12", "11", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1"}
	colorText := sdl.Color{G: 255, R: 0, B: 0, A: 255}
	for j := range 12 {
		paddle[j], err = font.RenderUTF8Blended(tpaddle[j], colorText)
		if err != nil {
			panic(err)
		}
	}
	defer func() {
		for j := range 12 {
			paddle[j].Free()
		}
	}()

	go func() {
		min := 0

		color := sdl.Color{R: 255, G: 0, B: 255, A: 120}    // purple
		color1 := sdl.Color{R: 255, G: 255, B: 255, A: 120} // white
		color2 := sdl.Color{R: 255, G: 0, B: 0, A: 120}     // red
		selcolor := sdl.Color{R: 0, G: 0, B: 255, A: 255}   // yellow

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
				drawDial(surface, c, selcolor, min, paddle[j], j)
			}
			window.UpdateSurface()

			min += 5
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
