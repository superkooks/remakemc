package blocks

import (
	"remakemc/client/renderers"
	"remakemc/core"
)

var Grass = core.BlockType{
	RenderType: renderers.BasicOneTex("grass"),
}
