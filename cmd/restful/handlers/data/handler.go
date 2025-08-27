package data

import (
	"log"
	"net/http"
)

type Handler struct{}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	log.Println("received request POST")
	w.Write([]byte("Datacreated!"))
}

func (h *Handler) FindByID(w http.ResponseWriter, r *http.Request) {
	log.Println("handling READ request - Method:", r.Method)
}

func (h *Handler) UpdateByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling UPDATE request - Method:", r.Method)
}

func (h *Handler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	log.Println("received DELETE request for data")
}

func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
}
