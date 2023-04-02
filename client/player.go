package client

import (
	"math"
	"remakemc/client/renderers"
	"remakemc/core"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Player struct {
	core.Entity

	LookAzimuth   float64
	LookElevation float64

	Speed           float64
	MouseSensitivty float64
}

// The position of the camera
func (p *Player) CameraPos() mgl32.Vec3 {
	return p.Position.Add(mgl32.Vec3{0.5, 1.8, 0.5})
}

// The direction the player is looking
func (p *Player) LookVec() mgl32.Vec3 {
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

// Process the keyboard input for this frame
func (p *Player) ProcessKeyboard() {
	scal := float32(p.Speed)

	// Move forwards
	if renderers.Win.GetKey(glfw.KeyW) == glfw.Press {
		p.AddImpulse(core.Impulse{Force: p.ForwardVec().Mul(scal), Remaining: 0.01})
	}

	// Move backwards
	if renderers.Win.GetKey(glfw.KeyS) == glfw.Press {
		p.AddImpulse(core.Impulse{Force: p.ForwardVec().Mul(-scal), Remaining: 0.01})
	}

	// Move right
	if renderers.Win.GetKey(glfw.KeyD) == glfw.Press {
		p.AddImpulse(core.Impulse{Force: p.RightVec().Mul(scal), Remaining: 0.01})
	}

	// Move left
	if renderers.Win.GetKey(glfw.KeyA) == glfw.Press {
		p.AddImpulse(core.Impulse{Force: p.RightVec().Mul(-scal), Remaining: 0.01})
	}

	// Fly up
	if renderers.Win.GetKey(glfw.KeySpace) == glfw.Press {
		p.AddImpulse(core.Impulse{Force: mgl32.Vec3{0, 1, 0}.Mul(scal * 3), Remaining: 0.01})
	}

	// Fly down
	if renderers.Win.GetKey(glfw.KeyLeftShift) == glfw.Press {
		p.AddImpulse(core.Impulse{Force: mgl32.Vec3{0, 1, 0}.Mul(-scal * 3), Remaining: 0.01})
	}
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
