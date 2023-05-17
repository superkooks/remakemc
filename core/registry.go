package core

var BlockRegistry = map[string]*BlockType{
	"": nil,
}

var EntityRegistry = map[string]*EntityType{}

var ItemRegistry = map[string]*ItemType{}

func AddBlockToRegistry(b *BlockType) *BlockType {
	BlockRegistry[b.Name] = b
	return b
}

func AddEntityToRegistry(e *EntityType) *EntityType {
	EntityRegistry[e.Name] = e
	return e
}

func AddItemToRegistry(i *ItemType) *ItemType {
	ItemRegistry[i.Name] = i
	return i
}
