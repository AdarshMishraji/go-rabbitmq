package main

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// here routing keys are simple one word, unlike topic exchange, we cannot subscribe to subtopics
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

	queue, err := channel.QueueDeclare(
		"", false, false, true, false, nil,
	)
	failOnError(err, "Failed to create queue")

	if len(os.Args) < 2 {
		log.Printf("Usage: %s [info] [warning] [error]", os.Args[0])
		os.Exit(0)
	}
	for _, s := range os.Args[1:] {
		log.Printf("Binding queue %s to exchange %s with routing key %s", queue.Name, "logs-direct", s)
		err = channel.QueueBind(
			queue.Name,
			s,
			"logs-direct",
			false, nil,
		)
		failOnError(err, "Failed to bind queue")
	}

	msgs, err := channel.Consume(
		queue.Name,
		"",
		true, false, false, false, nil,
	)
	failOnError(err, "Failed to start consuming messages")

	var forever chan struct{}

	go func() {
		for msg := range msgs {
			log.Printf("[x] %s", msg.Body)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
