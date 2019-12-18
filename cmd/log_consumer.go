package main

import (
	"log"
	"os"

	"github.com/bwangelme/RabbitMQDemo/utils"
	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672/")
	utils.FailOnError(err, "Dial")
	defer conn.Close()

	ch, err := conn.Channel()
	utils.FailOnError(err, "Channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // args
	)
	utils.FailOnError(err, "Queue Declare")

	if len(os.Args) < 2 {
		log.Println("Usage: log_consumer `bind key`")
		os.Exit(0)
	}

	for _, bindKey := range os.Args[1:] {
		err = ch.QueueBind(
			q.Name,       // queue name
			bindKey,      // routing key
			"logs_topic", // exchange name
			false,        // no-wait
			nil,          // args
		)
		utils.FailOnError(err, "Queue Bind")
		log.Printf("Bind to the %s\n", bindKey)
	}

	forever := make(chan bool)

	msgs, err := ch.Consume(
		q.Name, // name
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	utils.FailOnError(err, "Consume")

	go func() {
		for d := range msgs {
			log.Printf("[x] %s:%s", d.RoutingKey, d.Body)
		}
	}()

	log.Printf("[*] Waiting for message. To exit press CTRL-C")
	<-forever
}
