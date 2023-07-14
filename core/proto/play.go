package proto

import (
	"bytes"
	"remakemc/core"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/pierrec/lz4"
	"github.com/vmihailenco/msgpack/v5"
)

// Start or finish digging a block.
// Digging must be started, then ended within the appropriate time for mining that block.
// Sent by clients
type BlockDig struct {
	Position      core.Vec3
	SubvoxelHit   mgl32.Vec3
	FinishDigging bool // start digging = false
}

// Message sent when a player right clicks on a block for any action.
// i.e. sent both when placing and interacting
// Sent by clients
type BlockInteraction struct {
	Position    core.Vec3 // the existing block that was interacted with
	SubvoxelHit mgl32.Vec3
}

// Updates a block within the render range of a client to a block type
// Sent by the server
type BlockUpdate struct {
	Position  core.Vec3
	BlockType string
}

// Indicates the desired state.
// Sent by clients
type PlayerSneaking bool

// Indicates the desired state.
// Sent by clients
type PlayerSprinting bool

// PlayerJump
// Has no type. Sent by clients

// Updates a player's position and rotations. EntityID will be ignored.
// Sent by clients
type PlayerPosition EntityPosition

// Updates the entity's position absolutely, as well as other movement-related values
// Sent by the server
type EntityPosition struct {
	EntityID uuid.UUID
	Position mgl32.Vec3
	Yaw      float64
	Pitch    float64
	AABB     mgl32.Vec3

	LookAzimuth   float64
	LookElevation float64
}

type EntityCreate struct {
	EntityPosition
	EntityType string
}

type EntityDelete uuid.UUID

// Instructs the client to unload the chunk
// Sent by the server
type UnloadChunks []core.Vec3

// Instructs the client to load and render the chunks listed.
// The message is an lz4 compressed blob of an array of chunks
// Sent by the server
type LoadChunks []byte

func NewLoadChunks(chunks []*core.Chunk) LoadChunks {
	b := new(bytes.Buffer)
	w := lz4.NewWriter(b)
	e := msgpack.NewEncoder(w)

	e.Encode(chunks)
	w.Close()
	return b.Bytes()
}

func (l LoadChunks) GetChunks() []*core.Chunk {
	r := lz4.NewReader(bytes.NewBuffer(l))
	d := msgpack.NewDecoder(r)

	var out []*core.Chunk
	d.Decode(&out)
	return out
}
