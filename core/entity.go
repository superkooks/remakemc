package core

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Minecraft has some significant gravity, apparently
var GRAVITY = mgl32.Vec3{0, -32, 0}

type Entity struct {
	Position  mgl32.Vec3
	AABB      mgl32.Vec3 // AABB cannot be < 0
	NoGravity bool

	// A callback that happens at each physics tick (not update)
	// Used for custom processing
	TickCallback func()

	Velocity mgl32.Vec3 // in m/s

	onGround       bool
	collectedDelta float64
}

func (e *Entity) PhysicsUpdate(deltaT float64, dim *Dimension) {
	// Move the player according to the current velocity
	e.Position = e.Position.Add(e.Velocity.Mul(float32(deltaT)))

	// Continually resolve collision, up to a maxiumum of 16 per update
	collisionsPerUpdate := 16
	var yAxisResolved bool
	for {
		// Iterate over each block that could intersect the AABB
		// and get the first one that intersects
		minX := FloorFloat32(e.Position.X())
		minY := FloorFloat32(e.Position.Y())
		minZ := FloorFloat32(e.Position.Z())

		maxX := CeilFloat32(e.Position.X() + e.AABB.X())
		maxY := CeilFloat32(e.Position.Y() + e.AABB.Y())
		maxZ := CeilFloat32(e.Position.Z() + e.AABB.Z())

		var intersectingBlock Vec3
		for x := minX; x < maxX; x++ {
			for y := minY; y < maxY; y++ {
				for z := minZ; z < maxZ; z++ {
					if dim.GetBlockAt(NewVec3(x, y, z)).Type != nil {
						intersectingBlock = NewVec3(x, y, z)
						goto resolve
					}
				}
			}
		}

		break

	resolve:
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

		// Resolve Y on first collision
		if collisionsPerUpdate == 16 && penTime.Y() > -0.03 {
			e.Position[1] += penTime.Y() * e.Velocity.Y()
			e.Velocity[1] = 0
			yAxisResolved = true
			continue
		}

		// Resolve the penetration by translating the entity to the latest time
		// the intersection could have happened
		if penTime.Y() >= penTime.X() && penTime.Y() >= penTime.Z() {
			e.Position[1] += penTime.Y() * e.Velocity.Y()
			e.Velocity[1] = 0
			yAxisResolved = true
		} else if penTime.X() >= penTime.Y() && penTime.X() >= penTime.Z() {
			e.Position[0] += penTime.X() * e.Velocity.X()
			e.Velocity[0] = 0
		} else if penTime.Z() >= penTime.X() && penTime.Z() >= penTime.Y() {
			e.Position[2] += penTime.Z() * e.Velocity.Z()
			e.Velocity[2] = 0
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

	// See if we need to do a physics tick
	e.collectedDelta += deltaT
	for ; e.collectedDelta >= 1.0/20; e.collectedDelta -= 1.0 / 20 {
		e.PhysicsTick()
		e.TickCallback()
	}
}

// Physics ticks happen at 20Hz, whereas physics updates are dependent
// on the deltaT
func (e *Entity) PhysicsTick() {
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
