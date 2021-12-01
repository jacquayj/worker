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

func (ll LogLevel) String() string {
	switch ll {
	case Info:
		return "INFO"
	case Warning:
		return "WARNING"
	case Error:
		return "ERROR"
	case None:
		return "NONE"
	default:
		return "UNKNOWN"
	}
}

func logMsg(maxLevel, level LogLevel, msg string, vals ...interface{}) {
	if level > maxLevel {
		return
	}
	msg = fmt.Sprintf(msg, vals...)
	log.Printf("%v: %s", level, msg)
}
