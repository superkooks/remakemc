package core

import (
	"time"

	"github.com/go-gl/mathgl/mgl32"
)

type LerpFace interface {
	GetLerpComp() *LerpComp
}

type LerpComp struct {
	lerpStartPos  mgl32.Vec3
	lerpEndPos    mgl32.Vec3
	lerpEndTime   time.Time
	lerpStartTime time.Time
}

func (p *LerpComp) GetLerpComp() *LerpComp {
	return p
}

func (e *LerpComp) NewLerp(end mgl32.Vec3) {
	e.lerpStartPos = e.lerpEndPos
	e.lerpStartTime = e.lerpEndTime
	e.lerpEndPos = end
	e.lerpEndTime = time.Now()
}

func LerpSystem(dim *Dimension) {
	dynamic := GetEntitiesSatisfying[interface {
		LerpFace
		PositionFace
	}](dim.Entities)

	for _, v := range dynamic {
		e := v.GetLerpComp()
		pos := v.GetPosition()

		lerpDelta := e.lerpEndTime.Sub(e.lerpStartTime)
		scalar := time.Since(e.lerpEndTime) / lerpDelta

		dir := e.lerpEndPos.Sub(e.lerpStartPos)
		*pos = dir.Mul(float32(scalar)).Add(e.lerpStartPos)
	}
}
