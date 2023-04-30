package server

import (
	"fmt"
	"remakemc/config"
	"remakemc/core"
	"remakemc/core/proto"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

func (c *Client) HandleJoin(j proto.Join) {
	c.Username = j.Username

	// Reply with a play event
	var msg proto.Play
	msg.Player = proto.EntityPosition{
		EntityID: uuid.New(),
		Position: mgl32.Vec3{0, 95, 0},
	}

	// Determine the chunks to load
	Dim.Lock.Lock()
	chunkPos := core.NewVec3(
		core.FlooredDivision(core.FloorFloat32(msg.Player.Position.X()), 16)*16,
		0,
		core.FlooredDivision(core.FloorFloat32(msg.Player.Position.Z()), 16)*16,
	)
	for x := -config.App.RenderDistance - 2; x < config.App.RenderDistance+2; x++ {
		for z := -config.App.RenderDistance - 2; z < config.App.RenderDistance+2; z++ {
			for y := 0; y < 16; y++ {
				c.loadedChunks = append(c.loadedChunks, core.NewVec3(x*16+chunkPos.X, y*16, z*16+chunkPos.Z))
			}
		}
	}

	var chunks []*core.Chunk
	for _, v := range c.loadedChunks {
		chunks = append(chunks, GetChunkOrGen(v))
	}
	msg.InitialChunks = proto.NewLoadChunks(chunks)
	Dim.Lock.Unlock()

	c.encoder.Encode(proto.PLAY)
	c.encoder.Encode(msg)
}

func (c *Client) HandlePlayerPosition(p proto.PlayerPosition) {
	c.OldPosition = c.Position
	c.Position = p

	Dim.Lock.Lock()
	if Dim.GetChunkContaining(core.NewVec3FromFloat(c.OldPosition.Position)) !=
		Dim.GetChunkContaining(core.NewVec3FromFloat(c.Position.Position)) {
		chunkPos := core.NewVec3(
			core.FlooredDivision(core.FloorFloat32(p.Position.X()), 16)*16,
			0,
			core.FlooredDivision(core.FloorFloat32(p.Position.Z()), 16)*16,
		)

		// Check whether we need to unload any chunks
		var unloadChunks []core.Vec3
		for _, v := range c.loadedChunks {
			if v.X-chunkPos.X < (-config.App.RenderDistance-2)*16 || v.X-chunkPos.X > (config.App.RenderDistance+2)*16 ||
				v.Z-chunkPos.Z < (-config.App.RenderDistance-2)*16 || v.Z-chunkPos.Z > (config.App.RenderDistance+2)*16 {
				unloadChunks = append(unloadChunks, v)
			}
		}

		// Check whether we need to load any chunks.
		// NB: We always load an entire chunk column
		var newChunks []core.Vec3
		var allChunks []core.Vec3
		for x := -config.App.RenderDistance - 2; x < config.App.RenderDistance+2; x++ {
			for z := -config.App.RenderDistance - 2; z < config.App.RenderDistance+2; z++ {
				for y := 0; y < 16; y++ {
					allChunks = append(allChunks, core.NewVec3(x*16+chunkPos.X, y*16, z*16+chunkPos.Z))
				}

				var found bool
				for _, v := range c.loadedChunks {
					if v.X-chunkPos.X == x*16 && v.Z-chunkPos.Z == z*16 {
						found = true
						break
					}
				}
				if !found {
					for y := 0; y < 16; y++ {
						newChunks = append(newChunks, core.NewVec3(x*16+chunkPos.X, y*16, z*16+chunkPos.Z))
					}
				}
			}
		}

		if len(newChunks) != 0 {
			fmt.Println(newChunks)
			fmt.Println("writing load chunks")
			var chunks []*core.Chunk
			for _, v := range newChunks {
				chunks = append(chunks, GetChunkOrGen(v))
			}

			c.encoder.Encode(proto.LOAD_CHUNKS)
			c.encoder.Encode(proto.NewLoadChunks(chunks))
		}
		if len(unloadChunks) != 0 {
			fmt.Println("writing unload chunks")
			c.encoder.Encode(proto.UNLOAD_CHUNKS)
			c.encoder.Encode(unloadChunks)
		}

		c.loadedChunks = allChunks
	}
	Dim.Lock.Unlock()
}
