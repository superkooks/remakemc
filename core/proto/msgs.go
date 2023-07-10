package proto

const (
	JOIN = iota
	PLAY

	ENTITY_CREATE
	ENTITY_POSITION

	PLAYER_POSITION
	PLAYER_JUMP
	PLAYER_SNEAKING
	PLAYER_SPRINTING

	BLOCK_UPDATE
	BLOCK_DIG
	BLOCK_INTERACTION

	UNLOAD_CHUNKS
	LOAD_CHUNKS

	PLAYER_HELD_ITEM
	ENTITY_EQUIPMENT
	CONTAINER_CONTENTS
)
