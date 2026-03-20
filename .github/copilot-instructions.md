# GitHub Copilot Instructions

Follow these repository instructions when working in this project.

## General guidance

- Keep changes focused and consistent with the current Go-based crawler workflow.
- Write new or updated repository instructions, comments, and documentation in English.
- Avoid local absolute paths, machine-specific assumptions, and hardcoded credentials.
- Treat downloaded artifacts and runtime output as generated data, not hand-maintained source files.
- Keep network fetching, extraction logic, and history tracking clearly separated.

## Project context

- The main implementation lives in `crawler.go`.
- Dependency management is defined in `go.mod` and `go.sum`.
- Downloaded ZIP files are written to `downloads/`, and the main JAR is extracted from them into the same directory.
- GitHub Actions workflows live under `.github/workflows/`.

## Validation

- Prefer `go test ./...` when tests exist or when new tests are added.
- Prefer `go build ./...` for build verification after logic changes.
- Clearly distinguish between checks you ran and checks you did not run.
