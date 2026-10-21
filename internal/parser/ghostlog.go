package parser

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidLogFormat = errors.New("invalid log format")
var ErrInvalidLogTimestamp = errors.New("invalid log timestamp")

type Log struct {
	Timestamp time.Time
	LogType   LogType
	Service   string
	Message   string
}

func New(rawLog string) (*Log, error) {
	splitLog := strings.Fields(rawLog)
	if len(splitLog) < 4 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidLogFormat, rawLog)
	}

	timestamp, err := parseTimestamp(splitLog[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidLogFormat, err)
	}

	logType, err := ToLogType(splitLog[1])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidLogFormat, err)
	}

	source := splitLog[2]
	message := strings.Join(splitLog[3:], " ")
	return &Log{Timestamp: timestamp, LogType: logType, Service: source, Message: message}, nil
}

func parseTimestamp(timestamp string) (time.Time, error) {
	t, e := time.Parse(time.RFC3339, timestamp)
	if e != nil {
		return t, fmt.Errorf("%w: %w", ErrInvalidLogTimestamp, e)
	}
	return t, nil
}
