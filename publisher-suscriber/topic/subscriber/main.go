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

// here routing keys are nested (delimited by .) with * and # wildcard characters. Ex: "kernel.*". "syslog.#". "#.critical"
// * for next any word
// # for any number of next words
func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"logs-topic",
		"topic",
		true, false, false, false, nil,
	)
	failOnError(err, "Failed to declare exchange")

	queue, err := channel.QueueDeclare(
		"",
		false, false, true, false, nil,
	)
	failOnError(err, "Failed to declare queue")

	if len(os.Args) < 2 {
		log.Printf("Usage: %s [binding_key]...", os.Args[0])
		os.Exit(0)
	}

	for _, s := range os.Args[1:] {
		log.Printf("Binding queue %s to exchange %s with routing key %s", queue.Name, "logs-topic", s)
		err = channel.QueueBind(
			queue.Name,
			s,
			"logs-topic",
			false, nil,
		)
		failOnError(err, "Failed to bind a queue")
	}

	msgs, err := channel.Consume(
		queue.Name,
		"",
		true, false, false, false, nil,
	)

	var forever chan struct{}

	go func() {
		for msg := range msgs {
			log.Printf("[x] %s", msg.Body)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
