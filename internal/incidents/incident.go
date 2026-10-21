package incidents

import (
	"strings"

	"github.com/AmiraliQdi/golang-ghostops/internal/parser"
)

type Incident struct {
	Name     string
	Level    parser.LogType
	keywords []string
}

var incidentRules = []Incident{
	{
		Name:     "Crashed",
		Level:    parser.Fatal,
		keywords: []string{"crashed"},
	},
	{
		Name:     "Database",
		Level:    parser.Error,
		keywords: []string{"database", "sql"},
	},
	{
		Name:     "Timeout",
		Level:    parser.Error,
		keywords: []string{"timeout", "timed out"},
	},
	{
		Name:     "BadGateway",
		Level:    parser.Error,
		keywords: []string{"bad gateway"},
	},
	{
		Name:     "UncaughtException",
		Level:    parser.Error,
		keywords: []string{"nullpointerexception"},
	},
	{
		Name:     "DatabasePoolPressure",
		Level:    parser.Warn,
		keywords: []string{"pool"},
	},
	{
		Name:     "Memory",
		Level:    parser.Warn,
		keywords: []string{"memory"},
	},
	{
		Name:     "RateLimit",
		Level:    parser.Warn,
		keywords: []string{"rate", "limit"},
	},
	{
		Name:     "API",
		Level:    parser.Warn,
		keywords: []string{"deprecated", "api"},
	},
}

func getLogIncident(entry *parser.Log) *Incident {
	for i := range incidentRules {
		if incidentRules[i].contains(entry) {
			return &incidentRules[i]
		}
	}
	return nil
}

func (inc *Incident) contains(entry *parser.Log) bool {
	if entry.LogType != inc.Level {
		return false
	}
	for _, keyword := range inc.keywords {
		if strings.Contains(strings.ToLower(entry.Message), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
