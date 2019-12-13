package main

import (
	"log"
	"os"
	"strings"

	"github.com/bwangelme/RabbitMQDemo/utils"
	"github.com/streadway/amqp"
)

func bodyForm(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1]== "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}

	return s
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	utils.FailOnError(err, "Dial")
	defer conn.Close()

	// Channel 中包含了大部分我们完成工作所需的API
	ch, err := conn.Channel()
	utils.FailOnError(err, "Channel")
	defer ch.Close()

	// 声明 Queue 是非幂等的，即你只可以在它不存在的时候创建它
	q, err := ch.QueueDeclare(
		"hello", // Queue name
		true,   // durable  持久性
		false,   // delete when unused
		false,   // exclusive 独占
		false,   // no-wait
		nil,     // arguments
	)
	utils.FailOnError(err, "Queue Declare")

	body := bodyForm(os.Args)
	err = ch.Publish(
		"",     //exchange
		q.Name, // routing key
		false,  // mandatory 强制的
		false,  // immediate 即时的
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	log.Printf("[x] sent %s", body)
	utils.FailOnError(err, "Publish")
}
