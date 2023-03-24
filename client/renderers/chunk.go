package renderers

import (
	"remakemc/core"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var chunkProg uint32
var pUniform int32
var vUniform int32
var mUniform int32

func initChunk() {
	// Compile shaders
	guiVert, err := compileShader(`
#version 410

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;

layout (location = 0) in vec3 vp;
layout (location = 1) in vec3 vertexNormal;

out vec3 fragNormal;
out vec3 fragVertex;
out mat4 fragModel;

void main() {
	gl_Position = projection * view * model * vec4(vp, 1.0);

	fragVertex = vp;
	fragNormal = vertexNormal;
	fragModel = model;
}`+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	guiFrag, err := compileShader(`
#version 410

uniform vec3 cameraPosition;

in vec3 fragNormal;
in vec3 fragVertex;
in mat4 fragModel;

out vec4 color;

vec3 ApplyLight(vec3 surfaceColor, vec3 normal, vec3 surfacePos, vec3 surfaceToCamera) {
    // Directional light
    vec3 surfaceToLight = normalize(vec3(0.4, 0.9, 0.2));
    vec3 lightIntensity = vec3(1.0, 1.0, 1.0);

    // Ambient
    vec3 ambient = 0.36f * surfaceColor.rgb * lightIntensity;

    // Diffuse
    float diffuseCoefficient = max(0.0, dot(normal, surfaceToLight));
    vec3 diffuse = diffuseCoefficient * surfaceColor.rgb * lightIntensity;

    // Linear color (color before gamma correction)
    return ambient + diffuse;
}

void main() {
	vec3 normal = normalize(transpose(inverse(mat3(fragModel))) * fragNormal);
    vec3 surfacePos = vec3((fragModel * vec4(fragVertex, 1)).xyz);
    // vec4 surfaceColor = texture(fragmentTexture, UV);
	vec4 surfaceColor = vec4(1.0);
    vec3 surfaceToCamera = normalize(cameraPosition - surfacePos);

    // Combine color from all the lights
    vec3 linearColor = ApplyLight(surfaceColor.rgb, normal, surfacePos, surfaceToCamera);
    
    // Final color (after gamma correction)
    vec3 gamma = vec3(1.0/1.2);
    color = vec4(pow(linearColor, gamma), surfaceColor.a);
}`+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	chunkProg = gl.CreateProgram()
	gl.AttachShader(chunkProg, guiVert)
	gl.AttachShader(chunkProg, guiFrag)
	gl.LinkProgram(chunkProg)

	pUniform = gl.GetUniformLocation(chunkProg, gl.Str("projection\x00"))
	vUniform = gl.GetUniformLocation(chunkProg, gl.Str("view\x00"))
	mUniform = gl.GetUniformLocation(chunkProg, gl.Str("model\x00"))
}

func RenderChunk(c *core.Chunk, view mgl32.Mat4, aspectRatio float32) {
	gl.UseProgram(chunkProg)
	gl.Enable(gl.CULL_FACE)

	// Blend transparency
	// gl.Enable(gl.BLEND)
	// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Assign view & projection mats
	projection := mgl32.Perspective(mgl32.DegToRad(float32(70)), aspectRatio, 0.1, 300.0)
	gl.UniformMatrix4fv(pUniform, 1, false, &projection[0])
	gl.UniformMatrix4fv(vUniform, 1, false, &view[0])

	// Translate chunk into position
	p := c.Position.ToFloat()
	model := mgl32.Translate3D(p[0], p[1], p[2])
	gl.UniformMatrix4fv(mUniform, 1, false, &model[0])

	// Draw
	gl.EnableVertexAttribArray(0)
	gl.BindVertexArray(c.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(c.Mesh)))
}

func MakeChunkVAO(d *core.Dimension, chunk *core.Chunk) {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var normals []float32
	chunk.Mesh, normals = MakeChunkMesh(d, chunk.Position)

	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, GlBufferFrom(chunk.Mesh))
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil) // vec3

	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, GlBufferFrom(normals))
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil) // vec3

	chunk.VAO = vao
}

func MakeChunkMesh(d *core.Dimension, chunkPos core.Vec3) (verts, normals []float32) {
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				local := core.NewVec3(x, y, z)
				global := chunkPos.Add(local)
				b := d.GetBlockAt(global)
				if b.Type == nil {
					continue
				}

				// Top
				if d.GetBlockAt(global.Add(core.Vec3{Y: 1})).Type == nil {
					newVert := makeFace(topFace, local.ToFloat())
					verts = append(verts, newVert...)
					normals = append(normals, makeNormals(newVert)...)
				}

				// Bottom
				if d.GetBlockAt(global.Add(core.Vec3{Y: -1})).Type == nil {
					newVert := makeFace(bottomFace, local.ToFloat())
					verts = append(verts, newVert...)
					normals = append(normals, makeNormals(newVert)...)
				}

				// Left
				if d.GetBlockAt(global.Add(core.Vec3{X: -1})).Type == nil {
					newVert := makeFace(leftFace, local.ToFloat())
					verts = append(verts, newVert...)
					normals = append(normals, makeNormals(newVert)...)
				}

				// Right
				if d.GetBlockAt(global.Add(core.Vec3{X: 1})).Type == nil {
					newVert := makeFace(rightFace, local.ToFloat())
					verts = append(verts, newVert...)
					normals = append(normals, makeNormals(newVert)...)
				}

				// Front
				if d.GetBlockAt(global.Add(core.Vec3{Z: 1})).Type == nil {
					newVert := makeFace(frontFace, local.ToFloat())
					verts = append(verts, newVert...)
					normals = append(normals, makeNormals(newVert)...)
				}

				// Back
				if d.GetBlockAt(global.Add(core.Vec3{Z: -1})).Type == nil {
					newVert := makeFace(backFace, local.ToFloat())
					verts = append(verts, newVert...)
					normals = append(normals, makeNormals(newVert)...)
				}
			}
		}
	}

	return
}

func makeFace(face []float32, pos mgl32.Vec3) []float32 {
	newV := make([]float32, 3*6)
	copy(newV, face)
	for i := 0; i < 6; i++ {
		newV[i*3] += pos.X()
		newV[i*3+1] += pos.Y()
		newV[i*3+2] += pos.Z()
	}

	return newV
}

func makeNormals(newV []float32) []float32 {
	normals := make([]float32, 18)
	for i := 0; i < 2; i++ {
		vecV := mgl32.Vec3{newV[3+i*9] - newV[0+i*9], newV[4+i*9] - newV[1+i*9], newV[5+i*9] - newV[2+i*9]}
		vecW := mgl32.Vec3{newV[6+i*9] - newV[0+i*9], newV[7+i*9] - newV[1+i*9], newV[8+i*9] - newV[2+i*9]}
		n := [3]float32(vecV.Cross(vecW))

		copy(normals[0+i*9:], n[:])
		copy(normals[3+i*9:], n[:])
		copy(normals[6+i*9:], n[:])
	}

	return normals
}

// Vertices for faces
var topFace = []float32{
	0, 1, 0,
	0, 1, 1,
	1, 1, 0,

	1, 1, 0,
	0, 1, 1,
	1, 1, 1,
}

var bottomFace = []float32{
	0, 0, 0,
	1, 0, 0,
	0, 0, 1,

	1, 0, 0,
	1, 0, 1,
	0, 0, 1,
}

var leftFace = []float32{
	0, 0, 1,
	0, 1, 0,
	0, 0, 0,

	0, 0, 1,
	0, 1, 1,
	0, 1, 0,
}

var rightFace = []float32{
	1, 0, 1,
	1, 0, 0,
	1, 1, 0,

	1, 0, 1,
	1, 1, 0,
	1, 1, 1,
}

var frontFace = []float32{
	0, 0, 1,
	1, 0, 1,
	0, 1, 1,

	1, 0, 1,
	1, 1, 1,
	0, 1, 1,
}

var backFace = []float32{
	0, 0, 0,
	0, 1, 0,
	1, 0, 0,

	1, 0, 0,
	0, 1, 0,
	1, 1, 0,
}

// UV maps for faces
var topFaceUV = []float32{
	0, 0, 9,
	0, 1, 9,
	1, 0, 9,

	1, 0, 9,
	0, 1, 9,
	1, 1, 9,
}

var bottomFaceUV = []float32{
	0, 0, 9,
	1, 0, 9,
	0, 1, 9,

	1, 0, 9,
	1, 1, 9,
	0, 1, 9,
}

var leftFaceUV = []float32{
	0, 1, 9,
	1, 0, 9,
	0, 0, 9,

	0, 1, 9,
	1, 1, 9,
	1, 0, 9,
}

var rightFaceUV = []float32{
	1, 1, 9,
	1, 0, 9,
	0, 0, 9,

	1, 1, 9,
	0, 0, 9,
	0, 1, 9,
}

var frontFaceUV = []float32{
	1, 0, 9,
	0, 0, 9,
	1, 1, 9,

	0, 0, 9,
	0, 1, 9,
	1, 1, 9,
}

var backFaceUV = []float32{
	0, 0, 9,
	0, 1, 9,
	1, 0, 9,

	1, 0, 9,
	0, 1, 9,
	1, 1, 9,
}
