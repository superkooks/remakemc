package proto

// The first message sent by the client to the server.
// In response, a server will send the play event
// Sent by clients
type Join struct {
	Username string
}

// A reply to the Join event. Informs the client of all information needed to begin gameplay.
// Sent by the server
type Play struct {
	Player        EntityPosition
	InitialChunks LoadChunks
}
