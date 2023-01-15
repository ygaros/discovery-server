package discover

import (
	"time"

	"github.com/google/uuid"
)

const DELETION_TIME = 90 * time.Second

type Storage interface {
	Add(service Service) error
	Remove(serviceName string, serviceId uuid.UUID) error
	Get(serviceName string) (*Service, error)
	GetById(serviceId uuid.UUID) (*Service, error)
	GetByUrl(serviceUrl string) (*Service, error)
	GetAllServices() ([]Service, error)
	UpdateLastHeartBeat(service Service, newTime time.Time) error
}
