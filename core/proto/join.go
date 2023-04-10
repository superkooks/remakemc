package proto

import (
	"bytes"
	"fmt"
	"remakemc/core"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/pierrec/lz4"
	"github.com/vmihailenco/msgpack/v5"
)

// The first message sent by the client to the server.
// In response, a server will send the play event
// Sent by clients
type Join struct {
	Username string
}

// A reply to the Join event. Informs the client of all information needed to begin gameplay.
// Sent by the server
type Play struct {
	Player struct {
		Pos           mgl32.Vec3
		LookAzimuth   float64
		LookElevation float64
		Flying        bool
	}

	// Chunk data compressed using lz4
	CompressedChunks []byte
}

func (p *Play) AddChunks(chunks []*core.Chunk) {
	b := new(bytes.Buffer)
	w := lz4.NewWriter(b)
	e := msgpack.NewEncoder(w)

	t := time.Now()
	e.Encode(chunks)
	w.Close()
	p.CompressedChunks = b.Bytes()

	fmt.Println("Compressed chunks into", b.Len(), "bytes in", time.Since(t))
}

func (p *Play) GetChunks() []*core.Chunk {
	r := lz4.NewReader(bytes.NewBuffer(p.CompressedChunks))
	d := msgpack.NewDecoder(r)

	var out []*core.Chunk
	t := time.Now()
	d.Decode(&out)

	fmt.Println("Decompressed chunks in", time.Since(t))
	return out
}
