package renderers

import (
	"fmt"
	"remakemc/core"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type ItemFromBlock struct {
	Block string

	verts []float32
	vao   uint32
}

func initItems() {
	for _, v := range core.ItemRegistry {
		v.RenderType.Init()
	}
}

func (t *ItemFromBlock) Init() {
	// Generate vertices for block
	b := core.BlockRegistry[t.Block]

	vert1, norm1, uv1 := b.RenderType.RenderFace(core.FaceTop, mgl32.Vec3{0, 0, 0})
	vert2, norm2, uv2 := b.RenderType.RenderFace(core.FaceLeft, mgl32.Vec3{0, 0, 0})
	vert3, norm3, uv3 := b.RenderType.RenderFace(core.FaceFront, mgl32.Vec3{0, 0, 0})

	// vert4, norm4, uv4 := b.RenderType.RenderFace(core.FaceTop, mgl32.Vec3{0, 0, 0})
	// vert5, norm5, uv5 := b.RenderType.RenderFace(core.FaceLeft, mgl32.Vec3{0, 0, 0})
	// vert6, norm6, uv6 := b.RenderType.RenderFace(core.FaceFront, mgl32.Vec3{0, 0, 0})

	t.verts = append(vert1, append(vert2, vert3...)...)
	normals := append(norm1, append(norm2, norm3...)...)
	uv := append(uv1, append(uv2, uv3...)...)

	// t.verts = append(t.verts, append(vert4, append(vert5, vert6...)...)...)
	// normals = append(normals, append(norm4, append(norm5, norm6...)...)...)
	// uv = append(uv, append(uv4, append(uv5, uv6...)...)...)

	fmt.Println(t.verts)

	// Generate VAO for block
	gl.GenVertexArrays(1, &t.vao)
	gl.BindVertexArray(t.vao)

	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, GlBufferFrom(t.verts))
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil) // vec3

	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, GlBufferFrom(normals))
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil) // vec3

	gl.EnableVertexAttribArray(2)
	gl.BindBuffer(gl.ARRAY_BUFFER, GlBufferFrom(uv))
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 0, nil) // vec2
}

func (t *ItemFromBlock) RenderItem(i *core.ItemType, start mgl32.Vec2, end mgl32.Vec2) {
	shader := compiledShaders["mc:item_from_block"]
	gl.UseProgram(shader.Program)
	gl.Disable(gl.DEPTH_TEST)

	// Assign view & projection mats
	projection := mgl32.Perspective(mgl32.DegToRad(70), 1, 0.05, 1024.0)
	gl.UniformMatrix4fv(shader.Uniforms["projection"], 1, false, &projection[0])

	view := mgl32.LookAtV(mgl32.Vec3{-4, 2.5, 4}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	gl.UniformMatrix4fv(shader.Uniforms["view"], 1, false, &view[0])

	// Translate block so it is centred
	model := mgl32.Translate3D(-0.5, -0.5, -0.5)
	gl.UniformMatrix4fv(shader.Uniforms["model"], 1, false, &model[0])

	// Scale into box
	gl.Uniform2fv(shader.Uniforms["modelStart"], 1, &start[0])
	gl.Uniform2fv(shader.Uniforms["modelEnd"], 1, &end[0])

	// Select texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, chunkTex)
	gl.Uniform1i(shader.Uniforms["tex"], int32(chunkTex-1))

	// Draw
	gl.BindVertexArray(t.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(t.verts)/3))
}
