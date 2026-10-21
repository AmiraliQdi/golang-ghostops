package incidents

import (
	"strings"

	"github.com/AmiraliQdi/golang-ghostops/internal/parser"
)

type Incident struct {
	Name     string
	MinLevel parser.LogType
	keywords []string
}

var incidentRules = []Incident{
	{
		Name:     "Crashed",
		MinLevel: parser.Fatal,
		keywords: []string{"crashed"},
	},
	{
		Name:     "Database",
		MinLevel: parser.Error,
		keywords: []string{"database", "sql"},
	},
	{
		Name:     "Timeout",
		MinLevel: parser.Error,
		keywords: []string{"timeout", "timed out"},
	},
	{
		Name:     "BadGateway",
		MinLevel: parser.Error,
		keywords: []string{"bad gateway"},
	},
	{
		Name:     "UncaughtException",
		MinLevel: parser.Error,
		keywords: []string{"nullpointerexception"},
	},
	{
		Name:     "DatabasePoolPressure",
		MinLevel: parser.Warn,
		keywords: []string{"pool"},
	},
	{
		Name:     "Memory",
		MinLevel: parser.Warn,
		keywords: []string{"memory"},
	},
	{
		Name:     "RateLimit",
		MinLevel: parser.Warn,
		keywords: []string{"rate", "limit"},
	},
	{
		Name:     "API",
		MinLevel: parser.Warn,
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
	if entry.LogType < inc.MinLevel {
		return false
	}
	for _, keyword := range inc.keywords {
		if strings.Contains(strings.ToLower(entry.Message), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
