package main

import (
	"fmt"
	"net/http"
)

func main() {
	url := "http://localhost:1234/post"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(fmt.Sprintf("Error creating request: %v", err))
	}

	for range 10 {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(fmt.Sprintf("Error posting: %v", err))
		}
		defer resp.Body.Close()
	}
}
