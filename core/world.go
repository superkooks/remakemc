package core

// A Chunk is a 16x16x16 group of blocks
type Chunk struct {
	Position Vec3
	Blocks   [][][]Block // x,y,z

	// Render data
	Mesh []float32
	VAO  uint32
}

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
	g := func(i int) int {
		if i < 0 {
			return i/16 - 1
		}
		return i / 16
	}

	chk := d.Chunks[NewVec3(g(pos.X), g(pos.Y), g(pos.Z))]
	if chk == nil {
		return Block{}
	}

	f := func(i int) int {
		j := i % 16
		if j < 0 {
			j += 16
		}
		return j
	}

	return chk.Blocks[f(pos.X)][f(pos.Y)][f(pos.Z)]
}

type World struct {
	Dimensions map[string]Dimension
}
