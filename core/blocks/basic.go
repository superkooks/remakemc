package blocks

import (
	"remakemc/client/renderers"
	"remakemc/core"
)

var Grass = core.AddBlockToRegistry(&core.BlockType{
	Name:       "mc:grass",
	RenderType: renderers.BlockBasicOneTex{Tex: "grass"},
})

var Dirt = core.AddBlockToRegistry(&core.BlockType{
	Name:       "mc:dirt",
	RenderType: renderers.BlockBasicOneTex{Tex: "dirt"},
})

var Stone = core.AddBlockToRegistry(&core.BlockType{
	Name:       "mc:stone",
	RenderType: renderers.BlockBasicOneTex{Tex: "stone"},
})

var Cobblestone = core.AddBlockToRegistry(&core.BlockType{
	Name:       "mc:cobblestone",
	RenderType: renderers.BlockBasicOneTex{Tex: "cobblestone"},
})
