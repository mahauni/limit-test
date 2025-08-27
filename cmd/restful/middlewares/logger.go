package middlewares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqChannel struct {
	*amqp.Channel
}

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s\n", msg, err)
	}
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func NewRabbitmqLoggingChannel(ch *amqp.Channel) RabbitmqChannel {
	return RabbitmqChannel{
		ch,
	}
}

func (ch *RabbitmqChannel) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := ch.ExchangeDeclare(
			"logs",
			"fanout",
			true,
			false,
			false,
			false,
			nil,
		)
		failOnError(err, "Failed to declare an exchange")

		body := fmt.Sprintf("%d %s %v", wrapped.statusCode, r.URL.Path, time.Since(startTime))
		err = ch.PublishWithContext(
			ctx,
			"logs",
			"",
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			},
		)
		failOnError(err, "Failed to publish a message")
	})
}
