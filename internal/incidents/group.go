package incidents

import (
	"time"

	"github.com/AmiraliQdi/golang-ghostops/internal/parser"
)

var levelWeights = map[parser.LogType]int{
	parser.Warn:  1,
	parser.Error: 3,
	parser.Fatal: 10,
}

type IncidentGroup struct {
	Incident         *Incident
	AffectedServices map[string]int
	LevelCounts      map[parser.LogType]int
	FirstOccurrence  time.Time
	LastOccurrence   time.Time
	Count            int
}

func NewIncidentGroups() map[string]*IncidentGroup {
	groups := make(map[string]*IncidentGroup)
	for i := range incidentRules {
		addNewIncidentGroup(groups, &incidentRules[i])
	}
	return groups
}

func addNewIncidentGroup(groups map[string]*IncidentGroup, incident *Incident) *IncidentGroup {
	groups[incident.Name] = &IncidentGroup{Incident: incident, AffectedServices: make(map[string]int),
		LevelCounts: make(map[parser.LogType]int), FirstOccurrence: time.Time{}, LastOccurrence: time.Time{}, Count: 0}
	return groups[incident.Name]
}

func AddLogToIncidentGroups(entry *parser.Log, incidentGroups map[string]*IncidentGroup) {
	logIncident := getLogIncident(entry)
	if logIncident == nil {
		return
	}

	group, ok := incidentGroups[logIncident.Name]
	if !ok {
		group = addNewIncidentGroup(incidentGroups, logIncident)
	}

	group.Count++
	if group.Count == 1 {
		group.LastOccurrence = entry.Timestamp
		group.FirstOccurrence = entry.Timestamp
	}

	group.AffectedServices[entry.Service]++
	group.LevelCounts[entry.LogType]++

	if entry.Timestamp.After(group.LastOccurrence) {
		group.LastOccurrence = entry.Timestamp
	}

	if entry.Timestamp.Before(group.FirstOccurrence) {
		group.FirstOccurrence = entry.Timestamp
	}
}

func (g *IncidentGroup) Severity() int {
	severity := 0
	for level, count := range g.LevelCounts {
		severity += count * levelWeights[level]
	}
	return severity
}
