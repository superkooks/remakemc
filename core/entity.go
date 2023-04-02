package core

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

var GRAVITY = mgl32.Vec3{0, -9.8, 0}

type Impulse struct {
	// The force vector
	Force mgl32.Vec3

	// How long the impulse has remaining, in seconds
	Remaining float64
}

type Entity struct {
	Position mgl32.Vec3
	AABB     mgl32.Vec3 // AABB cannot be < 0

	// Updated every physics tick. Should not be set manually.
	Velocity mgl32.Vec3
	OnGround bool

	impulses []Impulse
}

func (e *Entity) AddImpulse(i Impulse) {
	e.impulses = append(e.impulses, i)
}

func (e *Entity) PhysicsTick(deltaT float64, dim *Dimension) {
	// Calculate net force
	var netAcc mgl32.Vec3
	for _, v := range e.impulses {
		netAcc = netAcc.Add(v.Force)
	}
	netAcc = netAcc.Add(GRAVITY)

	// Calculate acceleration, and hence displacement
	e.Velocity = e.Velocity.Add(netAcc.Mul(float32(deltaT)))
	e.Position = e.Position.Add(e.Velocity.Mul(float32(deltaT)))

	// Continually resolve collision, up to a maxiumum of 16 per tick
	collisionsPerTick := 16
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
		var penTime mgl32.Vec3
		bl := intersectingBlock.ToFloat()

		// X Axis
		if e.Velocity.X() != 0 {
			// Compute the smallest intersection interval in terms of time
			d0 := bl.X() + 1 - e.Position.X()
			d1 := e.Position.X() + e.AABB.X() - bl.X()

			// fmt.Println(d0, d1)

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

		fmt.Println(penTime)

		// Resolve Y on first collision
		if collisionsPerTick == 16 && penTime.Y() > -0.03 {
			e.Position[1] += penTime.Y() * e.Velocity.Y()
			e.Velocity[1] = 0
			continue
		}

		// Resolve the penetration by translating the entity to the latest time
		// the intersection could have happened, with priority to Y axis
		if penTime.Y() >= penTime.X() && penTime.Y() >= penTime.Z() {
			e.Position[1] += penTime.Y() * e.Velocity.Y()
			e.Velocity[1] = 0
		} else if penTime.X() >= penTime.Y() && penTime.X() >= penTime.Z() {
			e.Position[0] += penTime.X() * e.Velocity.X()
			e.Velocity[0] = 0
		} else if penTime.Z() >= penTime.X() && penTime.Z() >= penTime.Y() {
			e.Position[2] += penTime.Z() * e.Velocity.Z()
			e.Velocity[2] = 0
		}

		collisionsPerTick--
		if collisionsPerTick == 0 {
			break
		}
	}

	// Decrement Remaining and remove expired impulses
	for i := len(e.impulses) - 1; i >= 0; i-- {
		e.impulses[i].Remaining -= deltaT
		if e.impulses[i].Remaining < 0 {
			e.impulses = append(e.impulses[:i], e.impulses[i+1:]...)
		}
	}
}
