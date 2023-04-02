package main

import (
	"net/http"
	_ "net/http/pprof"
	"remakemc/client"
	"remakemc/config"
)

func main() {
	config.ParseConfig()

	if config.App.ServePprof {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}

	if !config.App.PublicServer {
		client.Start()
	} else {

	}
}
