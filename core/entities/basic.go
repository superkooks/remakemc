package entities

import (
	"remakemc/client/renderers"
	"remakemc/core"
)

var RemotePlayer = core.AddEntityToRegistry(&core.EntityType{
	Name: "mc:remoteplayer",
	RenderType: &renderers.TestEntityRenderer{
		Vertices: []float32{
			0, 0, 0,
			0.6, 0, 0.6,
			0.6, 1.8, 0.6,

			0.6, 1.8, 0.6,
			0.6, 0, 0.6,
			0, 0, 0,
		},

		Shader: "mc:test_entity",
	},
})
