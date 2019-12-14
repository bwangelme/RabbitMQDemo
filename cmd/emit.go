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
	if (len(args) < 2) || os.Args[1] == "" {
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

	// 声明 Exchange
	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)

	body := bodyForm(os.Args)
	err = ch.Publish(
		"logs", // exchange
		"",     // routing key
		false,  // mandatory 强制的
		false,  // immediate 即时的
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	log.Printf("[x] sent %s", body)
	utils.FailOnError(err, "Publish")
}
