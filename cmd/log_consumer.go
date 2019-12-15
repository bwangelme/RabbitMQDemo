package main

import (
	"log"
	"os"
	"strings"

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
		log.Println("Usage: log_consumer [debug] [info] [warning] [error] [fatal]")
		os.Exit(0)
	}

	validLevels := []string{}
	for _, s := range os.Args[1:] {
		levels := "debug,info,warning,error,fatal"
		idx := strings.Index(levels, s)
		if idx == -1 {
			log.Printf("Invalid Level %s\n", s)
			continue
		}
		validLevels = append(validLevels, s)
	}

	if len(validLevels) == 0 {
		os.Exit(1)
	}

	for _, level := range validLevels {
		err = ch.QueueBind(
			q.Name,        // queue name
			level,         // routing key
			"logs_direct", // exchange name
			false,         // no-wait
			nil,           // args
		)
		utils.FailOnError(err, "Queue Bind")
		log.Printf("Bind to the %s\n", level)
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
