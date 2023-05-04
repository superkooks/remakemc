package core

import (
	"sync"

	"github.com/go-gl/mathgl/mgl32"
)

type BlockFace uint8

const (
	FaceTop BlockFace = iota
	FaceBottom
	FaceLeft
	FaceRight
	FaceFront
	FaceBack
)

var FaceDirection = map[BlockFace]Vec3{
	FaceTop:    NewVec3(0, 1, 0),
	FaceBottom: NewVec3(0, -1, 0),
	FaceLeft:   NewVec3(-1, 0, 0),
	FaceRight:  NewVec3(1, 0, 0),
	FaceFront:  NewVec3(0, 0, 1),
	FaceBack:   NewVec3(0, 0, -1),
}

type Block struct {
	Position Vec3
	Type     *BlockType
}

type BlockType struct {
	// The registered name of this block type.
	// It should be in the format of namespace:block
	Name string

	// If the block is transparent, then faces of another block that overlap
	// with the faces of this block will still be rendered anyway.
	//
	// This value should be true if the player can see through the block in
	// any way, or if the block does not take up the full area.
	Transparent bool
	RenderType  RenderBlockType
}

type RenderBlockType interface {
	// Load textures into atlas, etc.
	Init()

	// RenderFace should returns the data used for rendering, with the vertices
	// returned in block space. i.e. from (0,0,0) to (1,1,1)
	RenderFace(face BlockFace, pos mgl32.Vec3) (verts, normals, uvs []float32)
}

type Dimension struct {
	Lock *sync.RWMutex

	// Chunks are all the loaded chunks, addressed by their starting coordinates.
	// A chunk starts at (0,0,0) and ends at (16,16,16)
	Chunks map[Vec3]*Chunk

	Entities []*Entity
}

func (d *Dimension) GetChunkContaining(pos Vec3) *Chunk {
	return d.Chunks[NewVec3(
		FlooredDivision(pos.X, 16)*16,
		FlooredDivision(pos.Y, 16)*16,
		FlooredDivision(pos.Z, 16)*16,
	)]
}

func (d *Dimension) GetBlockAt(pos Vec3) Block {
	chk := d.Chunks[NewVec3(
		FlooredDivision(pos.X, 16)*16,
		FlooredDivision(pos.Y, 16)*16,
		FlooredDivision(pos.Z, 16)*16,
	)]

	if chk == nil {
		return Block{}
	}

	x := FlooredRemainder(pos.X, 16)
	y := FlooredRemainder(pos.Y, 16)
	z := FlooredRemainder(pos.Z, 16)
	return chk.GetBlockAt(NewVec3(x, y, z))
}

// Prevent a hashmap lookup if the pos falls within the given chunk
func (d *Dimension) GetBlockAtOptimised(pos Vec3, chunkGuess *Chunk) Block {
	var chk *Chunk
	if pos.X > chunkGuess.Position.X && pos.X < chunkGuess.Position.X+16 &&
		pos.Y > chunkGuess.Position.Y && pos.Y < chunkGuess.Position.Y+16 &&
		pos.Z > chunkGuess.Position.Z && pos.Z < chunkGuess.Position.Z+16 {
		chk = chunkGuess
	} else {
		chk = d.Chunks[NewVec3(
			FlooredDivision(pos.X, 16)*16,
			FlooredDivision(pos.Y, 16)*16,
			FlooredDivision(pos.Z, 16)*16,
		)]
	}

	if chk == nil {
		return Block{}
	}

	x := FlooredRemainder(pos.X, 16)
	y := FlooredRemainder(pos.Y, 16)
	z := FlooredRemainder(pos.Z, 16)
	return chk.GetBlockAt(NewVec3(x, y, z))
}

func (d *Dimension) SetBlockAt(b Block) {
	chk := d.Chunks[NewVec3(
		FlooredDivision(b.Position.X, 16)*16,
		FlooredDivision(b.Position.Y, 16)*16,
		FlooredDivision(b.Position.Z, 16)*16,
	)]

	if chk == nil {
		return
	}

	x := FlooredRemainder(b.Position.X, 16)
	y := FlooredRemainder(b.Position.Y, 16)
	z := FlooredRemainder(b.Position.Z, 16)
	chk.SetBlockAt(NewVec3(x, y, z), b)
}

type World struct {
	Dimensions map[string]Dimension
}

type Tickable interface {
	DoTick()
}

type RandomTickable interface {
	DoRandomTick()
}
