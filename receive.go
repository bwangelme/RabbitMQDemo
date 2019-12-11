package main

import (
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", err, msg)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672/")
	failOnError(err, "Dial")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Channel")
	defer ch.Close()

	// 因为在启动消费者的时候不确定 hello 队列是否已经存在了，所以我们在这里事先声明了一次
	q, err := ch.QueueDeclare(
		"hello",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Queue Declare")

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

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf("[*] Waiting for message. To exit press CTRL-C")
	<-forever
}
