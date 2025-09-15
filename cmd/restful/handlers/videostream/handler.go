package videostream

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
)

type Handler struct{}

func (h *Handler) HtmlServe(w http.ResponseWriter, r *http.Request) {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	t, _ := template.ParseFiles(pwd + "/cmd/restful/handlers/videostream/templates/index.html")
	t.Execute(w, nil)
}
