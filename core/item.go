package core

import "github.com/go-gl/mathgl/mgl32"

type ItemType struct {
	Name         string
	MaxStackSize int
	RenderType   RenderItemType
}

type ItemStack struct {
	Item  *ItemType
	Count int
}

type RenderItemType interface {
	Init()
	RenderItem(i *ItemType, boxStart mgl32.Vec2, boxEnd mgl32.Vec2)
}
