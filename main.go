package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/AmiraliQdi/golang-ghostops/internal/analysis"
	"github.com/AmiraliQdi/golang-ghostops/internal/parser"
	"github.com/AmiraliQdi/golang-ghostops/internal/report"
)

func main() {
	output := flag.String("o", "", "also write the report as JSON to `file`")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: ghostops [-o file] <logfile>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(flag.Arg(0), *output); err != nil {
		fmt.Fprintln(os.Stderr, "ghostops:", err)
		os.Exit(1)
	}
}

func run(inputPath, outputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer file.Close()

	analyzer, err := logReader(file)
	if err != nil {
		return err
	}

	result := report.New(analyzer)
	if outputPath != "" {
		if err := writeReportFile(outputPath, report.JSONReporter{}, result); err != nil {
			return err
		}
	}
	return report.TextReporter{}.Write(os.Stdout, result)
}

func writeReportFile(path string, reporter report.Reporter, result report.Report) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}

	err = reporter.Write(file, result)
	if closeErr := file.Close(); err == nil && closeErr != nil {
		err = fmt.Errorf("closing output file: %w", closeErr)
	}
	if err != nil {
		os.Remove(path)
		return err
	}
	return nil
}

func logReader(r io.Reader) (*analysis.Analyzer, error) {
	analyzer := analysis.NewAnalyzer()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		entry, err := parser.New(scanner.Text())
		if err != nil {
			analyzer.AddMalformed()
			continue
		}
		analyzer.AddEntry(entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading log file: %w", err)
	}
	return analyzer, nil
}
