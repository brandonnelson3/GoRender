package gfx

import (
	"log"

	"github.com/brandonnelson3/GoRender/gfx/shaders"

	"github.com/go-gl/gl/v4.5-core/gl"
)

const (
	TileSize = 16
)

var (
	Renderer r
)

type r struct {
	depthShaderPipeline uint32
	depthVertexShader   *shaders.DepthVertexShader
	depthFragmentShader *shaders.DepthFragmentShader

	lightCullingShader *shaders.LightCullingShader

	colorShaderPipeline uint32
	colorVertexShader   *shaders.ColorVertexShader
	colorFragmentShader *shaders.ColorFragmentShader

	depthMapFBO, depthMap uint32
}

func InitRenderer() {
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.MULTISAMPLE)
	gl.DepthMask(true)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	dvs, err := shaders.NewDepthVertexShader()
	if err != nil {
		log.Fatalf("Failed to compile DepthVertexShader: %v", err)
	}
	dfs, err := shaders.NewDepthFragmentShader()
	if err != nil {
		log.Fatalf("Failed to compile DepthVertexShader: %v", err)
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

	Renderer = r{
		depthShaderPipeline: dsp,
		depthVertexShader:   dvs,
		depthFragmentShader: dfs,
		lightCullingShader:  lcs,
		colorShaderPipeline: csp,
		colorVertexShader:   cvs,
		colorFragmentShader: cfs,
		depthMapFBO:         depthMapFBO,
		depthMap:            depthMap,
	}
}

// getNumTilesX returns back the number of tiles in each the X dimension that are needed for the current window size.
func getNumTilesX() uint32 {
	return uint32((Window.Width + TileSize - 1) / TileSize)
}

// getNumTilesY returns back the number of tiles in each the Y dimension that are needed for the current window size.
func getNumTilesY() uint32 {
	return uint32((Window.Height + TileSize - 1) / TileSize)
}

// getTotalNumTiles returns back the total number of tiles required to cover the entire screen.
func getTotalNumTiles() uint32 {
	return getNumTilesX() * getNumTilesY()
}

func (renderer *r) Render(renderables []*Renderable) {
	// Step 1: Depth Pass for pointlight culling
	/*gl.BindProgramPipeline(renderer.depthShaderPipeline)
	gl.BindFramebuffer(gl.FRAMEBUFFER, renderer.depthMapFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	renderer.depthVertexShader.View.Set(Active.GetView())
	renderer.depthVertexShader.Projection.Set(Window.GetProjection())
	for _, renderable := range renderables {
		renderer.depthVertexShader.Model.Set(renderable.GetModelMatrix())
		renderable.Render()
	}

	// Step 2: Light Culling
	renderer.lightCullingShader.Use()
	renderer.lightCullingShader.View.Set(Active.GetView())
	renderer.lightCullingShader.Projection.Set(Window.GetProjection())
	renderer.lightCullingShader.DepthMap.Set(gl.TEXTURE4, 4, renderer.depthMap)
	renderer.lightCullingShader.ScreenSize.Set(uniforms.UIVec2{Window.Width, Window.Height})
	renderer.lightCullingShader.LightCount.Set(GetNumPointLights())
	renderer.lightCullingShader.LightBuffer.Set(GetPointLightBuffer())
	renderer.lightCullingShader.VisibleLightIndicesBuffer.Set(GetPointLightVisibleLightIndicesBuffer())
	gl.DispatchCompute(getNumTilesX(), getNumTilesY(), 1)
	gl.UseProgram(0)*/

	// Step 3: Normal pass
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindProgramPipeline(renderer.colorShaderPipeline)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	renderer.colorVertexShader.View.Set(Active.GetView())
	renderer.colorVertexShader.Projection.Set(Window.GetProjection())
	renderer.colorFragmentShader.NumTilesX.Set(getNumTilesX())
	renderer.colorFragmentShader.LightBuffer.Set(GetPointLightBuffer())
	renderer.colorFragmentShader.VisibleLightIndicesBuffer.Set(GetPointLightVisibleLightIndicesBuffer())
	renderer.colorFragmentShader.DirectionalLightBuffer.Set(GetDirectionalLightBuffer())
	//renderer.colorFragmentShader.Diffuse.Set(gl.TEXTURE0, 0, diffuseTexture)
	for _, renderable := range renderables {
		renderer.colorVertexShader.Model.Set(renderable.GetModelMatrix())
		// TODO: This should be set by key input only, not every frame. Right now UVs only.
		renderer.colorFragmentShader.RenderMode.Set(4)
		renderable.Render()
	}
}
