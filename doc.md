# GhostOps: architecture and design decisions

## Pipeline

```
log file ──> main (read loop) ──> parser.New ──> analysis.Analyzer ──> report.New ──> Reporter
             bufio.Scanner        one line       statistics +          sorted         TextReporter -> stdout
                                  -> *Log        incident groups       Report         JSONReporter -> -o file
```

Each stage has one job and only knows about the stage directly before it:

| Package | Responsibility | Knows about |
|---|---|---|
| `main` | flags, opening and closing files, exit codes, the read loop | everything (it wires the pipeline) |
| `parser` | turn one line into a `Log` or an error | nothing else |
| `incidents` | the rule list, matching an entry to a rule, incident groups | `parser` |
| `analysis` | count entries and feed incident groups | `parser`, `incidents` |
| `report` | build the final `Report`, format it as text or JSON | `analysis` (only in `report.New`) |

## Streaming input

`logReader` takes an `io.Reader` and reads it with `bufio.Scanner`, one line at a time. A parsed line is counted and then discarded; nothing keeps a list of entries.

Memory use therefore depends on the number of distinct services and incident rules, not on the size of the file. Accepting an `io.Reader` instead of a file name also means the same code reads a file, standard input, or a `strings.Reader` in tests.


## Parsing and validation

`parser.New` splits a line with `strings.Fields`, which splits on any run of whitespace. The first three fields are the timestamp, level and service, and the rest are joined with single spaces to form the message. A line is rejected when:

- it has fewer than four fields (so a missing service or an empty message is malformed),
- the timestamp is not RFC3339 (`time.Parse(time.RFC3339, ...)`),
- the level is not `INFO`, `WARN`, `ERROR` or `FATAL`.

Errors wrap sentinel values (`ErrInvalidLogFormat`, `ErrInvalidLogTimestamp`, `ErrInvalidLogType`) with `%w`, so callers and tests can check the cause with `errors.Is`.

Splitting on a single space with `strings.SplitN` was the first approach, but a double space produced an empty field, so `... ERROR  svc msg` was accepted with an empty service name, and `... ERROR svc ` was accepted with an empty message. `strings.Fields` rules out both. The trade-off is that repeated spaces inside a message are collapsed to one, which does not affect keyword matching.

Levels are an `int` type (`LogType`) with a `String()` method. Comparing ints is cheap, and the fixed order `Info < Warn < Error < Fatal` is used to sort incidents by severity.

## Incident detection

A rule is a level plus a list of keywords. An entry matches when the level is equal **and** the lower-cased message contains one of the keywords.

Checking the level first matters. For example `INFO ... Token generated for user session` contains `rate` (in "gene**rate**d") and `INFO ... GET /api/v1/health` contains `api`, but neither is a WARN, so neither is reported as an incident.

### Overlapping rules

One entry can match more than one rule. In the sample log, `ERROR Failed to connect to database: timeout after 5000ms` (194 lines) matches both Database and Timeout.

The policy is: **rules are checked in list order and the first match wins.** Each entry belongs to at most one incident.

- **Deterministic.** The rules are stored in a slice, not a map, so the order is the same on every run.
- **One policy for every case.** There are no special cases for particular pairs of rules.
- **Counts are not inflated.** The sum of incident occurrences is never larger than the number of valid entries, and no entry is counted twice.

The order puts the more specific or root-cause rule first. A timeout while connecting to the database is reported as a database problem, so Database comes before Timeout.

### Bad gateway

The BadGateway rule (ERROR, keyword `bad gateway`) covers `Payment gateway returned 502 Bad Gateway`, which appears 141 times in the sample log and matched no rule before. It overlaps with no other rule. The keyword is the phrase `bad gateway` rather than `502`, because a bare number could also match unrelated text such as IDs or IP addresses.

### Grouping

There is one group per incident type. A group records:

- the incident (type and severity),
- the number of occurrences,
- the first and last timestamp,
- the number of occurrences per affected service.

Because entries are not in chronological order, every entry is compared against both ends of the time range. The first entry of a group sets both ends (`Count == 1`), so the zero `time.Time` never appears in a range.

Grouping by type rather than by (type, service) keeps the report short and still answers "which services are affected and how much" through the per-service counts.

## Reporting

### The `Report` model

`report.New(analyzer)` converts the analysis result into a `Report` struct once. All decisions about what to show and in which order are made there:

- severities in level order (INFO to FATAL), always all four, even with a count of zero,
- services sorted by name,
- incidents with zero occurrences dropped, the rest sorted by severity, then occurrences, then type,
- services inside an incident sorted by count, then name.

Go map iteration order is random, so every map is turned into a sorted slice here. This is what makes both outputs deterministic, and it means the text and JSON reports can never disagree. Slices are initialised as empty, not nil, so JSON shows `[]` rather than `null` for an empty input.

The `report` package depends on the analysis code only inside `report.New`. If the analyzer changes, the formatters do not.

### The `Reporter` interface

```go
type Reporter interface {
	Write(w io.Writer, r Report) error
}
```

`TextReporter` and `JSONReporter` implement it. This is the one interface defined in the project, because it is the one place with several interchangeable implementations of the same behaviour. `writeReportFile` in `main` takes a `Reporter` and handles creating, closing and cleaning up the file without knowing the format. Adding a new format means adding one type with a `Write` method.

The interface lives in `report` next to its implementations rather than in `main`, because `report` owns the set of available formats.

Writing to an `io.Writer` lets the same code print to the terminal, write a file, or fill a `bytes.Buffer` in tests. The text reporter builds the whole report in a buffer first and writes it in one call, so an output failure is reported as one error.

## Error handling

There are two kinds of problems, and they are handled differently:

| Kind | Examples | Handling |
|---|---|---|
| Invalid log entry | bad timestamp, unknown level, junk line | counted as malformed and skipped; the run continues |
| Execution error | file missing, read failure, output file cannot be created or written | the run stops, the error goes to stderr, exit code 1 |

Details:

- `main` only parses flags and turns an error into an exit code. The work happens in `run`, which returns an `error`. This keeps deferred `Close` calls working, because `os.Exit` skips deferred functions.
- `scanner.Err()` is checked after the loop, so a read failure in the middle of the file is not reported as a complete result.
- The JSON file is created only after analysis succeeds. The error from `Close()` is checked, because buffered data can fail to write at close time. If writing fails, the partial file is removed.
- Invalid arguments print the usage message and exit with code 2.

