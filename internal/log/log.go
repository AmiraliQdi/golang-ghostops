package log

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var InvalidLogFormat = errors.New("invalid log format")

type Log struct {
	timestamp time.Time
	logType   LogType
	source    string
	message   string
}

func New(rawLog string) (*Log, error) {
	splitLog := strings.SplitN(rawLog, " ", 4)
	if len(splitLog) != 4 {
		return nil, fmt.Errorf("%w: %s", InvalidLogFormat, rawLog)
	}
	timestamp, err := parseTimestamp(splitLog[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", InvalidLogFormat, err)
	}
	logType, err := ToLogType(splitLog[1])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", InvalidLogFormat, err)
	}
	source := splitLog[2]
	message := splitLog[3]
	return &Log{timestamp: timestamp, logType: logType, source: source, message: message}, nil
}

func parseTimestamp(timestamp string) (time.Time, error) {
	t, e := time.Parse(time.RFC3339, timestamp)
	if e != nil {
		return nil, fmt.Errorf("invalid log timestamp: %w", InvalidLogFormat)
	}
	return t, nil
}
