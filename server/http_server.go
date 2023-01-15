package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ygaros/discovery-server/dto"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type HttpServer interface {
	AddService(w http.ResponseWriter, r *http.Request)
	ListServices(w http.ResponseWriter, r *http.Request)
	HeartBeat(w http.ResponseWriter, r *http.Request)
	GetService(w http.ResponseWriter, r *http.Request)
	Serve(port int) error
}
type httpServer struct {
	dservice DiscoveryService
}

func (s *httpServer) AddService(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	service := dto.Service{}

	if err := json.Unmarshal(body, &service); err != nil {
		log.Println("Failed to unmarshal payload:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Registered %s on %s\n",
		service.Name,
		service.Url,
	)
	err = s.dservice.AddService(service)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func (s *httpServer) ListServices(w http.ResponseWriter, _ *http.Request) {

	if services, err := s.dservice.ListServices(); err == nil {
		if marshaled, err := json.Marshal(services); err == nil {
			w.Write(marshaled)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusBadRequest)
}
func (s *httpServer) HeartBeat(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	service := dto.Service{}

	if err := json.Unmarshal(body, &service); err != nil {
		log.Println("Failed to unmarshal payload:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("HeartBeat %s on %s\n",
		service.Name,
		service.Url,
	)
	err = s.dservice.HeartBeat(service)
	if err != nil {
		log.Println("Error occurred during updating new date", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
}
func (s *httpServer) GetService(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("serviceName")
	if len(serviceName) == 0 {
		log.Println("serviceName parameter is mandatory!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	get, err := s.dservice.GetService(serviceName)
	if err != nil {
		log.Printf("Service %s isnt registered!\n", serviceName)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if marshaled, err := json.Marshal(get); err == nil {
		w.Write(marshaled)
		return
	} else {
		log.Printf("error occurred during processing getService request on %s\n", serviceName)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (s *httpServer) Serve(port int) error {
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

func NewHttpDiscoveryServer(discoveryService *DiscoveryService) HttpServer {
	return &httpServer{dservice: *discoveryService}
}

func NewHttpDiscoveryServerInMemoryStorage() HttpServer {
	return &httpServer{dservice: NewDiscoveryServiceWithInMemoryStorage()}
}

// func IndexMapping(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, "index.html")
// }
