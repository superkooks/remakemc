package core

var BlockRegistry = map[string]*BlockType{
	"": nil,
}

var EntityRegistry = map[string]*EntityType{}

func AddBlockToRegistry(b *BlockType) *BlockType {
	BlockRegistry[b.Name] = b
	return b
}

func AddEntityToRegistry(e *EntityType) *EntityType {
	EntityRegistry[e.Name] = e
	return e
}
