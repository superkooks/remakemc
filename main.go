package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"remakemc/client"
	"remakemc/config"
	"remakemc/server"
)

func main() {
	config.ParseConfig()

	if config.App.ServePprof {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}

	if config.App.PublicServer {
		server.Start(fmt.Sprint(config.App.Server.Address, ":", config.App.Server.Port))
		select {}
	} else {
		// server.Start("localhost:53785")
		client.Start()
	}
}
