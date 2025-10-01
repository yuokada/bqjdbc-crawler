# Repository Guidelines

This document provides the minimal guidelines for participating in development and maintenance of this repository. The content will be updated as needed.

## Project Structure and Layout
- `crawler.py`: Logic to list, download, and extract BigQuery JDBC drivers.
- `download_history.txt`: Record of URLs that have already been fetched.
- `pyproject.toml` / `uv.lock`: Dependencies and tool settings (Python 3.12, ruff, etc.).
- `downloads/`: Output destination for downloaded artifacts (created automatically as needed).

## Build, Run, and Development Commands
- Run: `python crawler.py`
- Quick syntax check: `python -m py_compile crawler.py`
- Formatting/Lint (ruff): `ruff check .` / `ruff format .`
  - Example: `uv run ruff check .` (when using uv)

## Coding Conventions and Naming
- Python 3.12, 4-space indentation, 120 character line length (per `pyproject.toml`).
- Type hints are required; functions use `snake_case`, classes use `PascalCase`, and constants use `UPPER_SNAKE_CASE`.
- Use `structlog` for logging. HTTP client is `httpx` (synchronous, with `follow_redirects=True`).

## Testing Policy
- No tests are currently introduced. In the future, adding tests under `tests/` as `test_*.py` is recommended.
- Guideline: Unit test key functions (HTML parsing, exclusion logic, extraction process). Coverage targets can be introduced gradually.
- Example (after introduction): `pytest -q`.

## Commit/PR Guide
- Keep commits small and meaningful. Use the message style “type: summary” (e.g., `feat: switch to httpx`) when possible.
- PRs should include an overview, list of changes, verification steps, and related issues. Screenshots of logs/output are welcome.
- Pre-submit check: Ensure `ruff check .` reports no warnings. Do not include unnecessary diffs (artifacts/caches).

## Security and Configuration Tips
- Network access is required (drivers are fetched at runtime). In proxy environments, set `HTTP(S)_PROXY` appropriately.
- Ensure the output directory `downloads/` is writable. After unzipping, JAR files are saved with a prefixed name.
