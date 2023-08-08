package renderers

import (
	"remakemc/core"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var ReusableShaders = make(map[string](func() *Shader))
var compiledShaders = make(map[string]*Shader)

type Shader struct {
	Program  uint32
	Uniforms map[string]int32
}

func initEntity() {
	for k, v := range ReusableShaders {
		compiledShaders[k] = v()
	}

	for _, v := range core.EntityRegistry {
		if r, ok := v.(core.RenderFace); ok {
			r.RenderInit()
		}
	}
}

type TestEntityRenderer struct {
	Vertices []float32
	Shader   string // index into ReusableShaders

	shader *Shader
	vao    uint32
}

func (d *TestEntityRenderer) Init() {
	d.shader = compiledShaders[d.Shader]

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

func (d *TestEntityRenderer) RenderEntity(e core.Entity, view mgl32.Mat4) {
	gl.UseProgram(d.shader.Program)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)

	// Assign view & projection mats
	projection := mgl32.Perspective(mgl32.DegToRad(FOVDegrees), GetAspectRatio(), 0.05, 1024.0)
	gl.UniformMatrix4fv(d.shader.Uniforms["projection"], 1, false, &projection[0])
	gl.UniformMatrix4fv(d.shader.Uniforms["view"], 1, false, &view[0])

	// Translate entity into position
	pos := e.(core.PositionFace).GetPosition()
	model := mgl32.Translate3D(pos[0], pos[1], pos[2])
	gl.UniformMatrix4fv(d.shader.Uniforms["model"], 1, false, &model[0])

	// Draw
	gl.BindVertexArray(d.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(d.Vertices)))
}
