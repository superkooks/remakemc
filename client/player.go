package client

import (
	"math"
	"remakemc/client/renderers"
	"remakemc/core"
	"remakemc/core/container"
	"remakemc/core/proto"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

type Player struct {
	core.EntityBase
	core.PositionComp
	core.PhysicsComp
	core.LookComp

	Sprinting bool
	Sneaking  bool

	Speed           float64
	MouseSensitivty float64

	Inventory          *container.Inventory
	SelectedHotbarSlot int
}

func NewPlayer(position mgl32.Vec3, entityID uuid.UUID) *Player {
	p := &Player{
		Speed: 1,
		EntityBase: core.EntityBase{
			ID: entityID,
		},
		PhysicsComp: core.PhysicsComp{
			AABB: mgl32.Vec3{0.6, 1.8, 0.6},
		},
		PositionComp: core.PositionComp{
			Position: position,
		},
	}
	return p
}

func (p *Player) GetTypeName() string {
	return "mc:local_player"
}

// The position of the camera
func (p *Player) CameraPos() mgl32.Vec3 {
	if p.Sneaking {
		return p.Position.Add(mgl32.Vec3{0.3, 1.34, 0.3})
	}

	return p.Position.Add(mgl32.Vec3{0.3, 1.62, 0.3})
}

// The direction the player is looking
func (p *Player) LookDir() mgl32.Vec3 {
	return mgl32.Vec3{
		float32(math.Cos(p.Elevation) * math.Sin(p.Azimuth)),
		float32(math.Sin(p.Elevation)),
		float32(math.Cos(p.Elevation) * math.Cos(p.Azimuth)),
	}
}

// The direction the player would move if they pressed W
func (p *Player) ForwardVec() mgl32.Vec3 {
	return p.RightVec().Cross(mgl32.Vec3{0, -1, 0})
}

func (p *Player) RightVec() mgl32.Vec3 {
	return mgl32.Vec3{
		float32(math.Sin(p.Azimuth - math.Pi/2.0)),
		0,
		float32(math.Cos(p.Azimuth - math.Pi/2.0)),
	}
}

func (p *Player) SetSneaking(b bool) {
	p.Sneaking = b
	if b {
		p.AABB[1] = 1.5
	} else {
		p.AABB[1] = 1.8
	}
}

var inventoryButton = new(core.Debounced)
var escButton = new(core.Debounced)

// Process the keyboard input and physics for this tick
func PlayerSystem(dim *core.Dimension) {
	// Get player
	p := core.GetEntitiesSatisfying[*Player](dim.Entities)[0]

	// Inventory mode
	if containerOpen {
		if renderers.Win.GetKey(glfw.KeyE) == glfw.Press && inventoryButton.Invoke() {
			CloseContainer()
		} else if renderers.Win.GetKey(glfw.KeyE) == glfw.Release {
			inventoryButton.Reset()
		}

		if renderers.Win.GetKey(glfw.KeyEscape) == glfw.Press && escButton.Invoke() {
			CloseContainer()
		} else if renderers.Win.GetKey(glfw.KeyEscape) == glfw.Release {
			escButton.Reset()
		}

		// We still need to calculate drag
		slipperiness := 1.0
		if p.OnGround() {
			slipperiness = 0.6
		}
		p.Velocity[0] *= float32(0.91 * slipperiness)
		p.Velocity[2] *= float32(0.91 * slipperiness)

		return
	}

	// Hotbar slot
	for i := glfw.Key1; i <= glfw.Key9; i++ {
		if renderers.Win.GetKey(i) == glfw.Press {
			p.SelectedHotbarSlot = int(i - glfw.Key1)

			serverWrite <- proto.PLAYER_HELD_ITEM
			serverWrite <- p.SelectedHotbarSlot
		}
	}

	// Open inventory
	if renderers.Win.GetKey(glfw.KeyE) == glfw.Press && inventoryButton.Invoke() {
		OpenContainer(player.Inventory)
	} else if renderers.Win.GetKey(glfw.KeyE) == glfw.Release {
		inventoryButton.Reset()
	}

	if renderers.Win.GetKey(glfw.KeyEscape) == glfw.Release {
		escButton.Reset()
	}

	// Jump
	var jumping bool
	if renderers.Win.GetKey(glfw.KeySpace) == glfw.Press && p.OnGround() {
		// TODO Add timeout to next jump
		p.Velocity[1] = 8.4
		jumping = true
		serverWrite <- proto.PLAYER_JUMP
	}

	// Sneak
	if renderers.Win.GetKey(glfw.KeyLeftShift) == glfw.Press {
		// Nested for a reason
		if !p.Sneaking {
			p.SetSneaking(true)
			serverWrite <- proto.PLAYER_SNEAKING
			serverWrite <- proto.PlayerSneaking(true)
		}
	} else if p.Sneaking {
		p.SetSneaking(false)
		serverWrite <- proto.PLAYER_SNEAKING
		serverWrite <- proto.PlayerSneaking(false)
	}

	// Sprint
	if renderers.Win.GetKey(glfw.KeyLeftControl) == glfw.Press && !p.Sprinting {
		p.Sprinting = true
		serverWrite <- proto.PLAYER_SPRINTING
		serverWrite <- proto.PlayerSprinting(true)
	}

	var walkVec mgl32.Vec2

	// Move forwards
	if renderers.Win.GetKey(glfw.KeyW) == glfw.Press {
		walkVec[0] += 1
	}
	// Move backwards
	if renderers.Win.GetKey(glfw.KeyS) == glfw.Press {
		walkVec[0] += -1
	}
	// Move right
	if renderers.Win.GetKey(glfw.KeyD) == glfw.Press {
		walkVec[1] += 1
	}
	// Move left
	if renderers.Win.GetKey(glfw.KeyA) == glfw.Press {
		walkVec[1] += -1
	}

	if (walkVec.X() == 0 || p.Sneaking) && p.Sprinting {
		p.Sprinting = false
		serverWrite <- proto.PLAYER_SPRINTING
		serverWrite <- proto.PlayerSprinting(false)
	}

	// Procees horizontal velocity according to
	// https://www.mcpk.wiki/wiki/Horizontal_Movement_Formulas
	moveMult := 1.0
	if p.Sprinting {
		moveMult = 1.3
	} else if p.Sneaking {
		moveMult = 0.3
	} else if walkVec.X() == 0 && walkVec.Y() == 0 {
		moveMult = 0
	}
	if walkVec.X() == 0 || walkVec.Y() == 0 {
		moveMult *= 0.98
	} else if p.Sneaking {
		moveMult *= 0.98 * math.Sqrt(2)
	}

	slipperiness := 1.0
	if p.OnGround() {
		slipperiness = 0.6
	}

	groundAccel := float32(moveMult*p.Speed*math.Pow(0.6/slipperiness, 3)) * 0.1 * 20
	direction := p.Azimuth
	if moveMult != 0 {
		if walkVec.X() == 0 {
			if walkVec.Y() < 0 {
				direction += math.Pi / 2
			} else {
				direction -= math.Pi / 2
			}
		} else {
			direction -= math.Atan(float64(walkVec.Y() / walkVec.X()))
		}
	}
	if walkVec.X() < 0 {
		direction += math.Pi
	}

	p.Velocity[0] *= float32(0.91 * slipperiness)
	p.Velocity[2] *= float32(0.91 * slipperiness)

	if jumping {
		p.Velocity[0] += groundAccel * float32(math.Sin(direction))
		p.Velocity[2] += groundAccel * float32(math.Cos(direction))

		if p.Sprinting {
			p.Velocity[0] += 0.2 * float32(math.Sin(p.Azimuth)) * 20 * 0.91 * 0.6
			p.Velocity[2] += 0.2 * float32(math.Cos(p.Azimuth)) * 20 * 0.91 * 0.6
		}

	} else if p.OnGround() {
		p.Velocity[0] += groundAccel * float32(math.Sin(direction))
		p.Velocity[2] += groundAccel * float32(math.Cos(direction))

	} else {
		p.Velocity[0] += 0.02 * float32(moveMult*math.Sin(direction)*20)
		p.Velocity[2] += 0.02 * float32(moveMult*math.Cos(direction)*20)
	}

	// serverWrite <- proto.PLAYER_POSITION
	// serverWrite <- proto.PlayerPosition{
	// 	Position:      p.Position,
	// 	Yaw:           p.Yaw,
	// 	LookAzimuth:   p.Azimuth,
	// 	LookElevation: p.Elevation,
	// }
}

// func (p *Player) ScrollCallback(_ *glfw.Window, _, yoff float64) {
// 	if yoff < 0 && p.SelectedHotbarSlot < 8 {
// 		p.SelectedHotbarSlot++
// 	} else if yoff > 0 && p.SelectedHotbarSlot > 0 {
// 		p.SelectedHotbarSlot--
// 	}
// 	serverWrite <- proto.PLAYER_HELD_ITEM
// 	serverWrite <- p.SelectedHotbarSlot
// }

// Process the mouse input for this frame
func MouseSystem(dim *core.Dimension, deltaT float64) {
	// Get player
	p := core.GetEntitiesSatisfying[*Player](dim.Entities)[0]

	// Get and reset cursor position
	xpos, ypos := renderers.Win.GetCursorPos()

	if ypos < 0 {
		// Player is grabbing title bar, ignore it
		return
	}

	width, height := renderers.Win.GetSize()
	renderers.Win.SetCursorPos(float64(width)/2, float64(height)/2)

	// Calculate new orientation
	p.Azimuth += 0.001 * (float64(width)/2 - float64(xpos))
	p.Elevation += 0.001 * (float64(height)/2 - float64(ypos))

	// You can't look further than up or down
	if p.Elevation < -math.Pi/2 {
		p.Elevation = -math.Pi/2 + 0.0001 // prevent singularity
	} else if p.Elevation > math.Pi/2 {
		p.Elevation = math.Pi/2 - 0.0001
	}
}
