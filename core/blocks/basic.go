package blocks

import (
	"remakemc/client/renderers"
	"remakemc/core"
)

var Grass = &core.BlockType{
	RenderType: renderers.BasicOneTex{Tex: "grass"},
}

var Dirt = &core.BlockType{
	RenderType: renderers.BasicOneTex{Tex: "dirt"},
}

var Stone = &core.BlockType{
	RenderType: renderers.BasicOneTex{Tex: "stone"},
}

var Cobblestone = &core.BlockType{
	RenderType: renderers.BasicOneTex{Tex: "cobblestone"},
}
