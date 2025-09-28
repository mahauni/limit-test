package main

import (
	"log"

	"github.com/caarlos0/env/v11"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s\n", msg, err)
	}
}

type config struct {
	RabbitMQUrl string `env:"RABBITMQ_URL" envDefault:"amqp://user:password@localhost:5672/"`
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

	err = ch.ExchangeDeclare(
		"logs",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a exchange")

	q, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,
		"",
		"logs",
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s\n", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
