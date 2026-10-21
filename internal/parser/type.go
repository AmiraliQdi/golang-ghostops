package parser

import (
	"errors"
	"fmt"
)

type LogType int

const (
	Unknown LogType = iota
	Info
	Warn
	Error
	Fatal
)

var ErrInvalidLogType = errors.New("invalid log type")

var logTypesLookup = map[string]LogType{
	"INFO":  Info,
	"WARN":  Warn,
	"ERROR": Error,
	"FATAL": Fatal,
}

func (logType LogType) String() string {
	switch logType {
	case Unknown:
		return "UNKNOWN"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}
func ToLogType(rawLogType string) (LogType, error) {
	parsedLogType := logTypesLookup[rawLogType]
	if parsedLogType == Unknown {
		return Unknown, fmt.Errorf("%w %s", ErrInvalidLogType, rawLogType)
	}
	return parsedLogType, nil
}
