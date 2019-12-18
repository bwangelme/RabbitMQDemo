package main

import (
	syslog "log"
	"os"
	"strings"

	"github.com/bwangelme/RabbitMQDemo/log"
	"github.com/bwangelme/RabbitMQDemo/utils"
	"github.com/streadway/amqp"
)

func ParseArgs(args []string) *log.Record {
	var msg = "hello"
	var module string = "net"
	var level = log.INFO
	if len(args) == 2 {
		msg = args[1]
	} else if len(args) == 3 {
		level = log.NewLevelFromString(args[1])
		msg = args[2]
	} else if len(args) >= 4 {
		module = args[1]
		level = log.NewLevelFromString(args[2])
		msg = strings.Join(args[3:], " ")
	}

	return log.NewRecord(msg, level, module)
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
		"logs_topic", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	logRecord := ParseArgs(os.Args)

	// Exchange Routing Key 的格式为 {module}.{level}
	err = ch.Publish(
		"logs_topic",           // exchange
		logRecord.RoutingKey(), // routing key
		false,                  // mandatory 强制的
		false,                  // immediate 即时的
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(logRecord.Message),
		},
	)
	syslog.Printf("[x] sent %s", logRecord)
	utils.FailOnError(err, "Publish")
}
