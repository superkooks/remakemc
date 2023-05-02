package renderers

import "github.com/go-gl/gl/v4.1-core/gl"

func initEntityShaders() {
	ReusableShaders["mc:test"] = func() *EntityShader {
		// TEST SHADER
		vert, err := compileShader(`
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

		frag, err := compileShader(`
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

		prog := gl.CreateProgram()
		gl.AttachShader(prog, vert)
		gl.AttachShader(prog, frag)
		gl.LinkProgram(prog)

		eshad := EntityShader{Program: prog, Uniforms: make(map[string]int32)}

		eshad.Uniforms["projection"] = gl.GetUniformLocation(prog, gl.Str("projection\x00"))
		eshad.Uniforms["view"] = gl.GetUniformLocation(prog, gl.Str("view\x00"))
		eshad.Uniforms["model"] = gl.GetUniformLocation(prog, gl.Str("model\x00"))

		return &eshad
	}
}
