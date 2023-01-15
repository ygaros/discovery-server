package discover

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type inMemoryStorage struct {
	services []Service
	lock     sync.RWMutex
}

func (s *inMemoryStorage) Add(service Service) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if serv, err := s.Get(service.Name); err != nil {
		s.services = append(s.services, service)
		s.deleteIfOlderThan(service, DELETION_TIME)
		return nil
	} else {
		return fmt.Errorf("[err] cannot replace old instance of that service %s", serv.Name)
	}
}

func (s *inMemoryStorage) Remove(serviceName string, serviceId uuid.UUID) error {
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

func (s *inMemoryStorage) Get(serviceName string) (*Service, error) {
	for i := 0; i < len(s.services); i++ {
		service := &s.services[i]
		if service.Name == serviceName {
			return service, nil
		}
	}
	return &Service{}, fmt.Errorf("[err] service %s not found", serviceName)
}

func (s *inMemoryStorage) GetById(serviceId uuid.UUID) (*Service, error) {
	for i := 0; i < len(s.services); i++ {
		service := &s.services[i]
		if service.id == serviceId {
			return service, nil
		}
	}
	return &Service{}, fmt.Errorf("[err] service %s not found", serviceId)
}

func (s *inMemoryStorage) GetByUrl(serviceUrl string) (*Service, error) {
	for i := 0; i < len(s.services); i++ {
		service := &s.services[i]
		if service.Url == serviceUrl {
			return service, nil
		}
	}
	return &Service{}, fmt.Errorf("[err] service with url %s not found", serviceUrl)
}

func (s *inMemoryStorage) GetAllServices() ([]Service, error) {
	if s.services != nil {
		return s.services, nil
	}
	return nil, errors.New("[err] there arent any discovered services")
}

func (s *inMemoryStorage) UpdateLastHeartBeat(service Service, newTime time.Time) error {
	serv, err := s.Get(service.Name)
	if err != nil {
		return err
	}
	serv.LastHeartBeatCheck = newTime
	return nil
}

func (s *inMemoryStorage) deleteIfOlderThan(service Service, duration time.Duration) {
	log.Printf("Starting deletion procedure for unhealthy services %s 1/1.5min\n", service.id)
	quit := make(chan bool)
	go func(quit chan bool) {
		ticker := time.NewTicker(duration)
		for tick := range ticker.C {
			select {
			case <-quit:
				return
			default:
				s.performDeletion(quit, service, duration, tick)
			}
		}
	}(quit)
}

func (s *inMemoryStorage) performDeletion(
	quit chan bool,
	service Service,
	duration time.Duration,
	tick time.Time) {
	saved, err := s.GetById(service.id)
	if err == nil && saved.LastHeartBeatCheck.Add(duration).Before(tick) {
		log.Printf("Service %s unhealthy -> performing deletion\n", service.id)
		err := s.Remove(service.Name, service.id)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Stoping deletion procedure for unhealthy services %s 1/1.5min\n", service.id)
			quit <- true
		}
	}
}

func NewInMemoryStorage() Storage {
	return &inMemoryStorage{}
}
