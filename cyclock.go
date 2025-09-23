package main

import (
	"embed"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	fontPath = "assets/font.ttf"
	fontSize = 26
)

//go:embed assets/*.png assets/*.ttf
var assets embed.FS

func assetsFont(path string, size int) *ttf.Font {
	raw, err := assets.ReadFile(path)
	if err != nil {
		panic(err)
	}
	rwo, err := sdl.RWFromMem(raw)
	if err != nil {
		panic(err)
	}
	font, err := ttf.OpenFontRW(rwo, 0, size)
	if err != nil {
		panic(err)
	}

	return font
}

func assetsImage(path string) *sdl.Surface {
	raw, err := assets.ReadFile(path)
	if err != nil {
		panic(err)
	}
	rwo, err := sdl.RWFromMem(raw)
	if err != nil {
		panic(err)
	}
	surf, err := img.LoadRW(rwo, false)
	if err != nil {
		panic(err)
	}
	return surf
}

func drawLine(surface *sdl.Surface, x1, y1, x2, y2 int) {
	renderer, err := sdl.CreateSoftwareRenderer(surface)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	//renderer.SetDrawColor(255, 255, 255, 255)
	renderer.SetDrawColor(0, 0, 255, 0)
	//renderer.SetDrawColor(0, 0, 0, 255)
	renderer.DrawLine(int32(x1), int32(y1), int32(x2), int32(y2))
}

func drawDot(surface *sdl.Surface, x, y int, radius int, dot *sdl.Surface) {
	dot.Blit(nil, surface, &sdl.Rect{
		X: int32(x - radius),
		Y: int32(y - radius),
		W: 0, H: 0})
}

func drawDial(surface *sdl.Surface, dot *sdl.Surface, minutes int, paddle *sdl.Surface, dial int) {

	positionAtAngle := func(cx, cy, a, r0, aq, r1 int) (int, int) {
		rad := float64(a) / 180.0 * math.Pi
		rad_aq := float64(aq) / 180.0 * math.Pi

		x := cx + int(float64(r0)*math.Sin(rad)+float64(r1)*math.Sin(rad_aq))
		y := cy - int(float64(r0)*math.Cos(rad)+float64(r1)*math.Cos(rad_aq))

		return x, y
	}

	hour := dial % 12
	minutes += 60 * hour

	a0 := (minutes * 120 / 60) % 360
	h := minutes / 60

	q := ((h / 3) * 90) % 360
	aq := q

	rt0 := 120
	rt1 := 360
	if a0 >= rt0 {
		aq += (a0 - rt0) * (90 - 360*2) / (rt1 - rt0)
	}

	r0 := 210 //160 //215 //170
	r1 := 260 - r0
	rd := 20

	a0 -= 60
	a1 := a0 + aq

	x0, y0 := positionAtAngle(300, 300, a0, r0, 0, 0)
	x, y := positionAtAngle(300, 300, a0, r0, a1, r1)

	if paddle == nil {
		// just draw lines and pivots
		//drawLine(surface, 300, 300, x0, y0)
		drawLine(surface, x0, y0, x, y)
		drawDot(surface, x0, y0, rd, dot)
	} else {
		drawDot(surface, x, y, rd, dot)
		/*
			    // Set 'your_texture' as target
			SDL_SetRenderTarget(your_renderer, your_texture);

			    // We are now printing the rotated image on the texture
			SDL_RenderCopyEx(your_renderer, // we still use the renderer; it will be automatically printed into the texture 'your_texture'
			                   your_image,
			                   &srcrect,
			                   &dstrect,
			                   angle,
			                   &center,
			                   SDL_FLIP_NONE); // unless you want to flip vertically / horizontally

			    // Set the renderer as target and print the previous texture
			SDL_SetRenderTarget(your_renderer, NULL);
			SDL_RenderClear(your_renderer);
			SDL_RenderCopy (your_renderer, your_texture, NULL, NULL); // here the scale is automatically done
			SDL_RenderPresent(your_renderer);
		*/
		//paddle.CopyEx()

		paddle.Blit(nil, surface, &sdl.Rect{
			X: int32(x) - (paddle.W / 2),
			Y: int32(y) - (paddle.H / 2),
			W: 0, H: 0})
	}
}

var bigT int = 0

var debug bool = false

func getTime() (int, int) {

	if debug {
		bigT += 1
		bigT %= 1440 // 24 hours
		return bigT, (bigT / 3) % 60

	}
	t := time.Now()

	h := t.Hour()
	m := t.Minute()
	s := t.Second()

	return m + h*60, s
}

func formatTime(tim int) string {
	m := tim % 60
	h := tim / 60

	return fmt.Sprintf("time: %2d:%02d", h, m)
}

func Iif[T any](cond bool, v T, f T) T {
	if cond {
		return v
	}
	return f
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-d" {
		debug = true
	}

	if err := ttf.Init(); err != nil {
		panic(err)
	}
	defer ttf.Quit()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 600, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	update := make(chan bool)

	font := assetsFont(fontPath, fontSize)
	secBack := assetsImage("assets/back.png")
	quadrant := assetsImage("assets/quadrant.png")
	dot := assetsImage("assets/dot.png")
	dot_green := assetsImage("assets/dotgreen.png")

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	running := true

	// Create a red text with the font
	paddle := make([]*sdl.Surface, 12)
	tpaddle := []string{"12", "11", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1"}
	colorText := sdl.Color{G: 255, R: 255, B: 255, A: 255}
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

	offset := 0

	go func() {
		expected_min := -1

		fullDraw := func(minutes, seconds int) {
			surface.FillRect(nil, 0)

			window.SetTitle(formatTime(minutes))

			secBack.Blit(nil, surface, &sdl.Rect{
				X: 110,
				Y: 110,
				W: 0, H: 0})

			hour := minutes / 60
			highligth := (hour % 12)
			if highligth == 0 {
				highligth = 12
			}

			// lines
			for j := range 12 {
				drawDial(surface, dot, minutes, nil, j)
			}

			// numbers
			for j := range 12 {
				d_dot := Iif(highligth == (12-j), dot_green, dot)
				drawDial(surface, d_dot, minutes, paddle[j], j)
			}

			quadrant.Blit(nil, surface, &sdl.Rect{
				X: 0,
				Y: 0,
				W: 0, H: 0})

			// second dial and central hinge
			{
				a := float64(seconds*6-90) / 180.0 * math.Pi
				x := int(120.0*math.Cos(a)) + 300
				y := int(120.0*math.Sin(a)) + 300

				a += math.Pi / 2.0
				xa := int(5.0*math.Cos(a)) + 300
				ya := int(5.0*math.Sin(a)) + 300

				a -= math.Pi
				xb := int(5.0*math.Cos(a)) + 300
				yb := int(5.0*math.Sin(a)) + 300

				drawLine(surface, 300, 300, xa, ya)
				drawLine(surface, 300, 300, xb, yb)
				drawLine(surface, 300, 300, x, y)
				drawLine(surface, xa, ya, x, y)
				drawLine(surface, xb, yb, x, y)
			}
			drawDot(surface, 300, 300, 20, dot)

			//window.UpdateSurface()
			update <- true
		}

		for {
			if !running {
				return
			}
			minutes, seconds := getTime()

			minutes += offset
			minutes %= 1440

			if expected_min == -1 {
				expected_min = minutes
			}

			for expected_min != minutes {
				expected_min++
				expected_min %= 1440
				_, seconds = getTime()
				fullDraw(expected_min, seconds)
				<-time.After(10 * time.Millisecond)
			}
			//fmt.Printf("Offset %d\n", offset)

			fullDraw(minutes, seconds)
			expected_min = minutes - 1

			if debug {
				<-time.After(100 * time.Millisecond)
			} else {
				<-time.After(1000 * time.Millisecond)
			}
		}
	}()

	for running {
		select {
		case <-time.After(100 * time.Millisecond):

		case <-update:
			window.UpdateSurface()
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {

			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.OSEvent:
				fmt.Printf("OS event 0x%x\n", t.Type)

			case *sdl.RenderEvent:
				fmt.Printf("Render event 0x%x\n", t.Type)

			case *sdl.DisplayEvent:
				fmt.Printf("Display event 0x%x\n", t.Type)
				switch t.Type {
				case sdl.WINDOWEVENT_SHOWN:
					fmt.Printf("SHOW\n")
					go func() {
						update <- true
					}()
				}

			case *sdl.KeyboardEvent:
				if t.State == sdl.RELEASED {
					switch t.Keysym.Sym {
					case sdl.K_KP_ENTER:
						offset += 30
						fmt.Printf("OFFSET\n")

					case sdl.K_KP_0:
						offset = 0

					case sdl.K_ESCAPE:
						running = false
					}
				}
			}
		}
	}
}
