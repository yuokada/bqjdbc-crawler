# bqjdbc-crawler

A Go crawler that scans the BigQuery JDBC driver distribution page, downloads new ZIP files, and extracts the JAR.

## Overview

- Implementation: `crawler.go` (Go)
- Download directory: `downloads/`
- Download history: `download_history.txt`
- Target page: <https://cloud.google.com/bigquery/docs/reference/odbc-jdbc-drivers>

## Requirements

- Go 1.24+
- Outbound HTTPS access

## Usage

```bash
go run ./crawler.go
```

Main processing steps:

1. Extract JDBC-related links from the driver distribution page.
2. Skip old drivers listed in the exclusion list.
3. Save new ZIP files to `downloads/`.
4. Extract `GoogleBigQueryJDBC42.jar` from each ZIP and save it as `downloads/<zip-name>-GoogleBigQueryJDBC42.jar`.
5. Append downloaded URLs to `download_history.txt`.

## Development Commands

```bash
# Resolve dependencies
go mod tidy

# Build
go build ./...

# Static analysis
go vet ./...

# Test
go test ./...

# Format
gofmt -w .
```

## GitHub Actions

- `Go CI` (`.github/workflows/go.yml`)
  - Runs build, vet, gofmt checks, and tests on push/PR to `master`.
- `Update download_history.txt` (`.github/workflows/crawler.yaml`)
  - Runs the crawler every Monday at 08:00 UTC (and via manual dispatch).
  - Pushes a commit to `master` when changes are detected.
