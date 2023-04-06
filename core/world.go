package core

import (
	"github.com/go-gl/mathgl/mgl32"
)

// A Chunk is a 16x16x16 group of blocks
type Chunk struct {
	Position Vec3
	Blocks   [][][]Block // x,y,z

	// Render data
	MeshLen int
	VAO     uint32
}

type BlockFace int

const (
	FaceTop BlockFace = iota
	FaceBottom
	FaceLeft
	FaceRight
	FaceFront
	FaceBack
)

type Block struct {
	Type *BlockType
}

type BlockType struct {
	// If the block is transparent, then faces of another block that overlap
	// with the faces of this block will still be rendered anyway.
	//
	// This value should be true if the player can see through the block in
	// any way, or if the block does not take up the full area.
	Transparent bool
	RenderType  RenderType
}

type RenderType interface {
	// Load textures into atlas, etc.
	Init()

	// RenderFace should returns the data used for rendering, with the vertices
	// returned in block space. i.e. from (0,0,0) to (1,1,1)
	RenderFace(face BlockFace, pos mgl32.Vec3) (verts, normals, uvs []float32) // James Wray
}

type Dimension struct {
	// Chunks are all the loaded chunks, addressed by their starting coordinates.
	// A chunk starts at (0,0,0) and ends at (16,16,16)
	Chunks map[Vec3]*Chunk
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
	return chk.Blocks[x][y][z]
}

func (d *Dimension) SetBlockAt(b Block, pos Vec3) {
	chk := d.Chunks[NewVec3(
		FlooredDivision(pos.X, 16)*16,
		FlooredDivision(pos.Y, 16)*16,
		FlooredDivision(pos.Z, 16)*16,
	)]

	if chk == nil {
		return
	}

	x := FlooredRemainder(pos.X, 16)
	y := FlooredRemainder(pos.Y, 16)
	z := FlooredRemainder(pos.Z, 16)
	chk.Blocks[x][y][z] = b
}

type World struct {
	Dimensions map[string]Dimension
}
