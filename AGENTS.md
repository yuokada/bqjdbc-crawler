# Repository Guidelines

## Project Structure & Module Organization
The repository centers on `crawler.go` (Go implementation). `downloads/` holds fetched driver archives; it is generated at runtime. Dependency manifests live in `go.mod` and `go.sum`. Keep auxiliary scripts or docs in clearly named top-level files; future tests should reside under Go package directories as `*_test.go`.

## Build, Test, and Development Commands
Run the production job with `go run ./crawler.go`; it downloads the latest BigQuery JDBC drivers and updates `download_history.txt`. Refresh dependencies with `go mod tidy`.

## Coding Style & Naming Conventions
Adopt Go's default formatting (`gofmt`) for `.go` files.

## Testing Guidelines
Formal test suites are not yet in place. When adding coverage, prefer Go table-driven tests (`go test ./...`). Focus on HTML parsing, exclusion rules, and archive extraction logic.

## Commit & Pull Request Guidelines
Follow the "type: summary" commit style, such as `feat: add go crawler` or `chore: refresh deps`. Ensure diffs exclude generated artifacts except updated driver archives. Pull requests should summarize the motivation, list key changes, describe validation (commands run, CI links), and reference related issues. Include representative logs when altering network interactions.

## Security & Configuration Tips
The crawler requires outbound HTTPS. Configure `HTTP_PROXY`/`HTTPS_PROXY` in restricted environments. Confirm that `downloads/` is writable on CI runners. Review third-party download URLs before approving changes to prevent supply-chain risks.
