package gfx

import (
	"log"

	"github.com/brandonnelson3/GoRender/gfx/shaders"
	"github.com/brandonnelson3/GoRender/gfx/uniforms"
	"github.com/brandonnelson3/GoRender/messagebus"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	// tileSize is the size in pixels of a tile for this renderer.
	tileSize = 16
	// shadowMapSize is the size of the square depth buffers used for CSM.
	shadowMapSize = 2048
)

// Renderer is the global instance of a Renderer.
var (
	Renderer r

	ambientLightColor = mgl32.Vec3{.2, .2, .2}
)

type r struct {
	lineShaderPipeline uint32
	lineVertexShader   *shaders.LineVertexShader
	lineFragmentShader *shaders.LineFragmentShader

	depthShaderPipeline uint32
	depthVertexShader   *shaders.DepthVertexShader
	depthFragmentShader *shaders.DepthFragmentShader

	lightCullingShader *shaders.LightCullingShader

	colorShaderPipeline uint32
	colorVertexShader   *shaders.ColorVertexShader
	colorFragmentShader *shaders.ColorFragmentShader

	csmDepthMapFBO uint32
	csmDepthMaps   [3]uint32

	depthMapFBO, depthMap uint32

	// TODO fix this and do something much better...
	diffuseTexture, sandTexture uint32
}

// InitRenderer instanciates the global Renderer instance.
func InitRenderer() {
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.MULTISAMPLE)
	gl.DepthMask(true)
	gl.PointSize(8.0)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	lvs, err := shaders.NewLineVertexShader()
	if err != nil {
		log.Fatalf("Failed to compile LineVertexShader: %v", err)
	}
	lfs, err := shaders.NewLineFragmentShader()
	if err != nil {
		log.Fatalf("Failed to compile LineFragmentShader: %v", err)
	}
	var lsp uint32
	gl.CreateProgramPipelines(1, &lsp)
	lvs.AddToPipeline(lsp)
	lfs.AddToPipeline(lsp)
	gl.ValidateProgramPipeline(lsp)
	lvs.BindVertexAttributes()

	dvs, err := shaders.NewDepthVertexShader()
	if err != nil {
		log.Fatalf("Failed to compile DepthVertexShader: %v", err)
	}
	dfs, err := shaders.NewDepthFragmentShader()
	if err != nil {
		log.Fatalf("Failed to compile DepthFragmentShader: %v", err)
	}
	var dsp uint32
	gl.CreateProgramPipelines(1, &dsp)
	dvs.AddToPipeline(dsp)
	dfs.AddToPipeline(dsp)
	gl.ValidateProgramPipeline(dsp)
	dvs.BindVertexAttributes()

	lcs, err := shaders.NewLightCullingShader()
	if err != nil {
		log.Fatalf("Failed to compile LightCullingShader: %v", err)
	}

	cvs, err := shaders.NewColorVertexShader()
	if err != nil {
		log.Fatalf("Failed to compile ColorVertexShader: %v", err)
	}
	cfs, err := shaders.NewColorFragmentShader()
	if err != nil {
		log.Fatalf("Failed to compile ColorFragmentShader: %v", err)
	}
	var csp uint32
	gl.CreateProgramPipelines(1, &csp)
	cvs.AddToPipeline(csp)
	cfs.AddToPipeline(csp)
	gl.ValidateProgramPipeline(csp)
	cvs.BindVertexAttributes()

	var depthMapFBO uint32
	gl.GenFramebuffers(1, &depthMapFBO)
	var depthMap uint32
	gl.GenTextures(1, &depthMap)
	gl.BindTexture(gl.TEXTURE_2D, depthMap)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT, int32(Window.Width), int32(Window.Height), 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
	borderColor := []float32{1.0, 1.0, 1.0, 1.0}
	gl.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR, &borderColor[0])
	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, depthMap, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Build CSM Depth FrameBuffers
	var csmDepthMapFBO uint32
	gl.GenFramebuffers(1, &csmDepthMapFBO)
	var csmDepthMaps [3]uint32
	gl.GenTextures(3, &csmDepthMaps[0])

	for _, m := range csmDepthMaps {
		gl.BindTexture(gl.TEXTURE_2D, m)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT, shadowMapSize, shadowMapSize, 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, csmDepthMapFBO)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, csmDepthMaps[0], 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	UpdatePip(&csmDepthMaps[0], Window.GetNearFar(0))

	messagebus.RegisterType("key", func(m *messagebus.Message) {
		pressedKeys := m.Data1.([]glfw.Key)
		for _, key := range pressedKeys {
			if key >= glfw.KeyF1 && key <= glfw.KeyF25 {
				cfs.RenderMode.Set(int32(key - glfw.KeyF1))
			}
		}
		pressedKeysThisFrame := m.Data2.([]glfw.Key)
		for _, key := range pressedKeysThisFrame {
			switch key {
			case glfw.KeyKP7:
				UpdatePip(&csmDepthMaps[0], Window.GetNearFar(0))
			case glfw.KeyKP8:
				UpdatePip(&csmDepthMaps[1], Window.GetNearFar(1))
			case glfw.KeyKP9:
				UpdatePip(&csmDepthMaps[2], Window.GetNearFar(2))
			}
		}
	})

	// TODO this should not be in renderer... probably should have some asset loading manager instead.
	diffuseTexture, err := NewFromPng("assets/crate1_diffuse.png")
	if err != nil {
		panic(err)
	}

	sandTexture, err := NewFromPng("assets/sand.png")
	if err != nil {
		panic(err)
	}

	Renderer = r{
		lineShaderPipeline:  lsp,
		lineVertexShader:    lvs,
		lineFragmentShader:  lfs,
		depthShaderPipeline: dsp,
		depthVertexShader:   dvs,
		depthFragmentShader: dfs,
		lightCullingShader:  lcs,
		colorShaderPipeline: csp,
		colorVertexShader:   cvs,
		colorFragmentShader: cfs,
		depthMapFBO:         depthMapFBO,
		depthMap:            depthMap,
		csmDepthMapFBO:      csmDepthMapFBO,
		csmDepthMaps:        csmDepthMaps,
		// TODO remove these...
		diffuseTexture: diffuseTexture,
		sandTexture:    sandTexture,
	}
}

// getNumTilesX returns back the number of tiles in each the X dimension that are needed for the current window size.
func getNumTilesX() uint32 {
	return uint32((Window.Width + tileSize - 1) / tileSize)
}

// getNumTilesY returns back the number of tiles in each the Y dimension that are needed for the current window size.
func getNumTilesY() uint32 {
	return uint32((Window.Height + tileSize - 1) / tileSize)
}

// getTotalNumTiles returns back the total number of tiles required to cover the entire screen.
func getTotalNumTiles() uint32 {
	return getNumTilesX() * getNumTilesY()
}

func (renderer *r) Render(renderables []*Renderable) {
	// Step 1: Depth Pass for each cascade for shadowing.
	gl.Viewport(0, 0, shadowMapSize, shadowMapSize)
	gl.BindProgramPipeline(renderer.depthShaderPipeline)
	renderer.depthVertexShader.View.Set(mgl32.Ident4())
	for i, m := range renderer.csmDepthMaps {
		gl.BindFramebuffer(gl.FRAMEBUFFER, renderer.csmDepthMapFBO)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, m, 0)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		renderer.depthVertexShader.Projection.Set(FirstPerson.shadowMatrices[i])
		for _, renderable := range renderables {
			renderer.depthVertexShader.Model.Set(renderable.GetModelMatrix())
			renderable.Render()
		}
	}

	// Step 2: Depth Pass for pointlight culling
	gl.Viewport(0, 0, int32(Window.Width), int32(Window.Height))
	gl.BindProgramPipeline(renderer.depthShaderPipeline)
	gl.BindFramebuffer(gl.FRAMEBUFFER, renderer.depthMapFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	renderer.depthVertexShader.View.Set(ActiveCamera.GetView())
	renderer.depthVertexShader.Projection.Set(Window.GetProjection())
	for _, renderable := range renderables {
		renderer.depthVertexShader.Model.Set(renderable.GetModelMatrix())
		renderable.Render()
	}

	// Step 3: Light culling
	renderer.lightCullingShader.Use()
	renderer.lightCullingShader.View.Set(ActiveCamera.GetView())
	renderer.lightCullingShader.Projection.Set(Window.GetProjection())
	renderer.lightCullingShader.DepthMap.Set(gl.TEXTURE4, 4, renderer.depthMap)
	renderer.lightCullingShader.ScreenSize.Set(uniforms.UIVec2{Window.Width, Window.Height})
	renderer.lightCullingShader.LightCount.Set(GetNumPointLights())
	renderer.lightCullingShader.LightBuffer.Set(GetPointLightBuffer())
	renderer.lightCullingShader.VisibleLightIndicesBuffer.Set(GetPointLightVisibleLightIndicesBuffer())
	gl.DispatchCompute(getNumTilesX(), getNumTilesY(), 1)
	gl.UseProgram(0)

	// Step 4: Normal pass
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindProgramPipeline(renderer.colorShaderPipeline)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	renderer.colorVertexShader.View.Set(ActiveCamera.GetView())
	renderer.colorVertexShader.Projection.Set(Window.GetProjection())
	renderer.colorVertexShader.LightViewProj1.Set(FirstPerson.shadowMatrices[0])
	renderer.colorVertexShader.LightViewProj2.Set(FirstPerson.shadowMatrices[1])
	renderer.colorVertexShader.LightViewProj3.Set(FirstPerson.shadowMatrices[2])
	renderer.colorFragmentShader.NumTilesX.Set(getNumTilesX())
	renderer.colorFragmentShader.LightBuffer.Set(GetPointLightBuffer())
	renderer.colorFragmentShader.ZNear.Set(Window.nearPlane)
	renderer.colorFragmentShader.ZFar.Set(Window.farPlane)
	renderer.colorFragmentShader.AmbientLightColor.Set(ambientLightColor)
	renderer.colorFragmentShader.VisibleLightIndicesBuffer.Set(GetPointLightVisibleLightIndicesBuffer())
	renderer.colorFragmentShader.DirectionalLightBuffer.Set(GetDirectionalLightBuffer())
	renderer.colorFragmentShader.Diffuse.Set(gl.TEXTURE0, 0, renderer.diffuseTexture)
	renderer.colorFragmentShader.ShadowMap1.Set(gl.TEXTURE1, 1, renderer.csmDepthMaps[0])
	renderer.colorFragmentShader.ShadowMap2.Set(gl.TEXTURE2, 2, renderer.csmDepthMaps[1])
	renderer.colorFragmentShader.ShadowMap3.Set(gl.TEXTURE3, 3, renderer.csmDepthMaps[2])
	for _, renderable := range renderables {
		renderer.colorVertexShader.Model.Set(renderable.GetModelMatrix())
		renderable.Render()
	}

	if ActiveCamera == ThirdPerson {
		gl.BindProgramPipeline(renderer.lineShaderPipeline)
		renderer.lineVertexShader.View.Set(ThirdPerson.GetView())
		renderer.lineVertexShader.Projection.Set(Window.GetProjection())
		FirstPerson.RenderFrustum()
	}

	RenderPip()
}
