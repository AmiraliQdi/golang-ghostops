package report

import (
	"encoding/json"
	"fmt"
	"io"
)

func (JSONReporter) Write(w io.Writer, r Report) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r); err != nil {
		return fmt.Errorf("writing JSON report: %w", err)
	}
	return nil
}
