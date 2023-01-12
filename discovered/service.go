package discovered

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	HTTPS = "https://"
	HTTP  = "http://"
)

type Service struct {
	id                 uuid.UUID
	Name               string    `json:"name"`
	Url                string    `json:"url"`
	LastHeartBeatCheck time.Time `json:"lastHeartBeatCheck"`
}

func NewService(name string, url string, secure bool) Service {
	var parsedUrl string
	if secure {
		parsedUrl = fmt.Sprintf("%s%s", HTTPS, url)
	} else {
		parsedUrl = fmt.Sprintf("%s%s", HTTP, url)
	}
	return Service{
		id:                 uuid.New(),
		Name:               name,
		Url:                parsedUrl,
		LastHeartBeatCheck: time.Now(),
	}
}
