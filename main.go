package main

import (
	"log"
	"runtime"

	"github.com/brandonnelson3/GoRender/console"
	"github.com/brandonnelson3/GoRender/gfx"
	"github.com/brandonnelson3/GoRender/input"
	"github.com/brandonnelson3/GoRender/terrain"

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
	windowFar    = 10000
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

	obj, err := gfx.LoadObjFile("assets/Tree.obj")
	if err != nil {
		panic(err)
	}
	objRenderable := obj.GetChunkedRenderable()
	objRenderable.Scale = mgl32.Scale3D(2, 2, 2)

	sky, err := gfx.NewSky()
	if err != nil {
		panic(err)
	}

	terr := terrain.NewTerrain()

	renderables := []gfx.Renderable{terr}
	updateables := []gfx.Updateable{terr}

	for x := 0; x <= 4; x++ {
		for z := 0; z <= 4; z++ {
			r := objRenderable.Copy()

			height := terr.GetHeight(float32(x*8+5), float32(z*8+5))

			r.Position = mgl32.Vec3{float32(x * 8+5), height, float32(z * 8+5)}
			renderables = append(renderables, r)
		}
	}

	for !gfx.Window.ShouldClose() {
		StartOfFrame()

		input.Update()
		sky.Update()

		gfx.FirstPerson.Update(GetPreviousFrameLength())
		gfx.ThirdPerson.Update(GetPreviousFrameLength())

		gfx.Renderer.Render(sky, renderables)
		gfx.Renderer.Update(updateables)

		gfx.Window.SwapBuffers()
		glfw.PollEvents()
		gfx.Window.RecenterCursor()
		EndOfFrame()
	}
}
