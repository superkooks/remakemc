package renderers

import "github.com/go-gl/gl/v4.1-core/gl"

func initReusableShaders() {
	ReusableShaders["mc:test_entity"] = func() *Shader {
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

		shad := Shader{Program: prog, Uniforms: make(map[string]int32)}

		shad.Uniforms["projection"] = gl.GetUniformLocation(prog, gl.Str("projection\x00"))
		shad.Uniforms["view"] = gl.GetUniformLocation(prog, gl.Str("view\x00"))
		shad.Uniforms["model"] = gl.GetUniformLocation(prog, gl.Str("model\x00"))

		return &shad
	}

	ReusableShaders["mc:item_from_block"] = func() *Shader {
		vert, err := compileShader(`
#version 410

layout (location = 0) in vec3 vp;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec2 uv;
uniform vec2 modelStart;
uniform vec2 modelEnd;
uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;
out vec2 fragUV;
out vec3 fragNormal;
out mat4 fragModel;
out vec3 fragVertex;

void main() {
	vec2 box = modelEnd - modelStart;
	vec4 inter = projection * view * model * vec4(vp, 1.0);
	inter *= 0.39;
	inter += vec4(0.5, 0.5, 0, 0);

	
	gl_Position.x = modelStart.x + inter.x*box.x;
	gl_Position.y = modelStart.y + inter.y*box.y;
	gl_Position.z = 1;
	gl_Position.w = 1;

	fragUV = uv;
	fragNormal = normal;
	fragModel = model;
	fragVertex = vp;
}`+"\x00", gl.VERTEX_SHADER)
		if err != nil {
			panic(err)
		}

		frag, err := compileShader(`
#version 410

in vec3 fragVertex;
in vec2 fragUV;
in vec3 fragNormal;
in mat4 fragModel;
uniform sampler2D tex;
out vec4 color;

vec3 ApplyLight(vec3 surfaceColor, vec3 normal, vec3 surfacePos, vec3 surfaceToCamera) {
	// Directional light
	vec3 surfaceToLight = normalize(vec3(-0.2, 0.6, -0.2));
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
	vec4 surfaceColor = texture(tex, fragUV);
	// vec4 surfaceColor = vec4(1.0);
	vec3 surfaceToCamera = normalize(surfacePos);

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

		shad := Shader{Program: prog, Uniforms: make(map[string]int32)}

		shad.Uniforms["modelStart"] = gl.GetUniformLocation(prog, gl.Str("modelStart\x00"))
		shad.Uniforms["modelEnd"] = gl.GetUniformLocation(prog, gl.Str("modelEnd\x00"))
		shad.Uniforms["projection"] = gl.GetUniformLocation(prog, gl.Str("projection\x00"))
		shad.Uniforms["view"] = gl.GetUniformLocation(prog, gl.Str("view\x00"))
		shad.Uniforms["model"] = gl.GetUniformLocation(prog, gl.Str("model\x00"))
		shad.Uniforms["tex"] = gl.GetUniformLocation(prog, gl.Str("tex\x00"))

		return &shad
	}
}
