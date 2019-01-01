package main

import (
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var window *sdl.Window
var windowRenderer *sdl.Renderer
var windowTexture *sdl.Texture
var buffer [w * h * 4]byte
var debugSurface *sdl.Surface
var debugRenderer *sdl.Renderer
var debugTexture *sdl.Texture

var system *System
var debug int
var framesRendered int
var fpsTimer time.Time

const debugNumScreens = 2
const scale = 2
const w = 256
const h = 240

var paused bool

func sdlInit() {
	var err error
	sdl.Init(sdl.INIT_EVERYTHING)

	window, windowRenderer, err = sdl.CreateWindowAndRenderer(w*scale, h*scale, 0)
	check(err)
	windowTexture, err = windowRenderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, w, h)
	check(err)

	debugSurface, err = sdl.CreateRGBSurface(0, w*scale, h*scale, 32, 0x00ff0000, 0x0000ff00, 0x000000ff, 0xff000000)
	check(err)
	debugRenderer, err = sdl.CreateSoftwareRenderer(debugSurface)
	check(err)
	debugTexture, err = windowRenderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, w*scale, h*scale)
	debugTexture.SetBlendMode(sdl.BLENDMODE_BLEND)

	debugRenderer.SetScale(scale, scale)
	fpsTimer = time.Now()
}

func sdlLoop() {
	var event sdl.Event
	running := true
	for running {
		frameStart := time.Now()

		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				pressed := t.Type == sdl.KEYDOWN
				switch t.Keysym.Scancode {
				case sdl.SCANCODE_RETURN:
					system.controller[0].buttons[ButtonStart] = pressed
				case sdl.SCANCODE_RSHIFT:
					system.controller[0].buttons[ButtonSelect] = pressed
				case sdl.SCANCODE_LEFT:
					system.controller[0].buttons[ButtonLeft] = pressed
				case sdl.SCANCODE_RIGHT:
					system.controller[0].buttons[ButtonRight] = pressed
				case sdl.SCANCODE_UP:
					system.controller[0].buttons[ButtonUp] = pressed
				case sdl.SCANCODE_DOWN:
					system.controller[0].buttons[ButtonDown] = pressed
				case sdl.SCANCODE_Z:
					system.controller[0].buttons[ButtonA] = pressed
				case sdl.SCANCODE_X:
					system.controller[0].buttons[ButtonB] = pressed
				case sdl.SCANCODE_GRAVE:
					if !pressed {
						debug = (debug + 1) % (debugNumScreens + 1)
					}
				case sdl.SCANCODE_SPACE:
					if !pressed {
						paused = !paused
					}
				}
			}
		}

		if !paused {
			system.EmulateFrame()
		}

		frameTime := time.Now().Sub(frameStart)
		delay := (16666667 - frameTime.Nanoseconds()) / 1000000
		if delay > 0 {
			sdl.Delay(uint32(delay))
		}

		framesRendered++
	}
}

func pushPixel(x int, y int, col uint32) {
	buffer[(y*w+x)*4+0] = byte((uint32(col) >> 0) & 0xFF)
	buffer[(y*w+x)*4+1] = byte((uint32(col) >> 8) & 0xFF)
	buffer[(y*w+x)*4+2] = byte((uint32(col) >> 16) & 0xFF)
	buffer[(y*w+x)*4+3] = byte((uint32(col) >> 24) & 0xFF)
}

func pushFrame() {
	windowTexture.Update(nil, buffer[:], 4*w)
	windowRenderer.Copy(windowTexture, nil, nil)
	windowRenderer.Present()
}

func sdlCleanup() {
	window.Destroy()
	sdl.Quit()
}

func startWithRom(romPath string) {
	system = NewSystem()
	system.ResetSystem(romPath)
	system.cpu.pc = system.cpu.getVectorReset()
	system.ppu.funcPushFrame = pushFrame
	system.ppu.funcPushPixel = pushPixel
	//Start emulating.
	system.EmulateFrame()

	sdlInit()
	sdlLoop()
	sdlCleanup()
}

func main() {
	//portaudio.Initialize()
	//defer portaudio.Terminate()
	romPath := "roms/Kirby's Adventure (E).nes"
	startWithRom(romPath)
	system = NewSystem()
	system.ResetSystem(romPath)
	//audio := NewAudio()
	//Start executing at the rom.
	//Start emulating.
	//audio.Start()
	//defer audio.Stop()
}
