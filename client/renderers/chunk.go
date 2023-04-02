package renderers

import (
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

out vec3 fragNormal;
out vec3 fragVertex;
out vec2 fragUV;
out mat4 fragModel;

void main() {
	gl_Position = projection * view * model * vec4(vp, 1.0);

	fragVertex = vp;
	fragNormal = vertexNormal;
	fragUV = vertexUV;
	fragModel = model;
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

func RenderChunks(dim *core.Dimension, view mgl32.Mat4) {
	gl.UseProgram(chunkProg)
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Blend transparency
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Assign view & projection mats
	projection := mgl32.Perspective(mgl32.DegToRad(FOVDegrees), GetAspectRatio(), 0.1, 3000.0)
	gl.UniformMatrix4fv(chunkPUniform, 1, false, &projection[0])
	gl.UniformMatrix4fv(chunkVUniform, 1, false, &view[0])

	// Activate texture atlas
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, chunkTex)
	gl.Uniform1i(chunkTUniform, int32(chunkTex-1))

	for x := 0; x < 128; x += 16 {
		for y := 0; y < 256; y += 16 {
			for z := 0; z < 128; z += 16 {
				RenderChunk(dim.Chunks[core.NewVec3(x, y, z)], view)
			}
		}
	}
}

func RenderChunk(c *core.Chunk, view mgl32.Mat4) {
	if c.MeshLen == 0 {
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

func MakeChunkVAO(d *core.Dimension, chunk *core.Chunk) {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var mesh, normals, uvs []float32
	mesh, normals, uvs = MakeChunkMesh(d, chunk.Position)
	chunk.MeshLen = len(mesh)

	if len(mesh) == 0 {
		return
	}

	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, GlBufferFrom(mesh))
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil) // vec3

	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, GlBufferFrom(normals))
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil) // vec3

	gl.EnableVertexAttribArray(2)
	gl.BindBuffer(gl.ARRAY_BUFFER, GlBufferFrom(uvs))
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 0, nil) // vec2

	chunk.VAO = vao
}

func UpdateRequiredMeshes(dim *core.Dimension, updatePos core.Vec3) {
	MakeChunkVAO(dim, dim.GetChunkContaining(updatePos))

	// X
	if core.FlooredRemainder(updatePos.X, 16) == 15 {
		chk := dim.GetChunkContaining(updatePos.Add(core.Vec3{X: 1}))
		if chk != nil {
			MakeChunkVAO(dim, chk)
		}
	}
	if core.FlooredRemainder(updatePos.X, 16) == 0 {
		chk := dim.GetChunkContaining(updatePos.Add(core.Vec3{X: -1}))
		if chk != nil {
			MakeChunkVAO(dim, chk)
		}
	}

	// Y
	if core.FlooredRemainder(updatePos.Y, 16) == 15 {
		chk := dim.GetChunkContaining(updatePos.Add(core.Vec3{Y: 1}))
		if chk != nil {
			MakeChunkVAO(dim, chk)
		}
	}
	if core.FlooredRemainder(updatePos.Y, 16) == 0 {
		chk := dim.GetChunkContaining(updatePos.Add(core.Vec3{Y: -1}))
		if chk != nil {
			MakeChunkVAO(dim, chk)
		}
	}

	// Z
	if core.FlooredRemainder(updatePos.Z, 16) == 15 {
		chk := dim.GetChunkContaining(updatePos.Add(core.Vec3{Z: 1}))
		if chk != nil {
			MakeChunkVAO(dim, chk)
		}
	}
	if core.FlooredRemainder(updatePos.Z, 16) == 0 {
		chk := dim.GetChunkContaining(updatePos.Add(core.Vec3{Z: -1}))
		if chk != nil {
			MakeChunkVAO(dim, chk)
		}
	}
}

func MakeChunkMesh(d *core.Dimension, chunkPos core.Vec3) (verts, normals, uvs []float32) {
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
				top := d.GetBlockAt(global.Add(core.Vec3{Y: 1}))
				if top.Type == nil || top.Type.Transparent {
					v, n, u := b.Type.RenderType.RenderFace(core.FaceTop, local.ToFloat())
					verts = append(verts, v...)
					normals = append(normals, n...)
					uvs = append(uvs, u...)
				}

				// Bottom
				bottom := d.GetBlockAt(global.Add(core.Vec3{Y: -1}))
				if bottom.Type == nil || bottom.Type.Transparent {
					v, n, u := b.Type.RenderType.RenderFace(core.FaceBottom, local.ToFloat())
					verts = append(verts, v...)
					normals = append(normals, n...)
					uvs = append(uvs, u...)
				}

				// Left
				left := d.GetBlockAt(global.Add(core.Vec3{X: -1}))
				if left.Type == nil || left.Type.Transparent {
					v, n, u := b.Type.RenderType.RenderFace(core.FaceLeft, local.ToFloat())
					verts = append(verts, v...)
					normals = append(normals, n...)
					uvs = append(uvs, u...)
				}

				// Right
				right := d.GetBlockAt(global.Add(core.Vec3{X: 1}))
				if right.Type == nil || right.Type.Transparent {
					v, n, u := b.Type.RenderType.RenderFace(core.FaceRight, local.ToFloat())
					verts = append(verts, v...)
					normals = append(normals, n...)
					uvs = append(uvs, u...)
				}

				// Front
				front := d.GetBlockAt(global.Add(core.Vec3{Z: 1}))
				if front.Type == nil || front.Type.Transparent {
					v, n, u := b.Type.RenderType.RenderFace(core.FaceFront, local.ToFloat())
					verts = append(verts, v...)
					normals = append(normals, n...)
					uvs = append(uvs, u...)
				}

				// Back
				back := d.GetBlockAt(global.Add(core.Vec3{Z: -1}))
				if back.Type == nil || back.Type.Transparent {
					v, n, u := b.Type.RenderType.RenderFace(core.FaceBack, local.ToFloat())
					verts = append(verts, v...)
					normals = append(normals, n...)
					uvs = append(uvs, u...)
				}
			}
		}
	}

	return
}
