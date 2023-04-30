package core

// A Chunk is a 16x16x16 group of blocks
type Chunk struct {
	Position Vec3

	// Render data
	MeshLen       int
	VAO           uint32
	VertexBuffers []uint32

	// Block data is stored using a palette, to save memory
	// NB: This data should never be accessed manually, except for
	//     (de)serialization
	BlockPalette []string
	BlockData    []byte

	// The number of bits used to represent the index into the palette
	// 4, 8, or 16
	PaletteBits int
}

// Returns an empty chunk
func NewChunk(pos Vec3) *Chunk {
	return &Chunk{
		Position: pos,

		BlockPalette: []string{""},
		BlockData:    make([]byte, 16*16*16/2),
		PaletteBits:  4,
	}
}

// Gets a block at the chunk local coordinate provided
func (c *Chunk) GetBlockAt(pos Vec3) Block {
	startBits := (pos.X*16*16 + pos.Y*16 + pos.Z) * c.PaletteBits
	start := startBits / 8
	end := startBits/8 + 1

	var ind int
	byts := c.BlockData[start:end]
	switch c.PaletteBits {
	case 4:
		if (startBits + 4) != end*8 {
			// Take the lowest bits
			ind = int(byts[0] & 0x0f)
		} else {
			// Take the highest bits
			ind = int(byts[0] >> 4)
		}
	case 8:
		ind = int(byts[0])
	case 16:
		ind = int(byts[1])<<8 | int(byts[0])
	default:
		panic("invalid number of palette bits")
	}

	return Block{Type: BlockRegistry[c.BlockPalette[ind]]}
}

func (c *Chunk) SetBlockAt(pos Vec3, bl Block) {
	// Determine if this would overflow the palette
	// This check isn't super thorough in the name of speed.
	// We will accept the occasional false positive.
	if len(c.BlockPalette)+1 >= 1<<c.PaletteBits {
		// Migrate the data to a larger palette
		if c.PaletteBits == 4 {
			// 4 bits --> 8 bits
			newData := make([]byte, len(c.BlockData)*2)
			for k, v := range c.BlockData {
				newData[k*2] = v & 0x0f
				newData[k*2+1] = v >> 4
			}

			c.PaletteBits = 8
		} else {
			// 8 bits --> 16 bits
			newData := make([]byte, len(c.BlockData)*2)
			for k, v := range c.BlockData {
				newData[k*2] = v
			}

			c.PaletteBits = 16
		}
	}

	// Get the index of the blocktype
	ind := -1
	for k, v := range c.BlockPalette {
		if (bl.Type == nil && v == "") || v == bl.Type.Name {
			ind = k
			break
		}
	}

	if ind == -1 {
		// Add the block type to the palette
		ind = len(c.BlockPalette)
		c.BlockPalette = append(c.BlockPalette, bl.Type.Name)
	}

	// Set the bits in the data
	startBits := (pos.X*16*16 + pos.Y*16 + pos.Z) * c.PaletteBits
	start := startBits / 8
	end := startBits/8 + 1
	switch c.PaletteBits {
	case 4:
		if (startBits + 4) != end*8 {
			// Set the lower bits
			masked := c.BlockData[start] & 0xf0
			c.BlockData[start] = masked | byte(ind)
		} else {
			// Set the higher bits
			masked := c.BlockData[start] & 0x0f
			c.BlockData[start] = masked | byte(ind)<<4
		}
	case 8:
		c.BlockData[start] = byte(ind)
	case 16:
		c.BlockData[start] = byte(ind & 0xff)
		c.BlockData[start+1] = byte(ind >> 8)
	}
}
