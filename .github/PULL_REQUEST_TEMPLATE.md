## Description

Provide a concise summary of the change. Link any related issues below.

Fixes # (issue)

## Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to break)
- [ ] Performance improvement
- [ ] Refactoring (no functional changes)
- [ ] Documentation update

## Checklist

- [ ] I have run `make lint` — no new linter warnings (0 issues reported)
- [ ] I have run `make test` or `make race` — all unit tests pass cleanly
- [ ] New functionality is covered by unit tests
- [ ] Commits follow Conventional Commits formatting (`feat:`, `fix:`, `perf:`, `docs:`)
- [ ] Public API changes are documented in `doc.go` and `README.md` / `README_RU.md`
- [ ] Zero external dependencies invariant is preserved
- [ ] Terminal escape sequences and `NO_COLOR` / non-interactive pipe contracts are respected

## How has this been tested?

Describe how you verified the changes work:

```bash
go test -race ./...
```
