package analysis

import "github.com/AmiraliQdi/golang-ghostops/internal/parser"

type Statistics struct {
	ValidEntries            int
	InvalidEntries          int
	SeverityLevelEntryCount map[parser.LogType]int
	ServiceEntryCount       map[string]int
}

func NewStatistics() Statistics {
	return Statistics{
		0, 0, make(map[parser.LogType]int), make(map[string]int),
	}
}

func (s *Statistics) AddValidEntry() {
	s.ValidEntries++
}

func (s *Statistics) AddInvalidEntry() {
	s.InvalidEntries++
}

func (s *Statistics) AddSeverityLevelEntryCount(logType parser.LogType) {
	s.SeverityLevelEntryCount[logType]++
}

func (s *Statistics) AddServiceEntryCount(serviceName string) {
	s.ServiceEntryCount[serviceName]++
}
