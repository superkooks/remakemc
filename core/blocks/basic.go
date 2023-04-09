package blocks

import (
	"remakemc/client/renderers"
	"remakemc/core"
)

var Grass = core.AddBlockToRegistry(&core.BlockType{
	Name:       "mc:grass",
	RenderType: renderers.BasicOneTex{Tex: "grass"},
})

var Dirt = core.AddBlockToRegistry(&core.BlockType{
	Name:       "mc:dirt",
	RenderType: renderers.BasicOneTex{Tex: "dirt"},
})

var Stone = core.AddBlockToRegistry(&core.BlockType{
	Name:       "mc:stone",
	RenderType: renderers.BasicOneTex{Tex: "stone"},
})

var Cobblestone = core.AddBlockToRegistry(&core.BlockType{
	Name:       "mc:cobblestone",
	RenderType: renderers.BasicOneTex{Tex: "cobblestone"},
})
