package main

import (
	"remakemc/client"
	"remakemc/config"
)

func main() {
	config.ParseConfig()

	if !config.App.PublicServer {
		client.Start()
	} else {

	}
}
