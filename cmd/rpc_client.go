package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwangelme/RabbitMQDemo/utils"
	"github.com/streadway/amqp"
)

func ParseArgs(args []string) int {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "30"
	} else {
		s = strings.Join(args[1:], " ")
	}
	n, err := strconv.Atoi(s)
	utils.FailOnError(err, "Convert to Int")
	return n
}

func fibonacciRPC(n int) (res int) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	utils.FailOnError(err, "Dial")
	defer conn.Close()

	// Channel 中包含了大部分我们完成工作所需的API
	ch, err := conn.Channel()
	utils.FailOnError(err, "Channel")
	defer ch.Close()

	replyQueue, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	utils.FailOnError(err, "Queue Declare")

	msgs, err := ch.Consume(
		replyQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	utils.FailOnError(err, "Consume")

	corrID := utils.RandomString(32)

	err = ch.Publish(
		"",
		"rpc_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrID,
			// ReplyTo 指定服务端发送返回消息的队列
			ReplyTo: replyQueue.Name,
			Body:    []byte(strconv.Itoa(n)),
		},
	)

	for d := range msgs {
		if corrID == d.CorrelationId {
			res, err = strconv.Atoi(string(d.Body))
			utils.FailOnError(err, "conv")
			return res
		}
	}
	return 0
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	n := ParseArgs(os.Args)
	log.Printf(" [x] Requesting fib(%d)\n", n)
	res := fibonacciRPC(n)
	log.Printf("[x] fib(%d)=%d", n, res)
}
