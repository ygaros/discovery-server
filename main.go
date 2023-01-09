package main

import (
	"log"
	"ygaros-discovery-server/server"
)

func main() {
	discoveryServer := server.NewDiscoveryServerInMemoryStorage()
	log.Fatalln(discoveryServer.Serve(0))
}
