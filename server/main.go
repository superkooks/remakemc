package server

import (
	"fmt"
	"net"
	"remakemc/core"
	"remakemc/core/proto"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Client struct {
	Conn      *net.TCPConn
	SendQueue chan interface{}
	encoder   *msgpack.Encoder

	Username    string
	Position    proto.PlayerPosition
	OldPosition proto.PlayerPosition

	loadedChunks []core.Vec3
}

var clients []*Client

func (c *Client) Listen() {
	d := msgpack.NewDecoder(c.Conn)
	c.encoder = msgpack.NewEncoder(c.Conn)
	c.SendQueue = make(chan interface{}, 32)
	for {
		var msgType int
		err := d.Decode(&msgType)
		if err != nil {
			panic(err)
		}

		switch msgType {
		case proto.JOIN:
			var j proto.Join
			err := d.Decode(&j)
			if err != nil {
				panic(err)
			}

			c.HandleJoin(j)

		case proto.PLAYER_JUMP:
			fmt.Println(c.Username, "jumped")

		case proto.PLAYER_SNEAKING:
			var s proto.PlayerSneaking
			err := d.Decode(&s)
			if err != nil {
				panic(err)
			}

			if s {
				fmt.Println(c.Username, "started sneaking")
			} else {
				fmt.Println(c.Username, "stopped sneaking")
			}

		case proto.PLAYER_SPRINTING:
			var s proto.PlayerSprinting
			err := d.Decode(&s)
			if err != nil {
				panic(err)
			}

			if s {
				fmt.Println(c.Username, "started sprinting")
			} else {
				fmt.Println(c.Username, "stopped sprinting")
			}

		case proto.PLAYER_POSITION:
			var p proto.PlayerPosition
			err := d.Decode(&p)
			if err != nil {
				panic(err)
			}

			c.HandlePlayerPosition(p)

		case proto.BLOCK_DIG:
			var b proto.BlockDig
			err := d.Decode(&b)
			if err != nil {
				panic(err)
			}

			c.HandleBlockDig(b)

		case proto.BLOCK_INTERACTION:
			var b proto.BlockInteraction
			err := d.Decode(&b)
			if err != nil {
				panic(err)
			}

			c.HandleBlockInteraction(b)
		}
	}
}

func Start(addr string) {
	// Generate terrain
	Dim.Lock.Lock()
	t := time.Now()
	for x := -16; x < 512+16; x += 16 {
		for z := -16; z < 512+16; z += 16 {
			GenTerrainColumn(core.NewVec3(x, 0, z), Dim)
		}
	}
	Dim.Lock.Unlock()
	fmt.Println("Generated initial terrain in", time.Since(t))

	go func() {
		// Start listening for connections
		a, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			panic(err)
		}

		l, err := net.ListenTCP("tcp", a)
		if err != nil {
			panic(err)
		}

		for {
			conn, err := l.AcceptTCP()
			if err != nil {
				panic(err)
			}

			c := &Client{Conn: conn}
			clients = append(clients, c)
			go c.Listen()
		}
	}()
}
