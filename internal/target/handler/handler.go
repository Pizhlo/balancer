package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Pizhlo/balancer/internal/target/service"
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

func (h *Handler) GetCounter(w http.ResponseWriter, r *http.Request) {

	log.Println("got request: ", r.URL)

	w.WriteHeader(http.StatusOK)

	count := h.srv.GetCount()

	countJSON, err := json.Marshal(count)
	if err != nil {
		http.Error(w, fmt.Sprintf("error while making json counter: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write(countJSON)
}
