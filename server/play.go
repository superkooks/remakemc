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
	fmt.Println("join event")
	c.Username = j.Username

	// Reply with a play event
	var msg proto.Play
	msg.Player = proto.EntityPosition{
		EntityID: uuid.New(),
		Position: mgl32.Vec3{0, 95, 0},
	}
	c.OldPosition = proto.PlayerPosition(msg.Player)
	c.Position = proto.PlayerPosition(msg.Player)

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

	// Update all clients
	fmt.Println("no. of clients", len(clients))
	for _, v := range clients {
		if v != c {
			fmt.Println("updating old client", c.Position.EntityID, "of new player", v.Position.EntityID)
			v.encoder.Encode(proto.ENTITY_CREATE)
			v.encoder.Encode(proto.EntityCreate{
				EntityPosition: msg.Player,
				EntityType:     "mc:remoteplayer",
			})
		}
	}

	// Update client of all entities
	for _, v := range clients {
		if v != c {
			fmt.Println("updating new client of player", v.Position.EntityID)
			c.encoder.Encode(proto.ENTITY_CREATE)
			c.encoder.Encode(proto.EntityCreate{
				EntityPosition: proto.EntityPosition(v.Position),
				EntityType:     "mc:remoteplayer",
			})
		}
	}
}

func (c *Client) HandlePlayerPosition(p proto.PlayerPosition) {
	// TODO Check whether the player's position is valid
	// TODO Rubberband player

	p.EntityID = c.OldPosition.EntityID

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

	// Update all clients
	for _, v := range clients {
		if v != c {
			v.encoder.Encode(proto.ENTITY_POSITION)
			v.encoder.Encode(p)
		}
	}

}

func (c *Client) HandleBlockInteraction(b proto.BlockInteraction) {
	// TODO Check for reach

	Dim.Lock.Lock()

	// TODO Support interactions

	// Place the block
	face := core.FaceFromSubvoxel(b.SubvoxelHit)
	newBlock := core.Block{
		Position: b.Position.Add(core.FaceDirection[face]),
		Type:     core.BlockRegistry["mc:cobblestone"],
		// TODO INVENTORY Type: core.BlockRegistry[b.BlockType],
	}
	Dim.SetBlockAt(newBlock)

	// Update clients if the chunk is loaded for them
	chunkPos := Dim.GetChunkContaining(newBlock.Position).Position

	for _, v := range clients {
		for _, u := range v.loadedChunks {
			if u == chunkPos {
				// Update the client
				v.encoder.Encode(proto.BLOCK_UPDATE)
				v.encoder.Encode(proto.BlockUpdate{
					Position:  newBlock.Position,
					BlockType: newBlock.Type.Name,
				})
				break
			}
		}
	}

	Dim.Lock.Unlock()
}

func (c *Client) HandleBlockDig(b proto.BlockDig) {
	// TODO Check for reach
	// TODO survival mining

	Dim.Lock.Lock()

	newBlock := core.Block{Position: b.Position, Type: nil}
	Dim.SetBlockAt(newBlock)

	// Update clients if the chunk is loaded for them
	chunkPos := Dim.GetChunkContaining(b.Position).Position

	for _, v := range clients {
		for _, u := range v.loadedChunks {
			if u == chunkPos {
				// Update the client
				v.encoder.Encode(proto.BLOCK_UPDATE)
				v.encoder.Encode(proto.BlockUpdate{
					Position:  newBlock.Position,
					BlockType: "",
				})
				break
			}
		}
	}

	Dim.Lock.Unlock()
}
