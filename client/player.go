package client

import (
	"math"
	"remakemc/client/renderers"
	"remakemc/core"
	"remakemc/core/proto"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Player struct {
	core.Entity

	Sprinting bool
	Sneaking  bool

	Speed           float64
	MouseSensitivty float64
}

func NewPlayer(position mgl32.Vec3) *Player {
	p := &Player{Speed: 1, Entity: core.Entity{NoGravity: true, Position: position, AABB: mgl32.Vec3{0.6, 1.8, 0.6}}}
	return p
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
		float32(math.Cos(p.LookElevation) * math.Sin(p.LookAzimuth)),
		float32(math.Sin(p.LookElevation)),
		float32(math.Cos(p.LookElevation) * math.Cos(p.LookAzimuth)),
	}
}

// The direction the player would move if they pressed W
func (p *Player) ForwardVec() mgl32.Vec3 {
	return p.RightVec().Cross(mgl32.Vec3{0, -1, 0})
}

func (p *Player) RightVec() mgl32.Vec3 {
	return mgl32.Vec3{
		float32(math.Sin(p.LookAzimuth - math.Pi/2.0)),
		0,
		float32(math.Cos(p.LookAzimuth - math.Pi/2.0)),
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

// Process the keyboard input and physics for this frame
func (p *Player) DoTick() {
	p.Entity.DoTick()

	// Jump
	var jumping bool
	if renderers.Win.GetKey(glfw.KeySpace) == glfw.Press && p.OnGround() {
		// TODO Add timeout to next jump
		p.Velocity[1] = 8.4
		jumping = true
		serverWrite.Encode(proto.PLAYER_JUMP)
	}

	// Sneak
	if renderers.Win.GetKey(glfw.KeyLeftShift) == glfw.Press {
		// Nested for a reason
		if !p.Sneaking {
			p.SetSneaking(true)
			serverWrite.Encode(proto.PLAYER_SNEAKING)
			serverWrite.Encode(proto.PlayerSneaking(true))
		}
	} else if p.Sneaking {
		p.SetSneaking(false)
		serverWrite.Encode(proto.PLAYER_SNEAKING)
		serverWrite.Encode(proto.PlayerSneaking(false))
	}

	// Sprint
	if renderers.Win.GetKey(glfw.KeyLeftControl) == glfw.Press && !p.Sprinting {
		p.Sprinting = true
		serverWrite.Encode(proto.PLAYER_SPRINTING)
		serverWrite.Encode(proto.PlayerSprinting(true))
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
		serverWrite.Encode(proto.PLAYER_SPRINTING)
		serverWrite.Encode(proto.PlayerSprinting(false))
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
	direction := p.LookAzimuth
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
			p.Velocity[0] += 0.2 * float32(math.Sin(p.LookAzimuth)) * 20 * 0.91 * 0.6
			p.Velocity[2] += 0.2 * float32(math.Cos(p.LookAzimuth)) * 20 * 0.91 * 0.6
		}

	} else if p.OnGround() {
		p.Velocity[0] += groundAccel * float32(math.Sin(direction))
		p.Velocity[2] += groundAccel * float32(math.Cos(direction))

	} else {
		p.Velocity[0] += 0.02 * float32(moveMult*math.Sin(direction)*20)
		p.Velocity[2] += 0.02 * float32(moveMult*math.Cos(direction)*20)
	}

	serverWrite.Encode(proto.PLAYER_POSITION)
	serverWrite.Encode(proto.PlayerPosition{
		Position:      p.Position,
		Yaw:           p.Yaw,
		LookAzimuth:   p.LookAzimuth,
		LookElevation: p.LookElevation,
	})
}

// Process the mouse input for this frame
func (p *Player) ProcessMousePosition(deltaT float64) {
	// Get and reset cursor position
	xpos, ypos := renderers.Win.GetCursorPos()
	width, height := renderers.Win.GetSize()
	renderers.Win.SetCursorPos(float64(width)/2, float64(height)/2)

	// Calculate new orientation
	p.LookAzimuth += 0.05 * deltaT * (float64(width)/2 - float64(xpos))
	p.LookElevation += 0.05 * deltaT * (float64(height)/2 - float64(ypos))

	// You can't look further than up or down
	if p.LookElevation < -math.Pi/2 {
		p.LookElevation = -math.Pi/2 + 0.0001 // prevent singularity
	} else if p.LookElevation > math.Pi/2 {
		p.LookElevation = math.Pi/2 - 0.0001
	}
}
