package main

import (
	"log"

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

	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // args
	)
	utils.FailOnError(err, "ExchangeDeclare")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // args
	)
	utils.FailOnError(err, "Queue Declare")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"logs", // exchange name
		false,  // no-wait
		nil,    // args
	)

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
	utils.FailOnError(err, "Comsume")

	go func() {
		for d := range msgs {
			log.Printf("[x] %s", d.Body)
		}
	}()

	log.Printf("[*] Waiting for message. To exit press CTRL-C")
	<-forever
}
