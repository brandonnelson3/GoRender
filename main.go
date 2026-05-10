package main

import (
	"flag"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/brandonnelson3/GoRender/console"
	"github.com/brandonnelson3/GoRender/gfx"
	"github.com/brandonnelson3/GoRender/input"
	"github.com/brandonnelson3/GoRender/rendertest"
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

var (
	renderTestMode = flag.Bool("rendertest", false, "render all test scenes and exit (no interactive window)")
	renderScene    = flag.String("scene", "", "if set, only render this scene name (used with -rendertest)")
	renderOut      = flag.String("out", filepath.Join("rendertest", "testdata", "actual"), "output directory for render-test PNGs")
)

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	flag.Parse()

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	if *renderTestMode {
		runRenderTests()
		return
	}

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

			r.Position = mgl32.Vec3{float32(x*8 + 5), height, float32(z*8 + 5)}
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

// runRenderTests opens a hidden GL window, renders every test scene into an
// offscreen FBO, and writes PNGs to *renderOut.
func runRenderTests() {
	// Hidden window still gives us a valid GL 4.5 context on Windows.
	glfw.WindowHint(glfw.Visible, glfw.False)
	gfx.CreateWindow("rendertest", 640, 360, windowNear, windowFar, windowFOV)
	gfx.Window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatalf("gl.Init: %v", err)
	}

	gfx.InitRenderer()
	gfx.InitCameras()
	gfx.InitPointLights()
	gfx.InitDirectionalLights()
	gfx.InitPip()

	sky, err := gfx.NewSky()
	if err != nil {
		log.Fatalf("NewSky: %v", err)
	}

	if err := os.MkdirAll(*renderOut, 0755); err != nil {
		log.Fatalf("mkdir %s: %v", *renderOut, err)
	}

	for _, scene := range rendertest.All {
		if *renderScene != "" && scene.Name != *renderScene {
			continue
		}

		log.Printf("rendering scene: %s (%dx%d)", scene.Name, scene.Width, scene.Height)

		fbo, err := gfx.NewOffscreenFBO(scene.Width, scene.Height)
		if err != nil {
			log.Fatalf("NewOffscreenFBO for %s: %v", scene.Name, err)
		}

		// Temporarily resize Window to match the scene so that every internal
		// viewport/projection calculation (depth pass, light culling, etc.) is correct.
		gfx.Window.Resize(scene.Width, scene.Height)

		// Configure camera, lighting, objects for this scene.
		// Setup also returns the scene's renderables.
		renderables := scene.Setup()

		// Advance the camera one tick so shadow matrices are computed.
		gfx.FirstPerson.Update(0)

		// Update the sky vertices for the current camera orientation.
		sky.Update()

		// Route the renderer's color pass into the offscreen FBO.
		// Render() hardcodes gl.BindFramebuffer(0) for the normal pass —
		// TargetFramebuffer overrides that binding.
		gfx.Renderer.TargetFramebuffer = fbo.Handle()
		gfx.Renderer.Render(sky, renderables)
		gfx.Renderer.TargetFramebuffer = 0

		// Flush so all GPU commands are complete before reading.
		gl.Flush()

		// Read pixels and save as PNG.
		img := fbo.ReadPixels()
		outPath := filepath.Join(*renderOut, scene.Name+".png")
		f, err := os.Create(outPath)
		if err != nil {
			log.Fatalf("create %s: %v", outPath, err)
		}
		if err := png.Encode(f, img); err != nil {
			log.Fatalf("encode %s: %v", outPath, err)
		}
		f.Close()
		fbo.Delete()

		log.Printf("wrote: %s", outPath)
	}
}
