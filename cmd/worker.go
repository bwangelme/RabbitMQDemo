package main

import (
	"bytes"
	"log"
	"time"

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

	// 因为在启动消费者的时候不确定 hello 队列是否已经存在了，所以我们在这里事先声明了一次
	q, err := ch.QueueDeclare(
		"task_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	utils.FailOnError(err, "Queue Declare")

	forever := make(chan bool)

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	utils.FailOnError(err, "Set Qos")

	msgs, err := ch.Consume(
		q.Name, // name
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	utils.FailOnError(err, "Comsume")

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			dot_count := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dot_count)
			time.Sleep(t * time.Second)
			log.Printf("Done")
			d.Ack(false)
		}
	}()

	log.Printf("[*] Waiting for message. To exit press CTRL-C")
	<-forever
}
