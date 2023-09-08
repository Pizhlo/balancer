package handler

import (
	"log"
	"net/http"

	"github.com/Pizhlo/balancer/internal/balancer/service"
)

type Handler struct {
	srv *service.Service
}

func New(srv *service.Service) *Handler {
	return &Handler{srv}
}

func (h *Handler) GetRequest(w http.ResponseWriter, r *http.Request) {

	log.Println("got request: ", r.URL)

	w.WriteHeader(http.StatusOK)

	h.srv.Handle()
}
