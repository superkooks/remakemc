package server

import (
	"fmt"
	"remakemc/config"
	"remakemc/core"
	"remakemc/core/container"
	"remakemc/core/items"
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

	// Set the inventory
	c.Inventory = new(container.Inventory)
	c.Inventory.Init(false, msg.Player.EntityID)
	c.Inventory.Slots[5].SetStack(core.ItemStack{Item: items.Cobblestone.Name, Count: 64})
	c.Inventory.Slots[6].SetStack(core.ItemStack{Item: items.Cobblestone.Name, Count: 64})
	c.Inventory.Slots[7].SetStack(core.ItemStack{Item: items.Dirt.Name, Count: 64})
	c.Inventory.Slots[0].SetStack(core.ItemStack{Item: items.Furnace.Name, Count: 1})

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

	msg.Inventory = core.GetStacksFromSlots(c.Inventory.GetSlots())

	c.SendQueue <- proto.PLAY
	c.SendQueue <- msg

	// Update all clients
	// fmt.Println("no. of clients", len(clients))
	// for _, v := range clients {
	// 	if v != c {
	// 		fmt.Println("updating old client", c.Position.EntityID, "of new player", v.Position.EntityID)
	// 		v.SendQueue <- proto.ENTITY_CREATE
	// 		v.SendQueue <- proto.EntityCreate{
	// 			EntityPosition: msg.Player,
	// 			EntityType:     "mc:remoteplayer",
	// 		}
	// 	}
	// }

	// // Update client of all entities
	// for _, v := range clients {
	// 	if v != c {
	// 		fmt.Println("updating new client of player", v.Position.EntityID)
	// 		c.SendQueue <- proto.ENTITY_CREATE
	// 		c.SendQueue <- proto.EntityCreate{
	// 			EntityPosition: proto.EntityPosition(v.Position),
	// 			EntityType:     "mc:remoteplayer",
	// 		}
	// 	}
	// }
	// for _, v := range Dim.Entities {
	// 	c.SendQueue <- proto.ENTITY_CREATE
	// 	c.SendQueue <- proto.EntityCreate{
	// 		EntityPosition: proto.EntityPosition{
	// 			EntityID:      v.ID,
	// 			Yaw:           v.Yaw,
	// 			Pitch:         v.Pitch,
	// 			AABB:          v.AABB,
	// 			LookAzimuth:   v.LookAzimuth,
	// 			LookElevation: v.LookElevation,
	// 		},
	// 		EntityType: v.EntityType,
	// 	}
	// }
}

// func (c *Client) HandlePlayerPosition(p proto.PlayerPosition) {
// 	// TODO Check whether the player's position is valid
// 	// TODO Rubberband player

// 	p.EntityID = c.OldPosition.EntityID

// 	c.OldPosition = c.Position
// 	c.Position = p

// 	Dim.Lock.Lock()
// 	if Dim.GetChunkContaining(core.NewVec3FromFloat(c.OldPosition.Position)) !=
// 		Dim.GetChunkContaining(core.NewVec3FromFloat(c.Position.Position)) {
// 		chunkPos := core.NewVec3(
// 			core.FlooredDivision(core.FloorFloat32(p.Position.X()), 16)*16,
// 			0,
// 			core.FlooredDivision(core.FloorFloat32(p.Position.Z()), 16)*16,
// 		)

// 		// Check whether we need to unload any chunks
// 		var unloadChunks []core.Vec3
// 		for _, v := range c.loadedChunks {
// 			if v.X-chunkPos.X < (-config.App.RenderDistance-2)*16 || v.X-chunkPos.X > (config.App.RenderDistance+2)*16 ||
// 				v.Z-chunkPos.Z < (-config.App.RenderDistance-2)*16 || v.Z-chunkPos.Z > (config.App.RenderDistance+2)*16 {
// 				unloadChunks = append(unloadChunks, v)
// 			}
// 		}

// 		// Check whether we need to load any chunks.
// 		// NB: We always load an entire chunk column
// 		var newChunks []core.Vec3
// 		var allChunks []core.Vec3
// 		for x := -config.App.RenderDistance - 2; x < config.App.RenderDistance+2; x++ {
// 			for z := -config.App.RenderDistance - 2; z < config.App.RenderDistance+2; z++ {
// 				for y := 0; y < 16; y++ {
// 					allChunks = append(allChunks, core.NewVec3(x*16+chunkPos.X, y*16, z*16+chunkPos.Z))
// 				}

// 				var found bool
// 				for _, v := range c.loadedChunks {
// 					if v.X-chunkPos.X == x*16 && v.Z-chunkPos.Z == z*16 {
// 						found = true
// 						break
// 					}
// 				}
// 				if !found {
// 					for y := 0; y < 16; y++ {
// 						newChunks = append(newChunks, core.NewVec3(x*16+chunkPos.X, y*16, z*16+chunkPos.Z))
// 					}
// 				}
// 			}
// 		}

// 		if len(newChunks) != 0 {
// 			fmt.Println(newChunks)
// 			fmt.Println("writing load chunks")
// 			var chunks []*core.Chunk
// 			for _, v := range newChunks {
// 				chunks = append(chunks, GetChunkOrGen(v))
// 			}

// 			c.SendQueue <- proto.LOAD_CHUNKS
// 			c.SendQueue <- proto.NewLoadChunks(chunks)
// 		}
// 		if len(unloadChunks) != 0 {
// 			fmt.Println("writing unload chunks")
// 			c.SendQueue <- proto.UNLOAD_CHUNKS
// 			c.SendQueue <- unloadChunks
// 		}

// 		c.loadedChunks = allChunks
// 	}
// 	Dim.Lock.Unlock()

// 	// Update all other clients
// 	for _, v := range clients {
// 		if v != c {
// 			v.SendQueue <- proto.ENTITY_POSITION
// 			v.SendQueue <- p
// 		}
// 	}

// }

// func (c *Client) HandleBlockInteraction(b proto.BlockInteraction) {
// 	// TODO Check for reach

// 	// Check whether the player is inside the block
// 	// Nasty hack to use entity function

// 	selectedSlot := c.Inventory.GetSlots()[c.HotbarSlotSelected]

// 	if selectedSlot.GetStack().IsEmpty() {
// 		// TODO Error
// 		return
// 	}

// 	Dim.Lock.Lock()

// 	// Is the block clicked interactable?
// 	old := Dim.GetBlockAt(b.Position)
// 	if old.Type.LinkWithEntity != "" {
// 		// Find the entity for this block
// 		for _, v := range Dim.Entities {
// 			if v.Position == b.Position.ToFloat() &&
// 				core.BlockRegistry[old.Type.Name].LinkWithEntity == v.EntityType {

// 				// This has an entity, try to interact

// 			}
// 		}
// 	}

// 	// Place the block
// 	face := core.FaceFromSubvoxel(b.SubvoxelHit)
// 	newBlock := core.Block{
// 		Position: b.Position.Add(core.FaceDirection[face]),
// 		// Type:     core.BlockRegistry["mc:cobblestone"],
// 		Type: core.BlockRegistry[selectedSlot.GetStack().Item], // TODO item interact
// 	}
// 	Dim.SetBlockAt(newBlock)

// 	// See whether the entity will intersect with this block
// 	e := &core.Entity{Position: c.Position.Position, AABB: mgl32.Vec3{0.6, 1.8, 0.6}}
// 	if _, intersects := e.GetBlockIntersecting(Dim); intersects {
// 		// Remove the block
// 		Dim.SetBlockAt(core.Block{Position: newBlock.Position, Type: nil})

// 		Dim.Lock.Unlock()
// 		return
// 	}

// 	linkEntityType := core.BlockRegistry[selectedSlot.GetStack().Item].LinkWithEntity
// 	if linkEntityType != "" {
// 		// Create the linked entity
// 		e := &core.Entity{
// 			ID:         uuid.New(),
// 			Position:   newBlock.Position.ToFloat(),
// 			EntityType: linkEntityType,
// 			IsBlock:    true,
// 		}

// 		Dim.Entities = append(Dim.Entities, e)
// 		for _, v := range clients {
// 			v.SendQueue <- proto.ENTITY_CREATE
// 			v.SendQueue <- proto.EntityCreate{
// 				EntityPosition: proto.EntityPosition{
// 					EntityID:      e.ID,
// 					Position:      e.Position,
// 					Yaw:           e.Yaw,
// 					Pitch:         e.Pitch,
// 					AABB:          e.AABB,
// 					LookAzimuth:   e.LookAzimuth,
// 					LookElevation: e.LookElevation,
// 				},
// 				EntityType: e.EntityType,
// 			}
// 		}
// 	}

// 	// Decrement itemstack and update clint
// 	s := selectedSlot.GetStack()
// 	s.Count--
// 	if s.Count == 0 {
// 		selectedSlot.SetStack(core.ItemStack{})
// 	} else {
// 		selectedSlot.SetStack(s)
// 	}

// 	c.SendQueue <- proto.CONTAINER_CONTENTS
// 	c.SendQueue <- proto.ContainerContents{
// 		Slots:         core.GetStacksFromSlots(c.Inventory.GetSlots()),
// 		FloatingStack: core.ItemStack{},
// 	}

// 	// If needed update player's equipment
// 	if selectedSlot.GetStack().Item == "" {
// 		for _, v := range clients {
// 			if v != c {
// 				v.SendQueue <- proto.ENTITY_EQUIPMENT
// 				v.SendQueue <- proto.EntityEquipment{
// 					EntityID: c.Position.EntityID,
// 					EntityEquipment: core.EntityEquipment{
// 						HeldItemType: selectedSlot.GetStack().Item,
// 					},
// 				}
// 			}
// 		}
// 	}

// 	// Update clients if the chunk is loaded for them
// 	chunkPos := Dim.GetChunkContaining(newBlock.Position).Position

// 	for _, v := range clients {
// 		for _, u := range v.loadedChunks {
// 			if u == chunkPos {
// 				// Update the client
// 				v.SendQueue <- proto.BLOCK_UPDATE
// 				v.SendQueue <- proto.BlockUpdate{
// 					Position:  newBlock.Position,
// 					BlockType: newBlock.Type.Name,
// 				}
// 				break
// 			}
// 		}
// 	}

// 	Dim.Lock.Unlock()
// }

// func (c *Client) HandleBlockDig(b proto.BlockDig) {
// 	// TODO Check for reach
// 	// TODO survival mining

// 	Dim.Lock.Lock()

// 	oldBlock := Dim.GetBlockAt(b.Position)

// 	newBlock := core.Block{Position: b.Position, Type: nil}
// 	Dim.SetBlockAt(newBlock)

// 	// Delete any linked entities
// 	for k, v := range Dim.Entities {
// 		if v.Position == b.Position.ToFloat() &&
// 			core.BlockRegistry[oldBlock.Type.Name].LinkWithEntity == v.EntityType {

// 			Dim.Entities = append(Dim.Entities[:k], Dim.Entities[k+1:]...)

// 			for _, w := range clients {
// 				w.SendQueue <- proto.ENTITY_DELETE
// 				w.SendQueue <- v.ID
// 			}
// 			break
// 		}
// 	}

// 	// Update clients if the chunk is loaded for them
// 	chunkPos := Dim.GetChunkContaining(b.Position).Position

// 	for _, v := range clients {
// 		for _, u := range v.loadedChunks {
// 			if u == chunkPos {
// 				// Update the client
// 				v.SendQueue <- proto.BLOCK_UPDATE
// 				v.SendQueue <- proto.BlockUpdate{
// 					Position:  newBlock.Position,
// 					BlockType: "",
// 				}
// 				break
// 			}
// 		}
// 	}

// 	Dim.Lock.Unlock()
// }
