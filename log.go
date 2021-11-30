package worker

import (
	"fmt"
	"log"
)

type LogLevel int

const (
	None LogLevel = iota
	Error
	Warning
	Info
)

func logMsg(maxLevel, level LogLevel, msg string, vals ...interface{}) {
	if level > maxLevel {
		return
	}

	msg = fmt.Sprintf(msg, vals...)

	switch level {
	case Info:
		log.Printf("INFO: %s\n", msg)
	case Warning:
		log.Printf("WARNING: %s\n", msg)
	case Error:
		log.Printf("ERROR: %s\n", msg)
	default:
		log.Println(msg)
	}
}
