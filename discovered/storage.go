package discovered

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

type ServiceStorage interface {
	Add(service Service) error
	Remove(serviceId uuid.UUID) error
	Get(serviceName string) (*Service, error)
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
		s.deleteIfOlderThan3Min(service)
		return nil
	} else {
		return errors.New(fmt.Sprintf("Service %s already registered!", serv.Name))
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
	return errors.New(fmt.Sprintf("Service with id = %v not found", serviceId))
}

func (s *inMemoryServiceStorage) Get(serviceName string) (*Service, error) {
	for i := 0; i < len(s.services); i++ {
		service := &s.services[i]
		if service.Name == serviceName {
			return service, nil
		}
	}
	return &Service{}, errors.New(fmt.Sprintf("Service %s not found!", serviceName))
}
func (s *inMemoryServiceStorage) GetAllServices() ([]Service, error) {
	if s.services != nil {
		return s.services, nil
	}
	return nil, errors.New("there arent any discovered services")
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
	for _, service := range s.services {
		s.deleteIfOlderThan3Min(service)
	}
}

func (s *inMemoryServiceStorage) deleteIfOlderThan3Min(service Service) {
	log.Printf("Starting deletion procedure for unhealthy services %s 1/3min\n", service.Name)
	duration := 3 * time.Minute
	quit := make(chan bool)
	go func(quit chan bool) {
		for range time.Tick(duration) {
			select {
			case <-quit:
				return
			default:
				s.performDeletion(quit, service, duration)
			}
		}
	}(quit)

}

func (s *inMemoryServiceStorage) performDeletion(quit chan bool, service Service, duration time.Duration) {
	if service.LastHeartBeatCheck.Add(duration).Before(time.Now()) {
		log.Printf("Service %s unhealthy -> performing deletion\n", service.Name)
		err := s.Remove(service.id)
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
