package proto

import (
	"bytes"
	"remakemc/core"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/pierrec/lz4"
	"github.com/vmihailenco/msgpack/v5"
)

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

	LookAzimuth   float64
	LookElevation float64
}

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
