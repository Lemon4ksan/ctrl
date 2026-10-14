# Contributing to ctrl

Thank you for considering a contribution! `ctrl` is a zero-dependency, high-performance CLI application framework and terminal UI toolkit for Go. We welcome bug fixes, performance optimizations, and ergonomics refinements.

## Ground Rules

- **Zero External Dependencies.** `ctrl` provides argument parsing, POSIX flag clumping, subcommand routing, and styled terminal primitives with standard library Go and zero third-party dependencies.
- **No Terminal Clutter.** Clean, readable outputs. Avoid garish color noise, redundant borders, or brittle terminal escape assumptions. Always respect `NO_COLOR` and non-TTY execution environments.
- **Stability & Performance.** Hot paths (width calculation, string padding, flag decomposing, argument token normalization) must minimize heap allocations and respect CPU cache locality.
- **Backward Compatibility.** Public API breaks require a major version bump and architectural rationale.

## Getting Started

```bash
git clone https://github.com/lemon4ksan/ctrl
cd ctrl
go mod download
go test -race ./...
```

## Workflow

1. **Fork** the repository and create a feature branch from `main`.
2. Write your code following the existing style (check `.golangci.yml`).
3. Add or update unit tests — ensure `go test -race ./...` and `golangci-lint run` pass cleanly.
4. Open a Pull Request against `main`.

## Commit Style

Use conventional commits:

```text
feat: add multi-progress bar and indeterminate task group
fix: handle multibyte runes correctly in table cell width calculation
perf: reduce allocations in flag clump decomposition
docs: update quickstart guide with subcommand routing
```

## Code Style

- Run `gofmt` before committing (`golangci-lint` enforces this).
- Keep all exported identifiers documented with Godoc comments.
- Favour zero-allocation stack buffers for string measurements and distance calculations.
