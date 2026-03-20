---
applyTo: "*.go,go.mod,go.sum"
---

When editing Go files in this repository:

- Keep the crawler logic easy to follow and centered on one responsibility per function.
- Preserve stable handling for URL extraction, exclusion rules, downloads, and archive extraction.
- Avoid baking environment-specific proxy or filesystem settings into source code.
- Keep dependency changes deliberate and aligned with the crawler's actual needs.
