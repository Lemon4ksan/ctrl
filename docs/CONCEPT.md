# ctrl — Philosophy & Architecture

```text
                  ┌─────────────────────────────────────────┐
                  │                 C T R L                 │
                  │        The Command Surface for Go       │
                  └─────────────────────────────────────────┘
```

> *«The command line is the pure interface between human intention and binary execution. In the terminal, latency and clutter are the default — let ctrl be your sharp, zero-dependency command surface.»*

## 1. The Core Philosophy: Zero-Dependency Precision

For years, building command-line applications and terminal tools in Go meant dragging in massive dependency trees:
- `spf13/cobra` + `spf13/pflag` + `spf13/viper` (pulling in hundreds of transitive dependencies, reflection layers, and convoluted lifecycle hooks).
- `charmbracelet/bubbletea` + `lipgloss` + `bubbles` (forcing complex Elm-architecture state machines and heavy styling overhead for simple CLIs).

**`ctrl` eliminates this dependency gravity.**

In `ctrl`, command-line routing, POSIX flag parsing, and terminal styling are implemented with **zero external dependencies**, strictly building upon Go's standard library while overcoming its historical limitations.

---

## 2. The Architectural Pillars

```text
                     ┌───────────────────────────────────┐
                     │          Raw OS Arguments         │
                     │       (os.Args / Shell Pipes)     │
                     └─────────────────┬─────────────────┘
                                       │
                ┌──────────────────────┴──────────────────────┐
                ▼                                             ▼
     ┌──────────────────────┐                      ┌──────────────────────┐
     │ NormalizeArgs        │                      │ Terminal Sniffer     │
     │ (Shell Token Repair) │                      │ (TTY & NO_COLOR)     │
     └──────────┬───────────┘                      └──────────┬───────────┘
                │                                             │
                └──────────────────────┬──────────────────────┘
                                       ▼
                     ┌───────────────────────────────────┐
                     │     POSIX Flag Decomposer         │
                     │ (-la clumping, attached values,   │
                     │  interspersed args, typo suggest) │
                     └─────────────────┬─────────────────┘
                                       ▼
                     ┌───────────────────────────────────┐
                     │      Subcommand App Router        │
                     │   (tuikit.App / CommandGroups)    │
                     └─────────────────┬─────────────────┘
                                       ▼
                     ┌───────────────────────────────────┐
                     │       Terminal Presentation       │
                     │  (Table, Box, Progress, Badges)   │
                     └───────────────────────────────────┘
```

### I. Full POSIX Semantics over `flag.FlagSet` (`argkit`)
Standard Go's `flag` package makes an unfortunate compromise: it halts flag parsing at the first positional argument, forbids short flag clumping, and provides no typo assistance. `argkit` wraps `flag.FlagSet` to provide:
- **Interspersed Parsing**: Positional arguments and flags can appear in any order (`tool src/ -v --out=bin dst/`).
- **Short Flag Clumping**: `-la` automatically decomposes into `-l` and `-a`.
- **Attached Values**: `-I*.tmp` maps directly to `-I=*.tmp`.
- **POSIX Terminator**: `--` strictly stops flag parsing, treating all following tokens as positional arguments.
- **Fuzzy Typo Suggestions**: Levenshtein distance matching identifies mistyped flags and prints intuitive suggestions (`"did you mean: --timeout?"`).
- **Zero Allocations on Common Paths**: Levenshtein calculations for strings up to 64 runes run entirely on stack buffers without heap allocations.

### II. Clean Subcommand Routing (`tuikit.App` / `ctrl.App`)
- Simple, testable command contracts ([`Command`](file:///d:/CodingProjects/ctrl/tuikit/command.go) and [`FuncCommand`](file:///d:/CodingProjects/ctrl/tuikit/command.go)).
- Automatic, legible help output organized into logical `CommandGroups`.
- Native handling of global flags (`--help`, `-h`, `--version`, `-v`).
- Independent input/output redirection (`io.Writer`) for 100% deterministic unit testing.

### III. Hardware-Efficient Terminal Presentation (`tuikit`)
- **TTY Capability Sniffing**: Automatic TTY detection, Windows 10/11 Virtual Terminal escape sequence enablement, and `NO_COLOR` / `TERM=dumb` awareness.
- **Visible Width Engine**: Accurately computes monospaced character widths (`VisibleWidth`) by stripping ANSI escape sequences.
- **Table**: Pixel-perfect columnar alignment with left, right, or center justification.
- **Box**: Clean card containers with configurable borders (`BorderSingle`, `BorderRounded`, `BorderHeavy`) and guaranteed straight borders.
- **Progress & Bars**: Smooth progress bars with throughput calculation, ETA estimation, and transparent fallback to low-frequency log lines when running in CI or piped outputs.
- **Badges & Steps**: Universal status tags (`✔ PASS`, `❌ FAIL`, `⚠️ WARN`, `⚡ ACTIVE`) and numbered pipeline stages (`[1/4] Build ........... ✔ PASSED`).

---

## 3. The Ecosystem Relationship

Within the broader engineering ecosystem:
- **`foundation`**: Replaces Go's stdlib for low-level systems (SIMD, memory arenas, L4 transport, cryptographic primitives). CLI/TUI logic is intentionally excluded.
- **`ctrl`**: The official CLI application framework and terminal UI toolkit across all tools in the ecosystem.
- **`aoni`**, **`sein`**, **`przm`**, **`seal`**, **`vortex`**: Consumer projects that rely on `ctrl` for their binary command interfaces, diagnostics formatting, progress visualization, and shell autocompletion.
