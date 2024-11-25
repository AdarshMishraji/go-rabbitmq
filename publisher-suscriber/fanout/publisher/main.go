package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func bodyFrom(args []string) string {
	var s string
	if len(args) < 2 || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "Failed to create channel")
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"logs",
		"fanout",
		true,
		false, false, false, nil,
	)
	failOnError(err, "Failed to create exchange")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := bodyFrom(os.Args)

	err = channel.PublishWithContext(ctx, "logs", "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	})
	log.Printf("[x] Sent %s", body)
}
