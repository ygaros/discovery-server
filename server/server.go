package server

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
	"time"
	"ygaros-discovery-server/discovered"
)

type ParsedServices struct {
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	LastHeartBeat time.Time `json:"lastHeartBeat"`
}
type serviceDTO struct {
	Name   string `json:"name"`
	Port   int    `json:"port"`
	Domain string `json:"domain"`
}

type Server interface {
	AddService(w http.ResponseWriter, r *http.Request)
	ListServices(w http.ResponseWriter, r *http.Request)
	HeartBeat(w http.ResponseWriter, r *http.Request)
	GetService(w http.ResponseWriter, r *http.Request)
	Serve(port int) error
}
type server struct {
	storage discovered.ServiceStorage
}

func (s *server) AddService(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	service := serviceDTO{}

	if err := json.Unmarshal(body, &service); err != nil {
		log.Println("Failed to unmarshal payload:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Registered %s on %s:%d\n",
		service.Name,
		service.Domain,
		service.Port,
	)
	createdService := discovered.NewService(
		service.Name,
		service.Domain,
		service.Port,
	)
	err = s.storage.Add(createdService)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func (s *server) ListServices(w http.ResponseWriter, _ *http.Request) {

	if services, err := s.storage.GetAllServices(); err == nil {

		var serviceNames []ParsedServices
		for _, service := range services {
			serviceNames = append(serviceNames, ParsedServices{
				Name:          service.Name,
				Path:          fmt.Sprintf("%s:%d", service.Domain, service.Port),
				LastHeartBeat: service.LastHeartBeatCheck,
			})
		}
		if marshaled, err := json.Marshal(serviceNames); err == nil {
			w.Write(marshaled)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusBadRequest)
}
func (s *server) HeartBeat(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	service := serviceDTO{}

	if err := json.Unmarshal(body, &service); err != nil {
		log.Println("Failed to unmarshal payload:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	savedService, err := s.storage.Get(service.Name)
	if err != nil {
		log.Println("Service not registered yet", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	log.Printf("HeartBeat %s on %s:%d\n",
		service.Name,
		service.Domain,
		service.Port,
	)
	err = s.storage.UpdateLastHeartBeat(savedService.Name, time.Now())
	if err != nil {
		log.Println("Error occurred during updating new date", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
}
func (s *server) GetService(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("serviceName")
	if len(serviceName) == 0 {
		log.Println("serviceName parameter is mandatory!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	get, err := s.storage.Get(serviceName)
	if err != nil {
		log.Println(fmt.Sprintf("Service %s isnt registered!", serviceName))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if marshaled, err := json.Marshal(get); err == nil {
		w.Write(marshaled)
		return
	} else {
		log.Println(fmt.Sprintf("error occurred during processing getService request on %s", serviceName))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}
func (s *server) Serve(port int) error {
	if port == 0 {
		port = 7654
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Group(func(r chi.Router) {
		r.Post("/register", s.AddService)
		r.Post("/heartbeat", s.HeartBeat)
		r.Get("/list", s.ListServices)
		r.Get("/service", s.GetService)
	})
	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
func NewDiscoveryServer(storage discovered.ServiceStorage) Server {
	return &server{storage: storage}
}
func NewDiscoveryServerInMemoryStorage() Server {
	storage := discovered.NewInMemoryServiceStorage()
	return &server{storage: storage}
}
