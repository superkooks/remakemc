package server

import (
	"math/rand"
	"remakemc/core"
	"remakemc/core/blocks"
	"sync"

	"github.com/aquilax/go-perlin"
)

const RAND_SEED = 1337
const WORLD_WIDTH = 65536

// The current loaded dimension
var Dim = &core.Dimension{Chunks: make(map[core.Vec3]*core.Chunk), Lock: new(sync.RWMutex)}

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
		chk := core.NewChunk(chunkPos.Add(core.Vec3{Y: cy * 16}))

		for x := 0; x < 16; x++ {
			for y := 0; y < 16; y++ {
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

					chk.SetBlockAt(core.NewVec3(x, y, z), bl)
				}
			}
		}

		dim.Chunks[chk.Position] = chk
	}
}

// You must lock Dim yourself
func GetChunkOrGen(pos core.Vec3) *core.Chunk {
	c, ok := Dim.Chunks[pos]
	if ok {
		return c
	}

	GenTerrainColumn(core.NewVec3(pos.X, 0, pos.Z), Dim)

	return Dim.Chunks[pos]
}
