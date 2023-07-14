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

var Grass = core.AddItemToRegistry(&core.ItemType{
	Name:         "mc:grass",
	MaxStackSize: 64,
	RenderType: &renderers.ItemFromBlock{
		Block: "mc:grass",
	},
})

var Dirt = core.AddItemToRegistry(&core.ItemType{
	Name:         "mc:dirt",
	MaxStackSize: 64,
	RenderType: &renderers.ItemFromBlock{
		Block: "mc:dirt",
	},
})

var Stone = core.AddItemToRegistry(&core.ItemType{
	Name:         "mc:stone",
	MaxStackSize: 64,
	RenderType: &renderers.ItemFromBlock{
		Block: "mc:stone",
	},
})

var Furnace = core.AddItemToRegistry(&core.ItemType{
	Name:         "mc:furnace",
	MaxStackSize: 64,
	RenderType: &renderers.ItemFromBlock{
		Block: "mc:furnace",
	},
})
