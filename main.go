package main

import (
	"log"
	"runtime"

	"github.com/brandonnelson3/GoRender/console"

	"github.com/brandonnelson3/GoRender/gfx"
	"github.com/brandonnelson3/GoRender/input"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	windowTitle  = "GoRender engine demo"
	windowWidth  = 1920
	windowHeight = 1080
	windowFOV    = 45.0
	windowNear   = .1
	windowFar    = 1000
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	gfx.CreateWindow(windowTitle, windowWidth, windowHeight, windowNear, windowFar, windowFOV)
	gfx.Window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	gfx.Window.SetKeyCallback(input.KeyCallBack)
	gfx.Window.SetMouseButtonCallback(input.MouseButtonCallback)
	gfx.Window.SetCursorPosCallback(input.CursorPosCallback)
	gfx.Window.MakeContextCurrent()
	gfx.Window.RecenterCursor()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	console.InitConsole()
	gfx.InitRenderer()
	gfx.InitCameras()
	gfx.InitPointLights()
	gfx.InitDirectionalLights()
	gfx.InitPip()

	renderables := []*gfx.Renderable{gfx.NewRenderable(gfx.PlaneVertices)}
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			for z := 0; z < 2; z++ {
				r := gfx.NewRenderable(gfx.CubeVertices)
				r.Rotation = &mgl32.Mat4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
				r.Scale = &mgl32.Mat4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
				r.Position = &mgl32.Vec3{float32(10 * x), float32(10 * y), float32(10 * z)}
				renderables = append(renderables, r)
			}
		}
	}

	for !gfx.Window.ShouldClose() {
		StartOfFrame()

		input.Update()
		gfx.FirstPerson.Update(GetPreviousFrameLength())
		gfx.ThirdPerson.Update(GetPreviousFrameLength())

		gfx.Renderer.Render(renderables)

		gfx.Window.SwapBuffers()
		glfw.PollEvents()
		gfx.Window.RecenterCursor()
		EndOfFrame()
	}
}
