# Repository Guidelines

## Project Structure & Module Organization
The repository centers on two crawlers: `crawler.go` (Go implementation) and `crawler.py` (Python legacy). `downloads/` holds fetched driver archives; it is generated at runtime. Dependency manifests live in `go.mod`, `go.sum`, `pyproject.toml`, and `uv.lock`. Keep auxiliary scripts or docs in clearly named top-level files; future tests should reside under `tests/` as `test_*.py` or Go equivalents.

## Build, Test, and Development Commands
Run the production job with `go run ./crawler.go`; it downloads the latest BigQuery JDBC drivers and updates `download_history.txt`. For Python parity or quick experiments, use `uv run python crawler.py`. Refresh Go modules via `go mod tidy` and sync Python dependencies with `uv sync`. Perform a syntax check on the Python crawler using `python -m py_compile crawler.py` before submitting changes.

## Coding Style & Naming Conventions
Adopt Go's default formatting (`gofmt`) for `.go` files. Python code targets 4-space indentation, 120-character lines, and full typing per `pyproject.toml`. Functions use `snake_case`, classes `PascalCase`, and constants `UPPER_SNAKE_CASE`. Logging is standardized through `structlog`; HTTP calls use `httpx` with `follow_redirects=True`. Run `ruff check .` and `ruff format .` to maintain Python style.

## Testing Guidelines
Formal test suites are not yet in place. When adding coverage, prefer `pytest` for Python (`pytest -q`) and Go table-driven tests (`go test ./...`). Focus on HTML parsing, exclusion rules, and archive extraction logic. Name tests descriptively (e.g., `test_parse_release_table`) and keep fixtures small to ease review.

## Commit & Pull Request Guidelines
Follow the "type: summary" commit style, such as `feat: add go crawler` or `chore: refresh deps`. Ensure diffs exclude generated artifacts except updated driver archives. Pull requests should summarize the motivation, list key changes, describe validation (commands run, CI links), and reference related issues. Include representative logs when altering network interactions.

## Security & Configuration Tips
The crawler requires outbound HTTPS. Configure `HTTP_PROXY`/`HTTPS_PROXY` in restricted environments. Confirm that `downloads/` is writable on CI runners. Review third-party download URLs before approving changes to prevent supply-chain risks.
