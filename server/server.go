package server

import (
	"log"
	"time"

	"github.com/ygaros/discovery-server/discover"
	"github.com/ygaros/discovery-server/dto"
)

const DEFAULT_PORT = 7654
const DEFAULT_PORT_FOR_UI = 7655
const TIME_FORMAT = "2006-01-02 15:04:05.999999999 -0700 MST"

type DiscoveryService interface {
	AddService(service dto.Service) error
	ListServices() ([]dto.ServiceHeartBeat, error)
	HeartBeat(service dto.Service) error
	GetService(serviceName string) (dto.ServiceHeartBeat, error)
}
type discoveryService struct {
	storage discover.Storage
}

func (s *discoveryService) AddService(service dto.Service) error {
	newService := discover.NewService(
		service.Name,
		service.Url,
		service.Secure,
	)
	err := s.storage.Add(newService)
	return err
}

func (s *discoveryService) ListServices() ([]dto.ServiceHeartBeat, error) {
	var parsedService []dto.ServiceHeartBeat
	var err error
	if services, err := s.storage.GetAllServices(); err == nil {
		for _, service := range services {
			parsedService = append(parsedService, dto.ServiceHeartBeat{
				Name:          service.Name,
				Url:           service.Url,
				LastHeartBeat: service.LastHeartBeatCheck,
			})
		}
	}

	return parsedService, err
}

func (s *discoveryService) HeartBeat(service dto.Service) error {
	savedService, err := s.storage.GetByUrl(discover.PrepareUrl(service.Url, service.Secure))
	if err != nil {
		log.Println(err)
		return err
	}
	err = s.storage.UpdateLastHeartBeat(*savedService, time.Now())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *discoveryService) GetService(serviceName string) (dto.ServiceHeartBeat, error) {
	if service, err := s.storage.Get(serviceName); err == nil {
		return dto.ServiceHeartBeat{
			Name:          service.Name,
			Url:           service.Url,
			LastHeartBeat: service.LastHeartBeatCheck,
		}, err
	} else {
		return dto.ServiceHeartBeat{}, err
	}
}

//MultiMap implementation that allow multiple instances of the same service
func NewDiscoveryServiceWithInMemoryStorage() DiscoveryService {
	return &discoveryService{
		storage: discover.NewMultiMapStorage(),
	}
}

//Slice implementation that doesn't allow multiple instances of the same service
func NewDiscoveryServiceWithSliceStorage() DiscoveryService {
	return &discoveryService{
		storage: discover.NewInMemoryStorage(),
	}
}
