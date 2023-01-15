package main

import (
	"github.com/ygaros/discovery-server/server"
)

func main() {
	discoveryService := server.NewDiscoveryServiceWithInMemoryStorage()

	grpcServer := server.NewDiscoveryGrpcServer(&discoveryService)
	httpServer := server.NewHttpDiscoveryServer(&discoveryService)
	// log.Fatalln(grpcServer.ServeDefaultPort())
	// log.Fatalln(httpServer.Serve(7655))
	go grpcServer.ServeDefaultPort()
	httpServer.Serve(7655)
}
