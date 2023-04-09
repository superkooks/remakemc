package core

var BlockRegistry = map[string]*BlockType{
	"": nil,
}

func AddBlockToRegistry(b *BlockType) *BlockType {
	BlockRegistry[b.Name] = b
	return b
}
