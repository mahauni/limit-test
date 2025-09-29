package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env/v11"
	"github.com/mahauni/limit-test/cmd/restful/middlewares"
	routers "github.com/mahauni/limit-test/cmd/restful/routers"
	"github.com/mahauni/limit-test/cmd/telemetry"
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

type config struct {
	Port        string `env:"PORT" envDefault:"1234"`
	RabbitMQUrl string `env:"RABBITMQ_URL" envDefault:"amqp://user:password@localhost:5672/"`
	OtelUrl     string `env:"OTEL_URL" envDefault:"localhost:4318"`
}

func main() {
	var cfg config
	err := env.Parse(&cfg)
	failOnError(err, "Failed to load environment variables")

	conn, err := amqp.Dial(cfg.RabbitMQUrl)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to connect to a channel")
	defer ch.Close()

	tel := telemetry.NewTelemetry(context.Background(), cfg.OtelUrl, "limit-test-service", "local")
	meter, err := tel.InitMeterProvider()
	failOnError(err, "Failed to init the meter provider")
	defer meter.Shutdown(tel.Context)
	tracer, err := tel.InitTracerProvider()
	failOnError(err, "Failed to init the tracer provider")
	defer tracer.Shutdown(tel.Context)

	router := http.NewServeMux()
	router.Handle("/file/", http.StripPrefix("/file/", http.FileServer(http.Dir("./media"))))
	routers.LoadRoutes(router)

	logging := middlewares.NewRabbitmqLoggingChannel(ch)
	telemetry := middlewares.NewOtelTelemtry(tracer, meter, "meter-test")

	stack := middlewares.CreateStack(
		logging.Logging,
		telemetry.Telemetry,
	)

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: stack(router),
	}

	fmt.Printf("Server listening on %s\n", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
