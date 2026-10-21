package report

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

const timeLayout = time.RFC3339

func (TextReporter) Write(w io.Writer, r Report) error {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	s := r.Summary

	fmt.Fprintln(tw, "GhostOps Report")
	fmt.Fprintln(tw, "===============")
	fmt.Fprintln(tw)

	section(tw, "Summary")
	fmt.Fprintf(tw, "  Total lines\t%d\n", s.TotalLines)
	fmt.Fprintf(tw, "  Valid entries\t%d\t%s\n", s.ValidEntries, percent(s.ValidEntries, s.TotalLines))
	fmt.Fprintf(tw, "  Malformed lines\t%d\t%s\n", s.MalformedLines, percent(s.MalformedLines, s.TotalLines))
	fmt.Fprintln(tw)

	section(tw, "Entries by severity")
	for _, c := range r.EntriesBySeverity {
		fmt.Fprintf(tw, "  %s\t%d\t%s\n", c.Name, c.Count, percent(c.Count, s.ValidEntries))
	}
	fmt.Fprintln(tw)

	section(tw, "Entries by service")
	if len(r.EntriesByService) == 0 {
		fmt.Fprintln(tw, "  (none)")
	}
	for _, c := range r.EntriesByService {
		fmt.Fprintf(tw, "  %s\t%d\t%s\n", c.Name, c.Count, percent(c.Count, s.ValidEntries))
	}
	fmt.Fprintln(tw)

	section(tw, fmt.Sprintf("Incidents (%d detected)", len(r.Incidents)))
	if len(r.Incidents) == 0 {
		fmt.Fprintln(tw, "  No incidents detected.")
	}
	for i, incident := range r.Incidents {
		fmt.Fprintf(tw, "  %d. %s [%s]\n", i+1, incident.Type, incident.Severity)
		fmt.Fprintf(tw, "     Occurrences:\t%d\n", incident.Occurrences)
		fmt.Fprintf(tw, "     First seen:\t%s\n", incident.FirstSeen.Format(timeLayout))
		fmt.Fprintf(tw, "     Last seen:\t%s\n", incident.LastSeen.Format(timeLayout))
		fmt.Fprintf(tw, "     Duration:\t%s\n", incident.Duration())
		fmt.Fprintf(tw, "     Services:\t%s\n", formatServices(incident.Services))
		if i < len(r.Incidents)-1 {
			fmt.Fprintln(tw)
		}
	}

	if err := tw.Flush(); err != nil {
		return fmt.Errorf("rendering report: %w", err)
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}
	return nil
}

func section(w io.Writer, title string) {
	fmt.Fprintln(w, title)
	fmt.Fprintln(w, strings.Repeat("-", len(title)))
}

func percent(part, whole int) string {
	if whole == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", float64(part)*100/float64(whole))
}

func formatServices(services []Count) string {
	parts := make([]string, len(services))
	for i, s := range services {
		parts[i] = fmt.Sprintf("%s (%d)", s.Name, s.Count)
	}
	return strings.Join(parts, ", ")
}
