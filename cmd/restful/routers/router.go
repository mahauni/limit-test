package routes

import (
	"net/http"

	"github.com/mahauni/limit-test/cmd/restful/handlers/data"
	"github.com/mahauni/limit-test/cmd/restful/handlers/videostream"
)

func LoadRoutes(router *http.ServeMux) {
	dataHandler := &data.Handler{}
	videoStreamHandler := &videostream.Handler{}

	router.HandleFunc("POST /post", dataHandler.Create)

	router.HandleFunc("/stream", videoStreamHandler.HtmlServe)
}
