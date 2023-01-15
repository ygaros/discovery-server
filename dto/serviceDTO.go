package dto

import (
	"time"

	proto "github.com/ygaros/discovery-server/gen/proto"
)

type Service struct {
	Name   string `json:"name"`
	Url    string `json:"url"`
	Secure bool   `json:"secure"`
}
type ServiceHeartBeat struct {
	Name          string    `json:"name"`
	Url           string    `json:"url"`
	LastHeartBeat time.Time `json:"lastHeartBeat"`
}

func ToService(service *proto.Service) Service {
	return Service{
		Name:   service.Name,
		Url:    service.Url,
		Secure: service.Secure,
	}
}
