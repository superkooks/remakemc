package entities

import (
	"remakemc/core"
)

var Furnace = core.AddEntityToRegistry(&core.EntityType{
	Name:       "mc:furnace",
	IsBlock:    true,
	RenderType: nil,
})
