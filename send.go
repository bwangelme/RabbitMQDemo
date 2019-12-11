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
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Dial")
	defer conn.Close()

	// Channel 中包含了大部分我们完成工作所需的API
	ch, err := conn.Channel()
	failOnError(err, "Channel")
	defer ch.Close()

	// 声明 Queue 是非幂等的，即你只可以在它不存在的时候创建它
	q, err := ch.QueueDeclare(
		"hello", // Queue name
		false,   // durable  持久性
		false,   // delete when unused
		false,   // exclusive 独占
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Queue Declare")

	body := "Hello, World"
	err = ch.Publish(
		"",     //exchange
		q.Name, // routing key
		false,  // mandatory 强制的
		false,  // immediate 即时的
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	log.Printf("[x] sent %s", body)
	failOnError(err, "Publish")
}
