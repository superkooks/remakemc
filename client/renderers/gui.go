package renderers

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

var guiProg uint32

type GUIElem struct {
	VAO       uint32
	Tex       uint32
	VertCount int
}

func initGUI() {
	// Compile shaders
	guiVert, err := compileShader(`
#version 410

layout (location = 0) in vec3 vp;
layout (location = 1) in vec2 uv;
uniform float aspectRatio;
out vec2 fragUV;

void main() {
	gl_Position = vec4(vp.x, vp.y*aspectRatio, vp.z, 1.0);
	fragUV = uv;
}`+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	guiFrag, err := compileShader(`
#version 410

in vec2 fragUV;
uniform sampler2D tex;
out vec4 frag_colour;

void main() {
	frag_colour = texture(tex, fragUV);
}`+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	guiProg = gl.CreateProgram()
	gl.AttachShader(guiProg, guiVert)
	gl.AttachShader(guiProg, guiFrag)
	gl.LinkProgram(guiProg)
}

func RenderGUIElement(e GUIElem) {
	gl.UseProgram(guiProg)

	// Blend transparency
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Enable texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, e.Tex)

	// Set aspect ratio uniform
	w, h := Win.GetSize()
	gl.Uniform1f(gl.GetUniformLocation(guiProg, gl.Str("aspectRatio\x00")), float32(w)/float32(h))

	// Draw
	gl.BindVertexArray(e.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(e.VertCount))
}
