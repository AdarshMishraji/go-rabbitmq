package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// consumer can only consume messages from queue
// as the publisher is not sending the messages to the queue rather sending it to the exchange, so we create queue and bind that queue to that exchange
func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to AMQP")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "Failed to create channel")

	err = channel.ExchangeDeclare(
		"logs",
		"fanout",
		true,
		false, false, false, nil,
	)
	failOnError(err, "Failed to declare exchange")

	queue, err := channel.QueueDeclare(
		"",
		false, false, true, false, nil,
	)
	failOnError(err, "Failed to create temporary queue")

	err = channel.QueueBind(queue.Name, "", "logs", false, nil)
	failOnError(err, "Failed to bind the queue with the exchange")

	msgs, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
	failOnError(err, "Failed to start consuming messages")

	var forever chan struct{}

	go func() {
		for msg := range msgs {
			log.Printf(" [x] %s", msg.Body)
		}
	}()
	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")

	<-forever
}
