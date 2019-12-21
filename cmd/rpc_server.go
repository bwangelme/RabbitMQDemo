package main

import (
	"log"
	"strconv"

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
		"rpc_queue", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // args
	)
	utils.FailOnError(err, "Queue Declare")

	// 这里声明每次只接受一个消息进行处理
	err = ch.Qos(
		1,
		0,
		false,
	)

	msgs, err := ch.Consume(
		q.Name,
		"",
		false, // auto ack
		false,
		false,
		false,
		nil,
	)
	utils.FailOnError(err, "Consume")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			n, err := strconv.Atoi(string(d.Body))
			if err != nil {
				log.Printf("Convert '%s' failed: %v", string(d.Body), err)
				sendResponse(0, ch, d.ReplyTo, d.CorrelationId)
				d.Ack(false)
				continue
			}

			log.Printf(" [.] fib(%d)", n)
			n = utils.Fib(n)
			sendResponse(n, ch, d.ReplyTo, d.CorrelationId)
			d.Ack(false)
		}
	}()

	log.Printf("[*] Waiting for message. To exit press CTRL-C")
	<-forever
}

func sendResponse(n int, ch *amqp.Channel, replyTo string, correlationID string) {
	err := ch.Publish(
		"",
		replyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: correlationID,
			Body:          []byte(strconv.Itoa(n)),
		},
	)
	utils.FailOnError(err, "Publish")
}
