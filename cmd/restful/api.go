package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mahauni/limit-test/cmd/restful/middlewares"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s\n", msg, err)
	}
}

type Data struct {
	ID string `json:"id"`
}

var address = ":1234"

func main() {
	conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to connect to a channel")
	defer ch.Close()

	router := http.NewServeMux()
	router.Handle("/file/", http.StripPrefix("/file/", http.FileServer(http.Dir("./media"))))
	loadRoutes(router)

	logging := middlewares.NewRabbitmqLoggingChannel(ch)

	stack := middlewares.CreateStack(
		logging.Logging,
	)

	server := http.Server{
		Addr:    address,
		Handler: stack(router),
	}

	fmt.Printf("Server listening on %s\n", address)
	log.Fatal(server.ListenAndServe())
}
