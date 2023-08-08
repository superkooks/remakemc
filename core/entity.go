package core

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

// Minecraft has some significant gravity, apparently
var GRAVITY = mgl32.Vec3{0, -32, 0}

type Entity interface {
	GetID() uuid.UUID
	GetTypeName() string
}

type EntityBase struct {
	ID uuid.UUID
}

func (b *EntityBase) GetID() uuid.UUID {
	return b.ID
}

func GetEntitiesSatisfying[V any](entities []Entity) (out []V) {
	for _, v := range entities {
		if e, ok := v.(V); ok {
			out = append(out, e)
		}
	}
	return
}

type PositionFace interface {
	GetPosition() *mgl32.Vec3
}

type PositionComp struct {
	Position mgl32.Vec3
}

func (p *PositionComp) GetPosition() *mgl32.Vec3 {
	return &p.Position
}

type LookFace interface {
	GetLookComp() *LookComp
}

type LookComp struct {
	Yaw       float64
	Elevation float64
	Azimuth   float64
}

func (p *LookComp) GetLookComp() *LookComp {
	return p
}

type RenderFace interface {
	RenderInit()
	GetRenderComp() RenderEntityType
}

// type EntityEquipment struct {
// 	HeldItemType string
// }

// type EntityType struct {
// 	Name       string
// 	RenderType RenderEntityType
// 	IsBlock    bool

// 	// Triggered by right click
// 	// PlayerInteraction func(e *Entity, Dim *Dimension)
// 	// Update func(e *Entity) // Called every tick
// }

type RenderEntityType interface {
	Init()
	RenderEntity(e Entity, view mgl32.Mat4)
}
