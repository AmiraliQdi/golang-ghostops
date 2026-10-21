package report

import (
	"cmp"
	"io"
	"maps"
	"slices"
	"time"

	"github.com/AmiraliQdi/golang-ghostops/internal/analysis"
	"github.com/AmiraliQdi/golang-ghostops/internal/incidents"
	"github.com/AmiraliQdi/golang-ghostops/internal/parser"
)

type Reporter interface {
	Write(w io.Writer, r Report) error
}

type TextReporter struct{}

type JSONReporter struct{}

type Report struct {
	Summary           Summary    `json:"summary"`
	EntriesBySeverity []Count    `json:"entries_by_severity"`
	EntriesByService  []Count    `json:"entries_by_service"`
	Incidents         []Incident `json:"incidents"`
}

type Summary struct {
	TotalLines     int `json:"total_lines"`
	ValidEntries   int `json:"valid_entries"`
	MalformedLines int `json:"malformed_lines"`
}

type Count struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Incident struct {
	Type            string    `json:"type"`
	Severity        string    `json:"severity"`
	Occurrences     int       `json:"occurrences"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	DurationSeconds int64     `json:"duration_seconds"`
	Services        []Count   `json:"services"`
}

func (i Incident) Duration() time.Duration {
	return i.LastSeen.Sub(i.FirstSeen)
}

func New(a *analysis.Analyzer) Report {
	stats := a.Statistics
	r := Report{
		Summary: Summary{
			TotalLines:     stats.ValidEntries + stats.InvalidEntries,
			ValidEntries:   stats.ValidEntries,
			MalformedLines: stats.InvalidEntries,
		},
		EntriesBySeverity: []Count{},
		EntriesByService:  []Count{},
		Incidents:         []Incident{},
	}

	for level := parser.Info; level <= parser.Fatal; level++ {
		r.EntriesBySeverity = append(r.EntriesBySeverity, Count{level.String(), stats.SeverityLevelEntryCount[level]})
	}
	for _, service := range slices.Sorted(maps.Keys(stats.ServiceEntryCount)) {
		r.EntriesByService = append(r.EntriesByService, Count{service, stats.ServiceEntryCount[service]})
	}

	for _, group := range detectedGroups(a.IncidentGroups) {
		incident := Incident{
			Type:        group.Incident.Name,
			Severity:    group.Incident.Level.String(),
			Occurrences: group.Count,
			FirstSeen:   group.FirstOccurrence,
			LastSeen:    group.LastOccurrence,
			Services:    servicesByCount(group.AffectedServices),
		}
		incident.DurationSeconds = int64(incident.Duration().Seconds())
		r.Incidents = append(r.Incidents, incident)
	}
	return r
}

func detectedGroups(groups map[string]*incidents.IncidentGroup) []*incidents.IncidentGroup {
	var detected []*incidents.IncidentGroup
	for _, group := range groups {
		if group.Count > 0 {
			detected = append(detected, group)
		}
	}
	slices.SortFunc(detected, func(a, b *incidents.IncidentGroup) int {
		return cmp.Or(
			cmp.Compare(b.Incident.Level, a.Incident.Level),
			cmp.Compare(b.Count, a.Count),
			cmp.Compare(a.Incident.Name, b.Incident.Name),
		)
	})
	return detected
}

func servicesByCount(services map[string]int) []Count {
	counts := []Count{}
	for name, count := range services {
		counts = append(counts, Count{name, count})
	}
	slices.SortFunc(counts, func(a, b Count) int {
		return cmp.Or(cmp.Compare(b.Count, a.Count), cmp.Compare(a.Name, b.Name))
	})
	return counts
}
