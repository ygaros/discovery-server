package discovered

import (
	"github.com/google/uuid"
	"time"
)

type Service struct {
	id                 uuid.UUID
	Name               string
	Port               int
	Domain             string
	LastHeartBeatCheck time.Time
}

func NewService(name string, domain string, port int) Service {
	return Service{
		id:                 uuid.New(),
		Name:               name,
		Domain:             domain,
		Port:               port,
		LastHeartBeatCheck: time.Now(),
	}
}
