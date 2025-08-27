package main

import (
	"net/http"

	"github.com/mahauni/limit-test/cmd/restful/handlers/data"
)

func loadRoutes(router *http.ServeMux) {
	dataHandler := &data.Handler{}

	router.HandleFunc("POST /post", dataHandler.Create)
}
