# GhostOps

GhostOps is a command-line tool that streams an application log file, computes statistics, detects predefined incident patterns and prints a report. It can also export the report as JSON.

It uses only the Go standard library and reads the input line by line, so large files do not need to fit in memory.

## Requirements

- Go 1.27 or newer

## Build

```sh
go build -o ghostops .
```

Or run it without building:

```sh
go run . logs/sample.log
```

## Usage

```
ghostops [-o file] <logfile>
```

| Argument | Description |
|---|---|
| `<logfile>` | Path to the log file to analyze (required, exactly one) |
| `-o file` | Also write the report as JSON to `file` |

The human-readable report is always printed to standard output. With `-o`, the JSON report is written to the given file as well.

Flags must come **before** the log file, as with Go's standard `flag` package:

```sh
./ghostops -o output/report.json logs/sample.log   # works
./ghostops logs/sample.log -o output/report.json   # rejected with a usage message
```

### Exit codes

| Code | Meaning |
|---|---|
| `0` | Report generated successfully |
| `1` | Execution error: the log file could not be opened or read, or the output could not be written |
| `2` | Invalid arguments (missing log file, extra arguments, unknown flag) |

Malformed log lines are **not** errors. They are counted and skipped, and the run still succeeds. If writing the JSON file fails, the partial file is removed.

## Log format

Each line is one entry:

```
<RFC3339 timestamp> <LEVEL> <service> <message>
```

Example:

```
2026-10-02T10:22:03Z ERROR payment-service Payment gateway timeout - transaction aborted
```

- `LEVEL` is one of `INFO`, `WARN`, `ERROR`, `FATAL`.
- Entries do not need to be in chronological order.
- Fields are separated by one or more spaces; extra spaces are ignored, and runs of spaces inside the message are collapsed to one.
- A line is malformed if it has fewer than four fields, a timestamp that is not RFC3339, or an unknown level.
- Lines longer than 64 KB stop the run with a read error.

## Incident rules

A log entry matches a rule when its level equals the rule's level **and** its message contains one of the rule's keywords (case-insensitive). Rules are checked in this order and the first match wins, so each entry belongs to at most one incident.

| Priority | Incident | Level | Keywords |
|---|---|---|---|
| 1 | Crashed | FATAL | `crashed` |
| 2 | Database | ERROR | `database`, `sql` |
| 3 | Timeout | ERROR | `timeout`, `timed out` |
| 4 | BadGateway | ERROR | `bad gateway` |
| 5 | UncaughtException | ERROR | `nullpointerexception` |
| 6 | DatabasePoolPressure | WARN | `pool` |
| 7 | Memory | WARN | `memory` |
| 8 | RateLimit | WARN | `rate`, `limit` |
| 9 | API | WARN | `deprecated`, `api` |

For example, `Failed to connect to database: timeout after 5000ms` matches both Database and Timeout, and is reported as Database.

## Example output

Terminal report for `logs/sample.log` (incident list shortened):

```
GhostOps Report
===============

Summary
-------
  Total lines      6000
  Valid entries    5532  92.2%
  Malformed lines  468   7.8%

Entries by severity
-------------------
  INFO   3335  60.3%
  WARN   1106  20.0%
  ERROR  820   14.8%
  FATAL  271   4.9%

Entries by service
------------------
  auth-service          1106  20.0%
  notification-service  1130  20.4%
  order-service         1082  19.6%
  payment-service       1090  19.7%
  user-service          1124  20.3%

Incidents (9 detected)
----------------------
  1. Crashed [FATAL]
     Occurrences:  271
     First seen:   2026-10-01T00:52:07Z
     Last seen:    2026-10-06T18:49:45Z
     Duration:     137h57m38s
     Services:     order-service (63), user-service (57), auth-service (52), notification-service (50), payment-service (49)
  ...
```

The full JSON report for the same file is in [`output/report.json`](output/report.json). Its top-level shape is:

```json
{
  "summary": { "total_lines": 6000, "valid_entries": 5532, "malformed_lines": 468 },
  "entries_by_severity": [ { "name": "INFO", "count": 3335 } ],
  "entries_by_service": [ { "name": "auth-service", "count": 1106 } ],
  "incidents": [
    {
      "type": "Crashed",
      "severity": "FATAL",
      "occurrences": 271,
      "first_seen": "2026-10-01T00:52:07Z",
      "last_seen": "2026-10-06T18:49:45Z",
      "duration_seconds": 496658,
      "services": [ { "name": "order-service", "count": 63 } ]
    }
  ]
}
```

Lists are shortened above. In the real output every list is complete and always present: an input with no valid entries or no incidents produces `[]`, never `null`.

## Project layout

```
main.go               CLI: flags, files, exit codes, read loop
internal/parser       parses one line into a Log entry
internal/analysis     statistics and the Analyzer that collects results
internal/incidents    incident rules and incident groups
internal/report       Report model, Reporter interface, text and JSON output
logs/sample.log       sample input
output/report.json    example JSON output
```

See [`doc.md`](doc.md) for the architecture and design decisions.
