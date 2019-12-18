package log

import (
	"fmt"
	"strings"
)

type Record struct {
	Message string
	Level   Level
	Module  string
}

func (r Record) String() string {
	return fmt.Sprintf("[%s]: %s", r.RoutingKey(), r.Message)
}

// Exchange Routing Key 的格式为 {module}.{level}
func (r Record) RoutingKey() string {
	return fmt.Sprintf("%s.%s", r.Module, strings.ToLower(fmt.Sprint(r.Level)))
}

func NewRecord(msg string, level Level, module string) *Record {
	return &Record{
		Message: msg,
		Level:   level,
		Module:  module,
	}
}

type Level int

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	FATAL
)

func NewLevelFromString(level string) Level {
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

func (l Level) String() string {
	nameMap := map[Level]string{
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
