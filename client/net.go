package client

import (
	"fmt"
	"remakemc/core/proto"

	"github.com/vmihailenco/msgpack/v5"
)

func readFromNet(serverRead chan interface{}) {
	d := msgpack.NewDecoder(conn)
	for {
		var msgType int
		err := d.Decode(&msgType)
		if err != nil {
			panic(err)
		}

		switch msgType {
		case proto.PLAY:
			var data proto.Play
			err = d.Decode(&data)
			if err != nil {
				panic(err)
			}
			serverRead <- data

		case proto.LOAD_CHUNKS:
			var data proto.LoadChunks
			err = d.Decode(&data)
			if err != nil {
				panic(err)
			}
			serverRead <- data

		case proto.UNLOAD_CHUNKS:
			var data proto.UnloadChunks
			err = d.Decode(&data)
			if err != nil {
				panic(err)
			}
			serverRead <- data

		case proto.ENTITY_CREATE:
			fmt.Println("received entity create")

			var data proto.EntityCreate
			err = d.Decode(&data)
			if err != nil {
				panic(err)
			}
			serverRead <- data

		case proto.ENTITY_DELETE:
			fmt.Println("received entity delete")

			var data proto.EntityDelete
			err = d.Decode(&data)
			if err != nil {
				panic(err)
			}
			serverRead <- data

		case proto.ENTITY_POSITION:
			var data proto.EntityPosition
			err = d.Decode(&data)
			if err != nil {
				panic(err)
			}
			serverRead <- data

		case proto.BLOCK_UPDATE:
			var data proto.BlockUpdate
			err = d.Decode(&data)
			if err != nil {
				panic(err)
			}
			serverRead <- data

		// case proto.ENTITY_EQUIPMENT:
		// 	var data proto.EntityEquipment
		// 	err = d.Decode(&data)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	serverRead <- data

		case proto.CONTAINER_CONTENTS:
			var data proto.ContainerContents
			err = d.Decode(&data)
			if err != nil {
				panic(err)
			}
			serverRead <- data

		default:
			panic("unknown packet type")
		}
	}
}

func writeFromQueue(queue chan interface{}) {
	e := msgpack.NewEncoder(conn)
	for {
		msg := <-queue
		err := e.Encode(msg)
		if err != nil {
			panic(err)
		}
	}
}
