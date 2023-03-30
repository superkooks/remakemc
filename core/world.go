package core

// A Chunk is a 16x16x16 group of blocks
type Chunk struct {
	Position Vec3
	Blocks   [][][]Block // x,y,z

	// Render data
	Mesh []float32
	VAO  uint32
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
	Position Vec3
	Data     interface{}
	Type     *BlockType
}

type BlockType struct {
	RenderFunc func()
}

type Dimension struct {
	// Chunks are all the loaded chunks, addressed by their starting coordinates.
	// A chunk starts at (0,0,0) and ends at (16,16,16)
	Chunks map[Vec3]*Chunk
}

func (d *Dimension) GetChunkContaining(pos Vec3) *Chunk {
	return d.Chunks[pos.Div(16)]
}

func (d *Dimension) GetBlockAt(pos Vec3) Block {
	chk := d.Chunks[NewVec3(
		FlooredDivision(pos.X, 16),
		FlooredDivision(pos.Y, 16),
		FlooredDivision(pos.Z, 16),
	)]

	if chk == nil {
		return Block{}
	}

	return chk.Blocks[FlooredRemainder(pos.X, 16)][FlooredRemainder(pos.Y, 16)][FlooredRemainder(pos.Z, 16)]
}

func (d *Dimension) SetBlockAt(b Block) {
	chk := d.Chunks[NewVec3(
		FlooredDivision(b.Position.X, 16),
		FlooredDivision(b.Position.Y, 16),
		FlooredDivision(b.Position.Z, 16),
	)]

	if chk == nil {
		return
	}

	x := FlooredRemainder(b.Position.X, 16)
	y := FlooredRemainder(b.Position.Y, 16)
	z := FlooredRemainder(b.Position.Z, 16)
	chk.Blocks[x][y][z] = b
}

type World struct {
	Dimensions map[string]Dimension
}
