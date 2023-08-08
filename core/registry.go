package core

var BlockRegistry = map[string]*BlockType{
	"": nil,
}

var EntityRegistry = map[string]Entity{}

var ItemRegistry = map[string]*ItemType{}

func AddBlockToRegistry(b *BlockType) *BlockType {
	BlockRegistry[b.Name] = b
	return b
}

func AddEntityToRegistry(e Entity) Entity {
	EntityRegistry[e.GetTypeName()] = e
	return e
}

func AddItemToRegistry(i *ItemType) *ItemType {
	ItemRegistry[i.Name] = i
	return i
}
