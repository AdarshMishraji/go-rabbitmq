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
	if len(args) < 2 || os.Args[2] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[2:], " ")
	}

	return s
}

func severityFrom(args []string) string {
	if len(args) < 2 || os.Args[1] == "" {
		return "info"
	} else {
		return os.Args[1]
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	failOnError(err, "Failed to connect to AMQP")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "Failed to create channel")
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"logs-direct",
		"direct",
		true, false, false, false, nil,
	)
	failOnError(err, "Failed to create exchange")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := bodyFrom(os.Args)
	err = channel.PublishWithContext(
		ctx,
		"logs-direct",
		severityFrom(os.Args),
		false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent %s", body)

}
