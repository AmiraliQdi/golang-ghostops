package analysis

import (
	"github.com/AmiraliQdi/golang-ghostops/internal/incidents"
	"github.com/AmiraliQdi/golang-ghostops/internal/parser"
)

type Analyzer struct {
	Statistics     Statistics
	IncidentGroups map[string]*incidents.IncidentGroup
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		Statistics:     NewStatistics(),
		IncidentGroups: incidents.NewIncidentGroups(),
	}
}

func (a *Analyzer) AddEntry(entry *parser.Log) {
	a.Statistics.AddValidEntry()
	a.Statistics.AddSeverityLevelEntryCount(entry.LogType)
	a.Statistics.AddServiceEntryCount(entry.Service)
	incidents.AddLogToIncidentGroups(entry, a.IncidentGroups)
}

func (a *Analyzer) AddMalformed() {
	a.Statistics.AddInvalidEntry()
}
