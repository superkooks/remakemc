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

var Furnace = core.AddBlockToRegistry(&core.BlockType{
	Name:           "mc:furnace",
	LinkWithEntity: "mc:furnace",
	RenderType: renderers.BlockBasicSixTex{
		Top:    "furnace_top",
		Bottom: "furnace_top",
		Left:   "furnace_side",
		Right:  "furnace_side",
		Front:  "furnace_front",
		Back:   "furnace_side",
	},
})
