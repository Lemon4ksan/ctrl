# GEMINI.md — Development Guidelines & AI Assistant Protocol

This document outlines the architecture, coding standards, build/test commands, and interaction rules for AI assistants and software engineers working in the **ctrl** repository.

## 1. Project Overview

**`ctrl`** is a unified, zero-dependency CLI application framework and high-performance terminal UI toolkit for Go. It eliminates runtime reflection, boilerplate flag wiring, and bloated third-party dependencies (`cobra`, `pflag`, `bubbletea`, `lipgloss`) by delivering pure-Go, standard-library-backed command routing, POSIX flag parsing, and hardware-efficient terminal styling.

> **Core Engineering Manifesto**:
> _«The command line is the pure interface between human intention and binary execution. In the terminal, latency and clutter are the default — let ctrl be your sharp, zero-dependency command surface.»_

> **The "ctrl v1" Compatibility & Forever-Frozen Core Manifesto**:
> _"Code written for **ctrl v1.0.0** is guaranteed to compile and run unchanged on any **v1.x** version 5, 10, and 20 years from now. The core CLI contracts and terminal primitives are permanent."_

### Key Capabilities & Architectural Pillars
- **Zero External Dependencies**:
  - Implements POSIX flag parsing, subcommand routing, terminal width probing, ANSI color control, tables, boxes, progress bars, and badges using only Go standard library primitives.
- **Full POSIX Flag Semantics (`argkit`)**:
  - Interspersed positional arguments and flags (`cmd arg1 -v arg2`).
  - Short flag clumping / stacking (`-la` expands to `-l` and `-a`).
  - Attached values (`-I*.tmp` -> `-I=*.tmp`).
  - Strict POSIX terminator `--` (arguments after `--` treated as literal positionals).
  - Levenshtein-distance fuzzy typo suggestions ("Did you mean: --flag?").
  - Shell tokenizer fragment normalization (PowerShell / CMD splitting repair).
- **Zero-Friction CLI Application Framework (`tuikit.App` / `ctrl`)**:
  - Clean subcommand routing, aliases, grouped help listings, and default fallback commands.
  - Automatic `help`, `-h`, `--help`, `version`, `-v`, `--version` handling.
- **Terminal UI & Styling Engine (`tuikit`)**:
  - Terminal capability sniffing (`ProbeTerminal`), Windows Virtual Terminal processing enablement.
  - Global `NO_COLOR` / `TERM=dumb` awareness via lock-free atomic toggles.
  - Monospaced visual width calculation (`VisibleWidth`) ignoring ANSI escapes.
  - Bordered boxes (`Box`), auto-aligned tables (`Table`), proportional progress bars (`Progress`, `RenderBar`), sparklines (`RenderSparkline`), badges (`BadgePass`, `BadgeFail`, `BadgeWarn`), and pipeline steps (`RenderStep`).

---

## 2. Repository Layout

```text
ctrl/
├── ctrl.go, version.go, doc.go       // Root facade, version metadata, and package documentation
├── argkit/                           // POSIX argument parser, flag clumper, typo suggester
│   ├── args.go                       // FlagSet parser, short-flag unbundling, Levenshtein distance
│   ├── doc.go                        // Package documentation and stdlib comparison
│   └── *_test.go                     // Comprehensive unit and adversarial flag tests
├── tuikit/                           // Terminal UI toolkit, widgets, and CLI application router
│   ├── app.go                        // Subcommand application runner and help generator
│   ├── command.go                    // Subcommand interface contract and functional adapter
│   ├── probe*.go                     // OS-specific terminal TTY and VT-processing detection
│   ├── style.go                      // ANSI styling, string padding, visible width, line iterator
│   ├── box.go                        // Bordered cards and panels (single, rounded, heavy)
│   ├── table.go                      // Auto-aligned tabular data with column alignments
│   ├── progress.go                   // Live terminal progress tracker, rate, and ETA calculation
│   ├── bar.go                        // Proportional ASCII/Unicode bars and sparkline graphs
│   ├── badge.go                      // Status badges (PASS, FAIL, WARN, INFO, ACTIVE)
│   ├── step.go                       // Numbered pipeline stage formatting with dotted connectors
│   ├── doc.go                        // Package documentation
│   └── *_test.go                     // Unit tests for terminal formatting and widgets
└── .agents/                          // AI assistant guidelines and capabilities
```

---

## 3. Core Architectural Principles & Technical Constraints

1. **Zero External Dependencies**:
   - Do not introduce third-party Go dependencies. Every capability must rely solely on Go standard library packages (`flag`, `fmt`, `io`, `os`, `strings`, `sync`, `unicode/utf8`, etc.).
2. **Zero-Allocation Hot Paths**:
   - In string measuring (`VisibleWidth`), rune decoding, and Levenshtein distance calculation for common flag lengths ($\le 64$ chars), use stack-allocated arrays (`[65]int`, `[64]rune`) to eliminate heap pressure.
3. **Respect Environment & TTY Semantics**:
   - Always honor `NO_COLOR` and non-interactive pipes. Progress bars and color escapes must gracefully switch to clean, periodic plain-text output when running inside CI/CD or piped to a file.
4. **Clean, Uncluttered Visual Design**:
   - Terminal outputs must remain legible, restrained, and professional. Avoid visual clutter, gratuitous blinking, or intrusive ASCII borders.

---

## 4. Code Style & Quality Standards

- **Go Version**: 1.27+ (using modern generics and standard library iterators).
- **Documentation**: Every exported package identifier (type, struct, function, interface, error constant) **must** have clear Godoc documentation.
- **Code Formatting & Linters**:
  - The repository adheres to `.golangci.yml`.
  - Run `go fmt ./...` and `golangci-lint run`.

---

## 5. Development & Testing Commands

| Command | Purpose |
| :--- | :--- |
| `go test ./...` | Run fast unit tests across all packages |
| `go test -race ./...` | Run full test suite with Go data race detector |
| `go test -bench=. -benchmem ./...` | Run benchmark suite |
| `make lint` | Run `golangci-lint` verification |
| `make format` | Format code (`gofmt`, `golangci-lint --fix`) |

---

## 6. Commit & Pull Request Guidelines

### Conventional Commits Format:
```text
<type>(<scope>): <short summary>

<detailed description of changes>
```

Commit Types:
- `feat`: New capability or widget
- `fix`: Bug fix
- `impr`: Performance optimization or allocation reduction
- `refactor`: Structural cleanup without behavioral changes
- `docs`: Documentation updates
- `test`: Adding or updating unit tests
- `chore`: Build scripts or tooling updates
