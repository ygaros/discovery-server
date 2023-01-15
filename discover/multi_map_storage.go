package discover

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

type serviceMimified struct {
	id                 uuid.UUID
	url                string
	lastHeartBeatCheck time.Time
}

func toMimified(service Service) serviceMimified {
	return serviceMimified{
		id:                 service.id,
		url:                service.Url,
		lastHeartBeatCheck: service.LastHeartBeatCheck,
	}
}
func toService(servMini serviceMimified, name string) Service {
	return Service{
		Name:               name,
		id:                 servMini.id,
		Url:                servMini.url,
		LastHeartBeatCheck: servMini.lastHeartBeatCheck,
	}
}

type multiMapStorage struct {
	services map[string][]serviceMimified
	lock     sync.RWMutex
}

func (s *multiMapStorage) Add(service Service) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if !s.checkIfDuplicate(service) {
		if s.services[service.Name] == nil {
			s.services[service.Name] = make([]serviceMimified, 0)
		}
		s.services[service.Name] =
			append(s.services[service.Name], toMimified(service))
		s.deleteIfOlderThan(service, DELETION_TIME)
		return nil
	}
	// return nil
	return fmt.Errorf("[err] duplicate found for %s", service.Url)
}

func (s *multiMapStorage) Remove(serviceName string, serviceId uuid.UUID) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for name, services := range s.services {
		for index, service := range services {
			if service.id == serviceId {
				slice := s.services[name]
				slice[index] = slice[len(slice)-1]
				s.services[name] = slice[:len(slice)-1]
				return nil
			}
		}
	}
	return fmt.Errorf("[err] service %v doesnt exists", serviceId)
}

func (s *multiMapStorage) Get(serviceName string) (*Service, error) {
	services := s.services[serviceName]
	size := len(services)
	if size == 0 {
		return &Service{}, fmt.Errorf("[err] there arent any services %s", serviceName)
	}
	parsedService := toService(services[rand.Intn(size)], serviceName)
	return &parsedService, nil
}
func (s *multiMapStorage) GetById(serviceId uuid.UUID) (*Service, error) {
	for name, services := range s.services {
		for _, service := range services {
			if service.id == serviceId {
				parsed := toService(service, name)
				return &parsed, nil
			}
		}
	}
	return &Service{}, fmt.Errorf("[err] service %v doesnt exists", serviceId)
}

func (s *multiMapStorage) GetByUrl(serviceUrl string) (*Service, error) {
	for name, services := range s.services {
		for _, service := range services {
			if service.url == serviceUrl {
				parsed := toService(service, name)
				return &parsed, nil
			}
		}
	}
	return &Service{}, fmt.Errorf("[err] service with url %s doesnt exists", serviceUrl)
}

func (s *multiMapStorage) GetAllServices() (result []Service, err error) {
	for name, services := range s.services {
		size := len(services)
		if size > 0 {
			result = append(result, toService(services[rand.Intn(size)], name))
		}
	}
	if len(result) == 0 {
		err = errors.New("[err] storage is empty")
	}
	return result, err
}

func (s *multiMapStorage) UpdateLastHeartBeat(service Service, newTime time.Time) error {
	for idx, savedService := range s.services[service.Name] {
		if savedService.url == service.Url {
			log.Printf("Updating time on %s with url %s matching %s at index %d\n", service.Name, savedService.url, service.Url, idx)
			s.services[service.Name][idx].lastHeartBeatCheck = newTime
			return nil
		}
	}
	return fmt.Errorf("[err] service %s doesnt exists", service.Name)
}

func (s *multiMapStorage) deleteIfOlderThan(service Service, duration time.Duration) {
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

//TODO fix deletion after checked for unhealthy in multiValueMap and Slice implementation.
func (s *multiMapStorage) performDeletion(
	quit chan bool,
	service Service,
	duration time.Duration,
	tick time.Time) {
	saved, err := s.GetById(service.id)
	if err == nil && saved.LastHeartBeatCheck.Add(duration).Before(tick) {
		err := s.Remove(service.Name, service.id)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Deleted unhealthy service %s\n", service.id)
			quit <- true
		}
	}
}
func (s *multiMapStorage) checkIfDuplicate(service Service) bool {
	for _, savedService := range s.services[service.Name] {
		if savedService.url == service.Url {
			return true
		}
	}
	return false
}
func NewMultiMapStorage() Storage {
	return &multiMapStorage{
		services: make(map[string][]serviceMimified),
	}
}
