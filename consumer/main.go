package main

import (
	"files/consumer/processing"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"github.com/streadway/amqp"
)

const (
	AMPQ_URL = "amqp://guest:guest@localhost:5672/"
)

func main() {
	// Define RabbitMQ server URL.
	// amqpServerURL := os.Getenv("AMQP_SERVER_URL")

	// Create a new RabbitMQ connection.
	connectRabbitMQ, err := amqp.Dial(AMPQ_URL)
	if err != nil {
		panic(err)
	}
	defer connectRabbitMQ.Close()

	// Opening a channel to our RabbitMQ instance over
	// the connection we have already established.
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}
	defer channelRabbitMQ.Close()

	// Subscribing to QueueService1 for getting messages.
	files, err := channelRabbitMQ.Consume(
		"UploadedFiles", // queue name
		"",              // consumer
		true,            // auto-ack
		false,           // exclusive
		false,           // no local
		false,           // no wait
		nil,             // arguments
	)
	if err != nil {
		log.Println(err)
	}

	// Build a welcome message.
	log.Println("Successfully connected to RabbitMQ")
	log.Println("Waiting for messages")

	// Make a channel to receive messages into infinite loop.
	forever := make(chan bool)

	p := processing.Processor{L: log.Default()}

	go func() {
		for f := range files {
			// For example, show received message in a console.
			log.Printf(" > Received message: %s\n", f.Body)
			err = p.ProcessFile(f.Body)
			if err != nil {
				log.Print(err)
			}

		}
	}()

	//  TODO: add select statement so to handle graceful shutdown
	<-forever
}
