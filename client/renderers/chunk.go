package renderers

import (
	"fmt"
	"remakemc/config"
	"remakemc/core"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var chunkProg uint32
var chunkPUniform int32
var chunkVUniform int32
var chunkMUniform int32
var chunkTUniform int32
var chunkTex uint32

func initChunk() {
	// Compile shaders
	guiVert, err := compileShader(`
#version 410

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;

layout (location = 0) in vec3 vp;
layout (location = 1) in vec3 vertexNormal;
layout (location = 2) in vec2 vertexUV;
layout (location = 3) in float vertexLL;

out vec3 fragNormal;
out vec3 fragVertex;
out vec2 fragUV;
out float fragLL;
out mat4 fragModel;

void main() {
	gl_Position = projection * view * model * vec4(vp, 1.0);

	fragVertex = vp;
	fragNormal = vertexNormal;
	fragUV = vertexUV;
	fragModel = model;
	fragLL = vertexLL;
}`+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	guiFrag, err := compileShader(`
#version 410

uniform vec3 cameraPosition;
uniform sampler2D fragTexture;

in vec3 fragNormal;
in vec3 fragVertex;
in vec2 fragUV;
in float fragLL;
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
	vec4 surfaceColor = texture(fragTexture, fragUV);
    vec3 surfaceToCamera = normalize(cameraPosition - surfacePos);

    // Combine color from all the lights
    vec3 linearColor = ApplyLight(surfaceColor.rgb, normal, surfacePos, surfaceToCamera) * fragLL;
    
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

	chunkPUniform = gl.GetUniformLocation(chunkProg, gl.Str("projection\x00"))
	chunkVUniform = gl.GetUniformLocation(chunkProg, gl.Str("view\x00"))
	chunkMUniform = gl.GetUniformLocation(chunkProg, gl.Str("model\x00"))
	chunkTUniform = gl.GetUniformLocation(chunkProg, gl.Str("fragTexture\x00"))

	// Load texture atlas and create texture
	i := BlockAtlas.Finalize()
	gl.GenTextures(1, &chunkTex)
	gl.BindTexture(gl.TEXTURE_2D, chunkTex)

	// Nearest when magnifying, linear when minifying
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// Send texture to GPU
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(i.Rect.Dx()), int32(i.Rect.Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(i.Pix))
	gl.GenerateMipmap(gl.TEXTURE_2D)
}

func RenderChunks(dim *core.Dimension, view mgl32.Mat4, playerPos mgl32.Vec3) {
	gl.UseProgram(chunkProg)
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Blend transparency
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Assign view & projection mats
	projection := mgl32.Perspective(mgl32.DegToRad(FOVDegrees), GetAspectRatio(), 0.05, 1024.0)
	gl.UniformMatrix4fv(chunkPUniform, 1, false, &projection[0])
	gl.UniformMatrix4fv(chunkVUniform, 1, false, &view[0])

	// Activate texture atlas
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, chunkTex)
	gl.Uniform1i(chunkTUniform, int32(chunkTex-1))

	i := 0
	chunkPos := core.NewVec3(
		core.FlooredDivision(core.FloorFloat32(playerPos.X()), 16)*16,
		0,
		core.FlooredDivision(core.FloorFloat32(playerPos.Z()), 16)*16,
	)
	for x := -config.App.RenderDistance * 16; x < config.App.RenderDistance*16; x += 16 {
		for y := 0; y < 256; y += 16 {
			for z := -config.App.RenderDistance * 16; z < config.App.RenderDistance*16; z += 16 {
				c := dim.Chunks[chunkPos.Add(core.NewVec3(x, y, z))]
				if c == nil {
					panic(fmt.Sprint("nil chunk at", chunkPos.Add(core.NewVec3(x, y, z))))
				}
				i++
				RenderChunk(c, view)
			}
		}
	}
}

func RenderChunk(c *core.Chunk, view mgl32.Mat4) {
	if c.MeshLen == 0 {
		if c.Position.Y == 0 {
			fmt.Println(c.Position)
			fmt.Println("skipping low chunk!!!")
		}
		return
	}

	// Translate chunk into position
	p := c.Position.ToFloat()
	model := mgl32.Translate3D(p[0], p[1], p[2])
	gl.UniformMatrix4fv(chunkMUniform, 1, false, &model[0])

	// Draw
	gl.BindVertexArray(c.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(c.MeshLen))
}

func MakeChunkMeshAndVAO(d *core.Dimension, chunk *core.Chunk) {
	mesh, normals, uvs, lightLevels := MakeChunkMesh(d, chunk.Position)
	MakeChunkVAO(chunk, mesh, normals, uvs, lightLevels)
}

func MakeChunkVAO(chunk *core.Chunk, mesh, normals, uvs, lightLevels []float32) {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	if len(mesh) == 0 {
		return
	}

	buf := GlBufferFrom(mesh)
	chunk.VertexBuffers = append(chunk.VertexBuffers, buf)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil) // vec3

	buf = GlBufferFrom(normals)
	chunk.VertexBuffers = append(chunk.VertexBuffers, buf)
	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil) // vec3

	buf = GlBufferFrom(uvs)
	chunk.VertexBuffers = append(chunk.VertexBuffers, buf)
	gl.EnableVertexAttribArray(2)
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 0, nil) // vec2

	buf = GlBufferFrom(lightLevels)
	chunk.VertexBuffers = append(chunk.VertexBuffers, buf)
	gl.EnableVertexAttribArray(3)
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.VertexAttribPointer(3, 1, gl.FLOAT, false, 0, nil) // float

	chunk.VAO = vao
	chunk.MeshLen = len(mesh)
}

func FreeChunk(c *core.Chunk) {
	gl.DeleteVertexArrays(1, &c.VAO)
	if len(c.VertexBuffers) > 0 {
		gl.DeleteBuffers(int32(len(c.VertexBuffers)), &c.VertexBuffers[0])
	}
}

func UpdateRequiredMeshes(dim *core.Dimension, updatePos core.Vec3) {
	MakeChunkMeshAndVAO(dim, dim.GetChunkContaining(updatePos))

	// X
	if core.FlooredRemainder(updatePos.X, 16) == 15 {
		for y := -1; y < 2; y++ {
			for z := -1; z < 2; z++ {
				chk := dim.GetChunkContaining(updatePos.Add(core.NewVec3(1, y, z)))
				if chk != nil {
					MakeChunkMeshAndVAO(dim, chk)
				}
			}
		}
	}
	if core.FlooredRemainder(updatePos.X, 16) == 0 {
		for y := -1; y < 2; y++ {
			for z := -1; z < 2; z++ {
				chk := dim.GetChunkContaining(updatePos.Add(core.NewVec3(-1, y, z)))
				if chk != nil {
					MakeChunkMeshAndVAO(dim, chk)
				}
			}
		}
	}

	// Y
	if core.FlooredRemainder(updatePos.Y, 16) == 15 {
		for x := -1; x < 2; x++ {
			for z := -1; z < 2; z++ {
				chk := dim.GetChunkContaining(updatePos.Add(core.NewVec3(x, 1, z)))
				if chk != nil {
					MakeChunkMeshAndVAO(dim, chk)
				}
			}
		}
	}
	if core.FlooredRemainder(updatePos.Y, 16) == 0 {
		for x := -1; x < 2; x++ {
			for z := -1; z < 2; z++ {
				chk := dim.GetChunkContaining(updatePos.Add(core.NewVec3(x, -1, z)))
				if chk != nil {
					MakeChunkMeshAndVAO(dim, chk)
				}
			}
		}
	}

	// Z
	if core.FlooredRemainder(updatePos.Z, 16) == 15 {
		for x := -1; x < 2; x++ {
			for y := -1; y < 2; y++ {
				chk := dim.GetChunkContaining(updatePos.Add(core.NewVec3(x, y, 1)))
				if chk != nil {
					MakeChunkMeshAndVAO(dim, chk)
				}
			}
		}
	}
	if core.FlooredRemainder(updatePos.Z, 16) == 0 {
		for x := -1; x < 2; x++ {
			for y := -1; y < 2; y++ {
				chk := dim.GetChunkContaining(updatePos.Add(core.NewVec3(x, y, -1)))
				if chk != nil {
					MakeChunkMeshAndVAO(dim, chk)
				}
			}
		}
	}
}

func MakeChunkMesh(d *core.Dimension, chunkPos core.Vec3) (verts, normals, uvs, lightLevels []float32) {
	chunkGuess := d.Chunks[chunkPos]

	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			for z := 0; z < 16; z++ {
				local := core.NewVec3(x, y, z)
				global := chunkPos.Add(local)
				b := d.GetBlockAtOptimised(global, chunkGuess)
				if b.Type == nil {
					continue
				}

				// Top
				top := d.GetBlockAtOptimised(global.Add(core.Vec3{Y: 1}), chunkGuess)
				if top.Type == nil || top.Type.Transparent {
					face := core.FaceTop
					ll, flip := MakeLightLevelsForFace(d, global, face)
					v, n, u := b.Type.RenderType.RenderFace(face, local.ToFloat())
					verts = append(verts, flipIfTrue(v, flip, face, 3)...)
					normals = append(normals, flipIfTrue(n, flip, face, 3)...)
					uvs = append(uvs, flipIfTrue(u, flip, face, 2)...)
					lightLevels = append(lightLevels, flipIfTrue(ll, flip, face, 1)...)
				}

				// Bottom
				bottom := d.GetBlockAtOptimised(global.Add(core.Vec3{Y: -1}), chunkGuess)
				if bottom.Type == nil || bottom.Type.Transparent {
					face := core.FaceBottom
					ll, flip := MakeLightLevelsForFace(d, global, face)
					v, n, u := b.Type.RenderType.RenderFace(face, local.ToFloat())
					verts = append(verts, flipIfTrue(v, flip, face, 3)...)
					normals = append(normals, flipIfTrue(n, flip, face, 3)...)
					uvs = append(uvs, flipIfTrue(u, flip, face, 2)...)
					lightLevels = append(lightLevels, flipIfTrue(ll, flip, face, 1)...)
				}

				// Left
				left := d.GetBlockAtOptimised(global.Add(core.Vec3{X: -1}), chunkGuess)
				if left.Type == nil || left.Type.Transparent {
					face := core.FaceLeft
					ll, flip := MakeLightLevelsForFace(d, global, face)
					v, n, u := b.Type.RenderType.RenderFace(face, local.ToFloat())
					verts = append(verts, flipIfTrue(v, flip, face, 3)...)
					normals = append(normals, flipIfTrue(n, flip, face, 3)...)
					uvs = append(uvs, flipIfTrue(u, flip, face, 2)...)
					lightLevels = append(lightLevels, flipIfTrue(ll, flip, face, 1)...)
				}

				// Right
				right := d.GetBlockAtOptimised(global.Add(core.Vec3{X: 1}), chunkGuess)
				if right.Type == nil || right.Type.Transparent {
					face := core.FaceRight
					ll, flip := MakeLightLevelsForFace(d, global, face)
					v, n, u := b.Type.RenderType.RenderFace(face, local.ToFloat())
					verts = append(verts, flipIfTrue(v, flip, face, 3)...)
					normals = append(normals, flipIfTrue(n, flip, face, 3)...)
					uvs = append(uvs, flipIfTrue(u, flip, face, 2)...)
					lightLevels = append(lightLevels, flipIfTrue(ll, flip, face, 1)...)
				}

				// Front
				front := d.GetBlockAtOptimised(global.Add(core.Vec3{Z: 1}), chunkGuess)
				if front.Type == nil || front.Type.Transparent {
					face := core.FaceFront
					ll, flip := MakeLightLevelsForFace(d, global, face)
					v, n, u := b.Type.RenderType.RenderFace(face, local.ToFloat())
					verts = append(verts, flipIfTrue(v, flip, face, 3)...)
					normals = append(normals, flipIfTrue(n, flip, face, 3)...)
					uvs = append(uvs, flipIfTrue(u, flip, face, 2)...)
					lightLevels = append(lightLevels, flipIfTrue(ll, flip, face, 1)...)
				}

				// Back
				back := d.GetBlockAtOptimised(global.Add(core.Vec3{Z: -1}), chunkGuess)
				if back.Type == nil || back.Type.Transparent {
					face := core.FaceBack
					ll, flip := MakeLightLevelsForFace(d, global, face)
					v, n, u := b.Type.RenderType.RenderFace(face, local.ToFloat())
					verts = append(verts, flipIfTrue(v, flip, face, 3)...)
					normals = append(normals, flipIfTrue(n, flip, face, 3)...)
					uvs = append(uvs, flipIfTrue(u, flip, face, 2)...)
					lightLevels = append(lightLevels, flipIfTrue(ll, flip, face, 1)...)
				}
			}
		}
	}

	return
}

// Returns the light levels for each vertex of a face, and whether the triangles should be flipped
// for that face
func MakeLightLevelsForFace(dim *core.Dimension, pos core.Vec3, face core.BlockFace) ([]float32, bool) {
	var out []float32
	var shared mgl32.Vec3
	m := make(map[mgl32.Vec3]float32)
	for i := 0; i < 18; i += 3 {
		fv := mgl32.Vec3{faceVertices[face][i], faceVertices[face][i+1], faceVertices[face][i+2]}

		ll := getLightLevelForVert(dim, pos.ToFloat().Add(fv), face)
		out = append(out, ll)
		if _, ok := m[fv]; ok {
			shared = fv
		}
		m[fv] = ll
	}

	// Check which diagonal has the brightest lighting
	diagLight := float32(-1)
	for k, v := range m {
		for l, w := range m {
			if k.Sub(l).LenSqr() == 2 {
				// This pair is a diagonal
				if diagLight < 0 {
					diagLight = v + w
				}
				if diagLight > 0 && v+w != float32(diagLight) {
					if k == shared || l == shared {
						return out, v+w < diagLight
					} else {
						return out, v+w > diagLight
					}
				}
			}
		}
	}

	return out, false
}

// Flip the triangles for a face, if the cond is true, and with vec size n
func flipIfTrue(verts []float32, cond bool, face core.BlockFace, n int) []float32 {
	if !cond {
		return verts
	}

	out := make([]float32, len(verts))
	for k, v := range flipMap[face] {
		for i := 0; i < n; i++ {
			out[k*n+i] = verts[v*n+i]
		}
	}

	return out
}

func getLightLevelForVert(dim *core.Dimension, pos mgl32.Vec3, face core.BlockFace) float32 {
	// Take the average of the block light levels surrounding this vert
	p := core.NewVec3FromFloat(pos)
	switch face {
	case core.FaceTop, core.FaceBottom:
		// Check the 2x2 area around this vert, on the XZ plane
		if face == core.FaceBottom {
			p = p.Add(core.NewVec3(0, -1, 0))
		}

		sum := getBlockLightLevel(dim, p)
		sum += getBlockLightLevel(dim, p.Add(core.NewVec3(-1, 0, 0)))
		sum += getBlockLightLevel(dim, p.Add(core.NewVec3(-1, 0, -1)))
		sum += getBlockLightLevel(dim, p.Add(core.NewVec3(0, 0, -1)))
		return sum / 4

	case core.FaceLeft, core.FaceRight:
		if face == core.FaceRight {
			p = p.Add(core.NewVec3(1, 0, 0))
		}

		sum := getBlockLightLevel(dim, p.Add(core.NewVec3(-1, 0, 0)))
		sum += getBlockLightLevel(dim, p.Add(core.NewVec3(-1, 0, -1)))
		sum += getBlockLightLevel(dim, p.Add(core.NewVec3(-1, -1, 0)))
		sum += getBlockLightLevel(dim, p.Add(core.NewVec3(-1, -1, -1)))
		return sum / 4

	case core.FaceFront, core.FaceBack:
		if face == core.FaceFront {
			p = p.Add(core.NewVec3(0, 0, 1))
		}

		sum := getBlockLightLevel(dim, p.Add(core.NewVec3(0, 0, -1)))
		sum += getBlockLightLevel(dim, p.Add(core.NewVec3(-1, 0, -1)))
		sum += getBlockLightLevel(dim, p.Add(core.NewVec3(0, -1, -1)))
		sum += getBlockLightLevel(dim, p.Add(core.NewVec3(-1, -1, -1)))
		return sum / 4

	default:
		panic("unknown face")
	}
}

func getBlockLightLevel(dim *core.Dimension, pos core.Vec3) float32 {
	// TODO STUB
	// Turns out optimising this GetBlockAt barely improves perf
	if dim.GetBlockAt(pos).Type == nil {
		return 1
	} else {
		return 0
	}
}
