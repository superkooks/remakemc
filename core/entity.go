package core

import (
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

// Minecraft has some significant gravity, apparently
var GRAVITY = mgl32.Vec3{0, -32, 0}

type Entity struct {
	ID         uuid.UUID
	Position   mgl32.Vec3
	AABB       mgl32.Vec3 // AABB cannot be < 0
	NoGravity  bool
	EntityType string

	// Disables physics and uses linear interpolation instead
	Lerp bool

	Velocity mgl32.Vec3 // in m/s

	Yaw           float64
	Pitch         float64
	LookAzimuth   float64
	LookElevation float64

	onGround bool

	lerpStartPos  mgl32.Vec3
	lerpStartTime time.Time
	lerpEndPos    mgl32.Vec3
	lerpEndTime   time.Time

	Equipment EntityEquipment
}

type EntityEquipment struct {
	HeldItemType string
}

type EntityType struct {
	Name       string
	RenderType RenderEntityType
}

type RenderEntityType interface {
	Init()
	RenderEntity(e *Entity, view mgl32.Mat4)
}

func (e *Entity) NewLerp(end mgl32.Vec3) {
	e.lerpStartPos = e.lerpEndPos
	e.lerpStartTime = e.lerpEndTime
	e.lerpEndPos = end
	e.lerpEndTime = time.Now()
}

func (e *Entity) DoUpdate(deltaT float64, dim *Dimension) {
	// If lerping, do lerp
	if e.Lerp {
		lerpDelta := e.lerpEndTime.Sub(e.lerpStartTime)
		scalar := time.Since(e.lerpEndTime) / lerpDelta

		dir := e.lerpEndPos.Sub(e.lerpStartPos)
		e.Position = dir.Mul(float32(scalar)).Add(e.lerpStartPos)

		return
	}

	// Otherwise, use physics

	// Move the player according to the current velocity
	e.Position = e.Position.Add(e.Velocity.Mul(float32(deltaT)))

	// Continually resolve collision, up to a maxiumum of 16 per update
	collisionsPerUpdate := 16
	var yAxisResolved bool
	for {
		intersectingBlock, intersects := e.GetBlockIntersecting(dim)
		if !intersects {
			break
		}

		// Calculate penetration time of the entity in the block
		// (How long ago did the entity start penetrating this block)
		var penTime mgl32.Vec3
		bl := intersectingBlock.ToFloat()

		// X Axis
		if e.Velocity.X() != 0 {
			// Compute the smallest intersection interval in terms of time
			d0 := bl.X() + 1 - e.Position.X()
			d1 := e.Position.X() + e.AABB.X() - bl.X()

			if d0 > 0 && d1 > 0 {
				if d0 < d1 {
					penTime[0] = d0 / e.Velocity.X()
				} else {
					penTime[0] = -d1 / e.Velocity.X()
				}
			}
		}
		if penTime[0] >= 0 {
			penTime[0] = mgl32.InfNeg
		}

		// Y Axis
		if e.Velocity.Y() != 0 {
			// Compute the smallest intersection interval
			d0 := bl.Y() + 1 - e.Position.Y()
			d1 := e.Position.Y() + e.AABB.Y() - bl.Y()

			if d0 > 0 && d1 > 0 {
				if d0 < d1 {
					penTime[1] = d0 / e.Velocity.Y()
				} else {
					penTime[1] = -d1 / e.Velocity.Y()
				}
			}
		}
		if penTime[1] >= 0 {
			penTime[1] = mgl32.InfNeg
		}

		// Z Axis
		if e.Velocity.Z() != 0 {
			// Compute the smallest intersection interval
			d0 := bl.Z() + 1 - e.Position.Z()
			d1 := e.Position.Z() + e.AABB.Z() - bl.Z()

			if d0 > 0 && d1 > 0 {
				if d0 < d1 {
					penTime[2] = d0 / e.Velocity.Z()
				} else {
					penTime[2] = -d1 / e.Velocity.Z()
				}
			}
		}
		if penTime[2] >= 0 {
			penTime[2] = mgl32.InfNeg
		}

		// Resolve the penetration by translating the entity to the latest time
		// the intersection could have happened
		if penTime.X() >= penTime.Y() && penTime.X() >= penTime.Z() {
			e.Position[0] += penTime.X() * e.Velocity.X()
			e.Velocity[0] = 0
		} else if penTime.Z() >= penTime.X() && penTime.Z() >= penTime.Y() {
			e.Position[2] += penTime.Z() * e.Velocity.Z()
			e.Velocity[2] = 0
		} else if penTime.Y() >= penTime.X() && penTime.Y() >= penTime.Z() {
			e.Position[1] += penTime.Y() * e.Velocity.Y()
			e.Velocity[1] = 0
			yAxisResolved = true
		}

		collisionsPerUpdate--
		if collisionsPerUpdate == 0 {
			break
		}
	}

	if yAxisResolved {
		e.onGround = true
	} else if e.Velocity.Y() != 0 {
		e.onGround = false
	}
}

// Gets the first block that with the entity's AABB
func (e *Entity) GetBlockIntersecting(dim *Dimension) (block Vec3, intersects bool) {
	// Iterate over each block that could intersect the AABB
	// and get the first one that intersects
	minX := FloorFloat32(e.Position.X())
	minY := FloorFloat32(e.Position.Y())
	minZ := FloorFloat32(e.Position.Z())

	maxX := CeilFloat32(e.Position.X() + e.AABB.X())
	maxY := CeilFloat32(e.Position.Y() + e.AABB.Y())
	maxZ := CeilFloat32(e.Position.Z() + e.AABB.Z())

	for x := minX; x < maxX; x++ {
		for y := minY; y < maxY; y++ {
			for z := minZ; z < maxZ; z++ {
				if dim.GetBlockAt(NewVec3(x, y, z)).Type != nil {
					return NewVec3(x, y, z), true
				}
			}
		}
	}

	return Vec3{}, false
}

// Physics ticks happen at 20Hz, whereas physics updates are dependent
// on the deltaT
func (e *Entity) DoTick() {
	if !e.NoGravity {
		// Update the entity's velocity based on the minecraft movement formula
		// https://www.mcpk.wiki/wiki/Vertical_Movement_Formulas
		// Note that we do things in m/s, not m/tick

		// Vertical speed decremented (less upward motion, more downward motion)
		// by 0.08 blocks per tick (32 m/s), then multiplied by 0.98 per tick.
		e.Velocity = e.Velocity.Add(GRAVITY.Mul(1.0 / 20))
		e.Velocity[1] *= 0.98
	}
}

func (e *Entity) OnGround() bool {
	return e.onGround
}
