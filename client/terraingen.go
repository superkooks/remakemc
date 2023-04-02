package client

import (
	"math/rand"
	"remakemc/core"
	"remakemc/core/blocks"

	"github.com/aquilax/go-perlin"
)

const RAND_SEED = 1337
const WORLD_WIDTH = 65536

func GenTerrainColumn(chunkPos core.Vec3, dim *core.Dimension) {
	perl := perlin.NewPerlinRandSource(2, 64, 3, rand.NewSource(RAND_SEED))

	// Generate height map
	heightmap := make([][]int, 16)
	for x := range heightmap {
		heightmap[x] = make([]int, 16)
		for z := range heightmap[x] {
			heightmap[x][z] = int(perl.Noise2D(float64(x+chunkPos.X)/WORLD_WIDTH, float64(z+chunkPos.Z)/WORLD_WIDTH)*32) + 64
		}
	}

	// Generate chunk blocks
	for cy := 0; cy < 16; cy++ {
		chk := &core.Chunk{
			Position: chunkPos.Add(core.Vec3{Y: cy * 16}),
		}

		chk.Blocks = make([][][]core.Block, 16)
		for x := 0; x < 16; x++ {
			b := make([][]core.Block, 16)
			for y := 0; y < 16; y++ {
				a := make([]core.Block, 16)
				for z := 0; z < 16; z++ {
					height := heightmap[x][z]
					bl := core.Block{}
					if y+cy*16 < height-3 {
						// Stone
						bl.Type = blocks.Stone
					} else if y+cy*16 < height {
						// Dirt
						bl.Type = blocks.Dirt
					} else if y+cy*16 == height {
						// Grass
						bl.Type = blocks.Grass
					}

					a[z] = bl
				}

				b[y] = a
			}

			chk.Blocks[x] = b
		}

		dim.Chunks[chk.Position] = chk
	}
}
