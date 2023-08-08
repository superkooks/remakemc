package core

import (
	"github.com/go-gl/mathgl/mgl32"
)

type PhysicsFace interface {
	GetPhysicsComp() *PhysicsComp
}

type PhysicsComp struct {
	AABB      mgl32.Vec3 // AABB cannot be < 0
	NoGravity bool
	Velocity  mgl32.Vec3 // in m/s

	onGround bool
}

func (p *PhysicsComp) GetPhysicsComp() *PhysicsComp {
	return p
}

func (p *PhysicsComp) OnGround() bool {
	return p.onGround
}

func PhysicsTickSystem(dim *Dimension) {
	bodies := GetEntitiesSatisfying[interface {
		PhysicsFace
		PositionFace
	}](dim.Entities)

	for _, v := range bodies {
		e := v.GetPhysicsComp()

		if !e.NoGravity {
			// Vertical speed decremented (less upward motion, more downward motion)
			// by 0.08 blocks per tick (32 m/s), then multiplied by 0.98 per tick.
			e.Velocity = e.Velocity.Add(GRAVITY.Mul(1.0 / 20))
			e.Velocity[1] *= 0.98
		}
	}
}

func PhysicsSystem(dim *Dimension, deltaT float32) {
	bodies := GetEntitiesSatisfying[interface {
		PhysicsFace
		PositionFace
	}](dim.Entities)

	for _, v := range bodies {
		e := v.GetPhysicsComp()
		pos := v.GetPosition()

		// Update the entity's velocity based on the minecraft movement formula
		// https://www.mcpk.wiki/wiki/Vertical_Movement_Formulas
		// Note that we do things in m/s, not m/tick

		// Move the player according to the current velocity
		*pos = pos.Add(e.Velocity.Mul(deltaT))

		// Continually resolve collision, up to a maxiumum of 16 per update
		collisionsPerUpdate := 16
		var yAxisResolved bool
		for {
			intersectingBlock, intersects := getBlockIntersecting(dim, *pos, e.AABB)
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
				d0 := bl.X() + 1 - pos.X()
				d1 := pos.X() + e.AABB.X() - bl.X()

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
				d0 := bl.Y() + 1 - pos.Y()
				d1 := pos.Y() + e.AABB.Y() - bl.Y()

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
				d0 := bl.Z() + 1 - pos.Z()
				d1 := pos.Z() + e.AABB.Z() - bl.Z()

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
				pos[0] += penTime.X() * e.Velocity.X()
				e.Velocity[0] = 0
			} else if penTime.Z() >= penTime.X() && penTime.Z() >= penTime.Y() {
				pos[2] += penTime.Z() * e.Velocity.Z()
				e.Velocity[2] = 0
			} else if penTime.Y() >= penTime.X() && penTime.Y() >= penTime.Z() {
				pos[1] += penTime.Y() * e.Velocity.Y()
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
}

// Gets the first block that with the entity's AABB
func getBlockIntersecting(dim *Dimension, pos mgl32.Vec3, aabb mgl32.Vec3) (block Vec3, intersects bool) {
	// Iterate over each block that could intersect the AABB
	// and get the first one that intersects
	minX := FloorFloat32(pos.X())
	minY := FloorFloat32(pos.Y())
	minZ := FloorFloat32(pos.Z())

	maxX := CeilFloat32(pos.X() + aabb.X())
	maxY := CeilFloat32(pos.Y() + aabb.Y())
	maxZ := CeilFloat32(pos.Z() + aabb.Z())

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
