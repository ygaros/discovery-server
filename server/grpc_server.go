package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/ygaros/discovery-server/dto"
	proto "github.com/ygaros/discovery-server/gen/proto"

	"google.golang.org/grpc"
)

const DEFAULT_PORT = 7654
const TIME_FORMAT = "2006-01-02 15:04:05.999999999 -0700 MST"

type GrpcServer interface {
	Serve(port int) error
	ServeDefaultPort() error
}
type grpcServer struct {
	proto.UnimplementedDiscoveryServer
	dservice DiscoveryService
}

func (gs *grpcServer) AddService(ctx context.Context, request *proto.Service) (*proto.Empty, error) {
	service := dto.ToService(request)
	log.Printf("processing registering service %v\n", service)
	return &proto.Empty{}, gs.dservice.AddService(service)
}
func (gs *grpcServer) ListServices(ctx context.Context, request *proto.Empty) (response *proto.ListServiceResponse, err error) {
	var parsedServices []*proto.ServiceWithHeartBeat
	response = &proto.ListServiceResponse{}
	services, err := gs.dservice.ListServices()
	log.Println("processing get request for all registered services")
	if err != nil {
		return response, err
	}
	for _, service := range services {
		parsedServices = append(parsedServices, &proto.ServiceWithHeartBeat{
			Name:          service.Name,
			Url:           service.Url,
			LastHeartBeat: service.LastHeartBeat.Format(TIME_FORMAT),
		})
	}
	response.Services = parsedServices
	return response, nil
}

func (gs *grpcServer) HeartBeat(ctx context.Context, request *proto.Service) (*proto.Empty, error) {
	log.Printf("processing heartbeating on %s\n", request.Name)
	return &proto.Empty{}, gs.dservice.HeartBeat(dto.ToService(request))
}

func (gs *grpcServer) GetService(ctx context.Context, request *proto.GetServiceRequest) (*proto.ServiceWithHeartBeat, error) {
	service, err := gs.dservice.GetService(request.GetServiceName())
	log.Printf("processing getting service data for %s\n", request.ServiceName)
	if err != nil {
		return &proto.ServiceWithHeartBeat{}, err
	}
	return &proto.ServiceWithHeartBeat{
		Name:          service.Name,
		Url:           service.Url,
		LastHeartBeat: service.LastHeartBeat.Format(TIME_FORMAT),
	}, nil
}

func (gs *grpcServer) Serve(port int) error {
	url := fmt.Sprintf("%s:%d", "localhost", port)
	log.Printf("Starting GRPC server on %s...\n", url)
	listen, err := net.Listen("tcp", url)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	proto.RegisterDiscoveryServer(grpcServer, gs)
	log.Println("Server started...")
	err = grpcServer.Serve(listen)
	if err != nil {
		return err
	}
	return nil
}
func (gs *grpcServer) ServeDefaultPort() error {
	return gs.Serve(DEFAULT_PORT)
}
func NewDiscoveryGrpcServerInMemoryStorage() GrpcServer {
	return &grpcServer{dservice: NewDiscoveryServiceWithInMemoryStorage()}
}
func NewDiscoveryGrpcServer(discoveryService *DiscoveryService) GrpcServer {
	return &grpcServer{dservice: *discoveryService}
}
