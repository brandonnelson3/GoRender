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

	obj, err := gfx.LoadObjFile("assets/Tree.obj")
	if err != nil {
		panic(err)
	}
	objRenderable := obj.GetChunkedRenderable()
	objRenderable.Scale = mgl32.Scale3D(10, 10, 10)

	diffuseTexture, err := gfx.LoadTexture("assets/crate1_diffuse.png")
	if err != nil {
		panic(err)
	}

	renderables := []*gfx.Renderable{gfx.NewRenderable(gfx.PlaneVertices, diffuseTexture), objRenderable}

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
