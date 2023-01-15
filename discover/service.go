package discover

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
	return Service{
		id:                 uuid.New(),
		Name:               name,
		Url:                PrepareUrl(url, secure),
		LastHeartBeatCheck: time.Now(),
	}
}
func PrepareUrl(url string, secure bool) string {
	if secure {
		return fmt.Sprintf("%s%s", HTTPS, url)
	} else {
		return fmt.Sprintf("%s%s", HTTP, url)
	}
}
