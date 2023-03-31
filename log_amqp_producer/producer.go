package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type LogMessage struct {
	Application string `json:"application"`
	Level       string `json:"level"`
	Message     string `json:"message"`
}

func main() {
	amqpConn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer amqpConn.Close()

	amqpChannel, err := amqpConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer amqpChannel.Close()

	queue, err := amqpChannel.QueueDeclare(
		"log_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	logMessage := LogMessage{
		Application: "demo producer",
		Level:       "info",
		Message:     "This is a test log message sent via AQMP.",
	}

	body, err := json.Marshal(logMessage)
	if err != nil {
		log.Fatalf("Failed to marshal log message: %v", err)
	}

	err = amqpChannel.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish a message: %v", err)
	}

	log.Println("Sent a log message")
}
