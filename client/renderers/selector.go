package renderers

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var selectorProg uint32
var selectorVao uint32

var selectorPUniform int32
var selectorVUniform int32
var selectorMUniform int32

func initSelector() {
	// Compile shaders
	selectorVert, err := compileShader(`
#version 410

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;

layout (location = 0) in vec3 vp;

void main() {
	gl_Position = projection * view * model * vec4(vp, 1.0);
}`+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	selectorFrag, err := compileShader(`
#version 410

out vec4 frag_colour;

void main() {
	frag_colour = vec4(0.0, 0.0, 0.0, 1.0);
}`+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	selectorProg = gl.CreateProgram()
	gl.AttachShader(selectorProg, selectorVert)
	gl.AttachShader(selectorProg, selectorFrag)
	gl.LinkProgram(selectorProg)

	// Make selector VAO
	verts := GlBufferFrom([]float32{
		0, 0, 0,
		1, 0, 0,

		0, 0, 0,
		0, 1, 0,

		0, 0, 0,
		0, 0, 1,

		1, 1, 0,
		0, 1, 0,

		1, 1, 0,
		1, 0, 0,

		1, 1, 0,
		1, 1, 1,

		1, 0, 1,
		0, 0, 1,

		1, 0, 1,
		1, 1, 1,

		1, 0, 1,
		1, 0, 0,

		0, 1, 1,
		1, 1, 1,

		0, 1, 1,
		0, 0, 1,

		0, 1, 1,
		0, 1, 0,
	})

	gl.GenVertexArrays(1, &selectorVao)
	gl.BindVertexArray(selectorVao)

	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, verts)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil) // vec3

	// Bind uniforms
	selectorPUniform = gl.GetUniformLocation(selectorProg, gl.Str("projection\x00"))
	selectorVUniform = gl.GetUniformLocation(selectorProg, gl.Str("view\x00"))
	selectorMUniform = gl.GetUniformLocation(selectorProg, gl.Str("model\x00"))
}

func RenderSelector(pos mgl32.Vec3, view mgl32.Mat4) {
	gl.UseProgram(selectorProg)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Assign view & projection mats
	projection := mgl32.Perspective(mgl32.DegToRad(FOVDegrees), GetAspectRatio(), 0.1, 300.0)
	gl.UniformMatrix4fv(selectorPUniform, 1, false, &projection[0])
	gl.UniformMatrix4fv(selectorVUniform, 1, false, &view[0])

	// Translate selector into position
	model := mgl32.Translate3D(pos[0], pos[1], pos[2])
	gl.UniformMatrix4fv(selectorMUniform, 1, false, &model[0])

	// Draw
	gl.BindVertexArray(selectorVao)
	gl.DrawArrays(gl.LINES, 0, 12*2)
}
