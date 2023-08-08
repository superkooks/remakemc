package entities

import (
	"remakemc/client/renderers"
	"remakemc/core"
)

type RemotePlayer struct {
	core.EntityBase
	core.PositionComp
	core.LerpComp
	core.LookComp
}

func (r *RemotePlayer) GetTypeName() string {
	return "mc:remote_player"
}

func (r *RemotePlayer) RenderInit() {
	r.GetRenderComp().Init()
}

func (r *RemotePlayer) GetRenderComp() core.RenderEntityType {
	return &renderers.TestEntityRenderer{
		Vertices: []float32{
			0, 0, 0,
			0.6, 0, 0.6,
			0.6, 1.8, 0.6,

			0.6, 1.8, 0.6,
			0.6, 0, 0.6,
			0, 0, 0,
		},

		Shader: "mc:test_entity",
	}
}

var _ = core.AddEntityToRegistry(new(RemotePlayer))
