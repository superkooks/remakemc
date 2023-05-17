package items

import (
	"remakemc/client/renderers"
	"remakemc/core"
)

var Cobblestone = core.AddItemToRegistry(&core.ItemType{
	Name:         "mc:cobblestone",
	MaxStackSize: 64,
	RenderType: &renderers.ItemFromBlock{
		Block: "mc:cobblestone",
	},
})
