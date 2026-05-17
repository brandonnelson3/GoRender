package gfx

import (
	"log"
	"strconv"
	"strings"

	"github.com/brandonnelson3/GoRender/messagebus"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	fpsProgram      uint32
	fpsVao          uint32
	fpsVbo          uint32
	fpsProjLoc      int32
	fpsTransLoc     int32
	fpsScaleLoc     int32
	fpsColorLoc     int32

	frameCount      int
	lastUpdateTime  float64
	lastLogTime     float64
	currentFPS      int

	fpsEnabled      bool = false
	fpsInitialized  bool
)

const fpsVertexShader = `
#version 450
uniform mat4 projection;
uniform vec2 translation;
uniform vec2 scale;

in vec2 position;

void main() {
    gl_Position = projection * vec4(position * scale + translation, 0.0, 1.0);
}
` + "\x00"

const fpsFragmentShader = `
#version 450
uniform vec3 color;
out vec4 outColor;

void main() {
    outColor = vec4(color, 1.0);
}
` + "\x00"

// init registers the keyboard listener early during package initialization.
func init() {
	messagebus.RegisterType("key", func(m *messagebus.Message) {
		pressedThisFrame := m.Data2.([]glfw.Key)
		for _, key := range pressedThisFrame {
			if key == glfw.KeyF {
				fpsEnabled = !fpsEnabled
				log.Printf("[HUD Overlay] Toggled FPS counter via F key. Active: %v\n", fpsEnabled)
			}
		}
	})
}

// InitFPS compiles the FPS overlay shader and sets up VAO/VBO handles.
func InitFPS() {
	if fpsInitialized {
		return
	}

	// 1. Compile vertex shader
	vs := gl.CreateShader(gl.VERTEX_SHADER)
	vsSrc, freeVs := gl.Strs(fpsVertexShader)
	gl.ShaderSource(vs, 1, vsSrc, nil)
	freeVs()
	gl.CompileShader(vs)
	var status int32
	gl.GetShaderiv(vs, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetShaderiv(vs, gl.INFO_LOG_LENGTH, &logLen)
		infoLog := strings.Repeat("\x00", int(logLen+1))
		gl.GetShaderInfoLog(vs, logLen, nil, gl.Str(infoLog))
		panic("Failed to compile FPS vertex shader: " + infoLog)
	}

	// 2. Compile fragment shader
	fs := gl.CreateShader(gl.FRAGMENT_SHADER)
	fsSrc, freeFs := gl.Strs(fpsFragmentShader)
	gl.ShaderSource(fs, 1, fsSrc, nil)
	freeFs()
	gl.CompileShader(fs)
	gl.GetShaderiv(fs, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetShaderiv(fs, gl.INFO_LOG_LENGTH, &logLen)
		infoLog := strings.Repeat("\x00", int(logLen+1))
		gl.GetShaderInfoLog(fs, logLen, nil, gl.Str(infoLog))
		panic("Failed to compile FPS fragment shader: " + infoLog)
	}

	// 3. Link program
	fpsProgram = gl.CreateProgram()
	gl.AttachShader(fpsProgram, vs)
	gl.AttachShader(fpsProgram, fs)
	gl.LinkProgram(fpsProgram)
	gl.GetProgramiv(fpsProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetProgramiv(fpsProgram, gl.INFO_LOG_LENGTH, &logLen)
		infoLog := strings.Repeat("\x00", int(logLen+1))
		gl.GetProgramInfoLog(fpsProgram, logLen, nil, gl.Str(infoLog))
		panic("Failed to link FPS program: " + infoLog)
	}

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)

	// Retrieve uniform locations
	fpsProjLoc = gl.GetUniformLocation(fpsProgram, gl.Str("projection\x00"))
	fpsTransLoc = gl.GetUniformLocation(fpsProgram, gl.Str("translation\x00"))
	fpsScaleLoc = gl.GetUniformLocation(fpsProgram, gl.Str("scale\x00"))
	fpsColorLoc = gl.GetUniformLocation(fpsProgram, gl.Str("color\x00"))

	// 4. Generate quad geometry
	vertices := []float32{
		0.0, 0.0,
		1.0, 0.0,
		1.0, 1.0,
		0.0, 0.0,
		1.0, 1.0,
		0.0, 1.0,
	}

	gl.GenVertexArrays(1, &fpsVao)
	gl.BindVertexArray(fpsVao)

	gl.GenBuffers(1, &fpsVbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, fpsVbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	posAttrib := uint32(gl.GetAttribLocation(fpsProgram, gl.Str("position\x00")))
	gl.EnableVertexAttribArray(posAttrib)
	gl.VertexAttribPointer(posAttrib, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))

	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	lastUpdateTime = glfw.GetTime()
	lastLogTime = glfw.GetTime()
	fpsInitialized = true
}

// RenderFPS calculates instant FPS and renders it in the upper-right corner of the window.
func RenderFPS() {
	if !fpsEnabled || RenderTestMode {
		return
	}

	if !fpsInitialized {
		InitFPS()
	}

	// Calculate FPS
	frameCount++
	now := glfw.GetTime()
	elapsed := now - lastUpdateTime
	if elapsed >= 0.05 { // Update every 50ms
		currentFPS = int(float64(frameCount) / elapsed)
		frameCount = 0
		lastUpdateTime = now
	}

	// Print to console every 1 second for fallback diagnostics
	if now-lastLogTime >= 1.0 {
		log.Printf("[HUD Overlay] Current FPS: %d\n", currentFPS)
		lastLogTime = now
	}

	// Draw FPS
	var cullingEnabled bool
	gl.GetBooleanv(gl.CULL_FACE, &cullingEnabled)

	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.DEPTH_TEST)
	gl.UseProgram(fpsProgram)

	// Set orthographic projection for the window size
	projection := mgl32.Ortho(0.0, float32(Window.Width), float32(Window.Height), 0.0, -1.0, 1.0)
	gl.UniformMatrix4fv(fpsProjLoc, 1, false, &projection[0])

	// Sleek neon green/yellow color
	gl.Uniform3f(fpsColorLoc, 0.2, 1.0, 0.2)

	fpsStr := strconv.Itoa(currentFPS)

	const (
		padding      = float32(20.0)
		digitWidth   = float32(18.0)
		digitHeight  = float32(30.0)
		thickness    = float32(3.0)
		digitSpacing = float32(25.0)
	)

	// Segment mappings for 7-segment display: A, B, C, D, E, F, G
	var segments = [10][7]bool{
		{true, true, true, true, true, true, false},     // 0
		{false, true, true, false, false, false, false}, // 1
		{true, true, false, true, true, false, true},    // 2
		{true, true, true, true, false, false, true},    // 3
		{false, true, true, false, false, true, true},   // 4
		{true, false, true, true, false, true, true},    // 5
		{true, false, true, true, true, true, true},     // 6
		{true, true, true, false, false, false, false},  // 7
		{true, true, true, true, true, true, true},      // 8
		{true, true, true, true, false, true, true},     // 9
	}

	gl.BindVertexArray(fpsVao)

	for i, char := range fpsStr {
		digitVal := int(char - '0')
		if digitVal < 0 || digitVal > 9 {
			continue
		}

		// Calculate top-left of this digit (right-aligned layout)
		dx := float32(Window.Width) - padding - float32(len(fpsStr)-i)*digitSpacing
		dy := padding

		activeSegments := segments[digitVal]

		drawSeg := func(sx, sy, sw, sh float32) {
			gl.Uniform2f(fpsTransLoc, dx+sx, dy+sy)
			gl.Uniform2f(fpsScaleLoc, sw, sh)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)
		}

		// Segment A (top horizontal)
		if activeSegments[0] {
			drawSeg(thickness, 0, digitWidth-2*thickness, thickness)
		}
		// Segment B (top-right vertical)
		if activeSegments[1] {
			drawSeg(digitWidth-thickness, thickness, thickness, digitHeight/2-thickness)
		}
		// Segment C (bottom-right vertical)
		if activeSegments[2] {
			drawSeg(digitWidth-thickness, digitHeight/2, thickness, digitHeight/2-thickness)
		}
		// Segment D (bottom horizontal)
		if activeSegments[3] {
			drawSeg(thickness, digitHeight-thickness, digitWidth-2*thickness, thickness)
		}
		// Segment E (bottom-left vertical)
		if activeSegments[4] {
			drawSeg(0, digitHeight/2, thickness, digitHeight/2-thickness)
		}
		// Segment F (top-left vertical)
		if activeSegments[5] {
			drawSeg(0, thickness, thickness, digitHeight/2-thickness)
		}
		// Segment G (middle horizontal)
		if activeSegments[6] {
			drawSeg(thickness, digitHeight/2-thickness/2, digitWidth-2*thickness, thickness)
		}
	}

	gl.BindVertexArray(0)
	
	if cullingEnabled {
		gl.Enable(gl.CULL_FACE)
	}
	gl.Enable(gl.DEPTH_TEST)
}
