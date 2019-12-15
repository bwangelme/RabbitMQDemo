package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwangelme/RabbitMQDemo/utils"
	"github.com/streadway/amqp"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
)

func NewLevel(level string) LogLevel {
	level = strings.ToUpper(level)

	if level == "DEBUG" {
		return DEBUG
	} else if level == "INFO" {
		return INFO
	} else if level == "WARNING" {
		return WARNING
	} else if level == "ERROR" {
		return ERROR
	} else if level == "FATAL" {
		return FATAL
	} else {
		return INFO
	}
}

func (l LogLevel) String() string {
	nameMap := map[LogLevel]string{
		DEBUG:   "DEBUG",
		INFO:    "INFO",
		WARNING: "WARNING",
		ERROR:   "ERROR",
		FATAL:   "FATAL",
	}
	name, ok := nameMap[l]
	if !ok {
		return ""
	} else {
		return name
	}
}

type Log struct {
	msg   string
	level LogLevel
}

func (l Log) String() string {
	return fmt.Sprintf("%s: %s", l.level, l.msg)
}

func ParseArgs(args []string) Log {
	var msg string
	var level LogLevel = INFO
	if (len(args) < 2) || os.Args[1] == "" {
		msg = "hello"
	} else if len(args) == 2 {
		msg = args[1]
	} else if len(args) >= 3 {
		level = NewLevel(args[1])
		msg = strings.Join(args[2:], " ")
	}

	return Log{
		msg:   msg,
		level: level,
	}
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
		"logs_direct", // name
		"direct",      // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)

	logRecord := ParseArgs(os.Args)
	strLevel := strings.ToLower(fmt.Sprint(logRecord.level))
	err = ch.Publish(
		"logs_direct", // exchange
		strLevel,      // routing key
		false,         // mandatory 强制的
		false,         // immediate 即时的
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(logRecord.msg),
		},
	)
	log.Printf("[x] sent %s", logRecord)
	utils.FailOnError(err, "Publish")
}
