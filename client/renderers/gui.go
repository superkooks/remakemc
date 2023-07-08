package renderers

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var guiProg uint32

type GUIElem struct {
	VAO       uint32
	Tex       uint32
	VertCount int
}

var tintVAO uint32

func initGUI() {
	// Compile shaders
	guiVert, err := compileShader(`
#version 410

layout (location = 0) in vec3 vp;
layout (location = 1) in vec2 uv;
uniform vec2 modelStart;
uniform vec2 modelEnd;
out vec2 fragUV;

void main() {
	vec2 box = modelEnd - modelStart;
	gl_Position.x = modelStart.x + vp.x*box.x;
	gl_Position.y = modelStart.y + vp.y*box.y;
	gl_Position.z = vp.z;
	gl_Position.w = 1.0;
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

	// Create tint VAO
	verts := GlBufferFrom([]float32{
		-1, -1, 1,
		1, -1, 1,
		1, 1, 1,
		1, 1, 1,
		-1, 1, 1,
		-1, -1, 1,
	})

	gl.GenVertexArrays(1, &tintVAO)
	gl.BindVertexArray(tintVAO)

	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, verts)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil) // vec3
}

func RenderGUIElement(e GUIElem, start, end mgl32.Vec2) {
	gl.UseProgram(guiProg)

	// Blend transparency and disable depth test
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.DEPTH_TEST)

	// Enable texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, e.Tex)

	// Set model uniforms
	gl.Uniform2fv(gl.GetUniformLocation(guiProg, gl.Str("modelStart\x00")), 1, &start[0])
	gl.Uniform2fv(gl.GetUniformLocation(guiProg, gl.Str("modelEnd\x00")), 1, &end[0])

	// Draw
	gl.BindVertexArray(e.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(e.VertCount))
}

func TintScreen(colorIn mgl32.Vec4) {
	s := compiledShaders["mc:tint_screen"]
	gl.UseProgram(s.Program)

	// Blend transparency and disable depth test
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.DEPTH_TEST)

	// Set model uniform to color
	gl.Uniform4fv(s.Uniforms["color_in"], 1, &colorIn[0])

	// Draw
	gl.BindVertexArray(tintVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}
