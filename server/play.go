package server

import (
	"remakemc/core"
	"remakemc/core/proto"

	"github.com/go-gl/mathgl/mgl32"
)

func (c *Client) HandleJoin(j proto.Join) {
	// Reply with a play event
	var msg proto.Play
	msg.Player.Pos = mgl32.Vec3{5, 95, 5}
	msg.Player.LookAzimuth = 0
	msg.Player.LookElevation = 0
	msg.Player.Flying = false

	DimLock.Lock()
	var chunks []*core.Chunk
	for _, v := range Dim.Chunks {
		chunks = append(chunks, v)
	}
	msg.AddChunks(chunks)
	DimLock.Unlock()

	c.encoder.Encode(proto.PLAY)
	c.encoder.Encode(msg)
}
