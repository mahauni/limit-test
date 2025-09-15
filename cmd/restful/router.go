package main

import (
	"net/http"

	"github.com/mahauni/limit-test/cmd/restful/handlers/data"
	"github.com/mahauni/limit-test/cmd/restful/handlers/videostream"
)

func loadRoutes(router *http.ServeMux) {
	dataHandler := &data.Handler{}
	videoStreamHandler := &videostream.Handler{}

	router.HandleFunc("POST /post", dataHandler.Create)

	router.HandleFunc("/stream", videoStreamHandler.HtmlServe)
}
