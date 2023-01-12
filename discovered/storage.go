package discovered

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ServiceStorage interface {
	Add(service Service) error
	Remove(serviceId uuid.UUID) error
	Get(serviceName string) (*Service, error)
	GetById(serviceId uuid.UUID) (*Service, error)
	GetAllServices() ([]Service, error)
	UpdateLastHeartBeat(serviceName string, newTime time.Time) error
	HandleUnHealthyServices()
}

type inMemoryServiceStorage struct {
	services []Service
	lock     sync.RWMutex
}

func (s *inMemoryServiceStorage) Add(service Service) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if serv, err := s.Get(service.Name); err != nil {
		s.services = append(s.services, service)
		s.deleteIfOlderThan3Min(service.id)
		return nil
	} else {
		return fmt.Errorf("[err] service %s already registered!", serv.Name)
	}
}

func (s *inMemoryServiceStorage) Remove(serviceId uuid.UUID) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	for index, service := range s.services {
		if service.id == serviceId {
			s.services[index] = s.services[len(s.services)-1]
			s.services = s.services[:len(s.services)-1]
			return nil
		}
	}
	return fmt.Errorf("[err] service with id = %v not found", serviceId)
}

func (s *inMemoryServiceStorage) Get(serviceName string) (*Service, error) {
	for i := 0; i < len(s.services); i++ {
		service := &s.services[i]
		if service.Name == serviceName {
			return service, nil
		}
	}
	return &Service{}, fmt.Errorf("[err] service %s not found!", serviceName)
}

func (s *inMemoryServiceStorage) GetById(serviceId uuid.UUID) (*Service, error) {
	for i := 0; i < len(s.services); i++ {
		service := &s.services[i]
		if service.id == serviceId {
			return service, nil
		}
	}
	return &Service{}, fmt.Errorf("[err] service %s not found!", serviceId)
}

func (s *inMemoryServiceStorage) GetAllServices() ([]Service, error) {
	if s.services != nil {
		return s.services, nil
	}
	return nil, errors.New("[err] there arent any discovered services")
}

func (s *inMemoryServiceStorage) UpdateLastHeartBeat(serviceName string, newTime time.Time) error {
	serv, err := s.Get(serviceName)
	if err != nil {
		return err
	}
	serv.LastHeartBeatCheck = newTime
	return nil
}
func (s *inMemoryServiceStorage) HandleUnHealthyServices() {
	for i := 0; i < len(s.services); i++ {
		service := &s.services[i]
		s.deleteIfOlderThan3Min(service.id)
	}
}

func (s *inMemoryServiceStorage) deleteIfOlderThan3Min(serviceId uuid.UUID) {
	log.Printf("Starting deletion procedure for unhealthy services %s 1/3min\n", serviceId)
	duration := 3 * time.Minute
	quit := make(chan bool)
	go func(quit chan bool) {
		ticker := time.NewTicker(duration)
		for tick := range ticker.C {
			select {
			case <-quit:
				return
			default:
				s.performDeletion(quit, serviceId, duration, tick)
			}
		}
	}(quit)

}

func (s *inMemoryServiceStorage) performDeletion(quit chan bool, serviceId uuid.UUID, duration time.Duration, tick time.Time) {
	service, err := s.GetById(serviceId)
	if err == nil && service.LastHeartBeatCheck.Add(duration).Before(tick) {
		log.Printf("Service %s unhealthy -> performing deletion\n", serviceId)
		err := s.Remove(serviceId)
		if err != nil {
			log.Println(err)
		} else {
			quit <- true
		}
	}
}

func NewInMemoryServiceStorage() ServiceStorage {
	return &inMemoryServiceStorage{}
}
