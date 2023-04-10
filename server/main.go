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
	Conn    *net.TCPConn
	encoder *msgpack.Encoder
}

func (c *Client) Listen() {
	d := msgpack.NewDecoder(c.Conn)
	c.encoder = msgpack.NewEncoder(c.Conn)
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
		}
	}
}

func Start(addr string) {
	// Generate terrain
	DimLock.Lock()
	t := time.Now()
	for x := -16; x < 512+16; x += 16 {
		for z := -16; z < 512+16; z += 16 {
			GenTerrainColumn(core.NewVec3(x, 0, z), Dim)
		}
	}
	DimLock.Unlock()
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
			go c.Listen()
		}
	}()
}
