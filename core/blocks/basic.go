package blocks

import (
	"remakemc/client/renderers"
	"remakemc/core"
)

var Grass = core.BlockType{
	RenderType: renderers.BasicOneTex("grass"),
}

var Dirt = core.BlockType{
	RenderType: renderers.BasicOneTex("dirt"),
}

var Stone = core.BlockType{
	RenderType: renderers.BasicOneTex("stone"),
}

var Cobblestone = core.BlockType{
	RenderType: renderers.BasicOneTex("cobblestone"),
}
