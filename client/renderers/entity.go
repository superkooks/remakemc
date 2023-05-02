package renderers

import (
	"remakemc/core"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var ReusableShaders = make(map[string](func() *EntityShader))
var compiledShaders = make(map[string]*EntityShader)

type EntityShader struct {
	Program  uint32
	Uniforms map[string]int32
}

func initEntity() {
	initEntityShaders()

	for k, v := range ReusableShaders {
		compiledShaders[k] = v()
	}

	for _, v := range core.EntityRegistry {
		v.RenderType.Init()
	}
}

type TestEntityRenderer struct {
	Vertices []float32
	Shader   string // index into ReusableShaders

	eshader *EntityShader
	vao     uint32
}

func (d *TestEntityRenderer) Init() {
	d.eshader = compiledShaders[d.Shader]

	// Make entity VAO
	verts := GlBufferFrom(d.Vertices)
	normals := GlBufferFrom(MakeNormals(d.Vertices))

	gl.GenVertexArrays(1, &d.vao)
	gl.BindVertexArray(d.vao)

	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, verts)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil) // vec3

	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, normals)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil) // vec3
}

func (d *TestEntityRenderer) RenderEntity(e *core.Entity, view mgl32.Mat4) {
	gl.UseProgram(d.eshader.Program)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)

	// Assign view & projection mats
	projection := mgl32.Perspective(mgl32.DegToRad(FOVDegrees), GetAspectRatio(), 0.05, 1024.0)
	gl.UniformMatrix4fv(d.eshader.Uniforms["projection"], 1, false, &projection[0])
	gl.UniformMatrix4fv(d.eshader.Uniforms["view"], 1, false, &view[0])

	// Translate entity into position
	model := mgl32.Translate3D(e.Position[0], e.Position[1], e.Position[2])
	gl.UniformMatrix4fv(d.eshader.Uniforms["model"], 1, false, &model[0])

	// Draw
	gl.BindVertexArray(d.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(d.Vertices)))
}
