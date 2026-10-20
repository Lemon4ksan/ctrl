<img src="assets/logo.svg" align="right" width="128" height="128" alt="ctrl logo" />

# ctrl

_«In the terminal, latency and clutter are the default — let ctrl be your sharp, zero-dependency command surface»_

[![Go Version](https://img.shields.io/badge/go-1.27%2B-007d9c?logo=go&logoColor=white&style=flat-square)](https://go.dev/)
[![Go Reference](https://img.shields.io/badge/godoc-reference-007d9c?style=flat-square)](https://pkg.go.dev/github.com/lemon4ksan/ctrl)
[![Architecture](https://img.shields.io/badge/arch-Zero--Alloc%20%7C%20POSIX%20%7C%20MCP-0d9488?style=flat-square)](#rules)
[![Ecosystem](https://img.shields.io/badge/ecosystem-foundation-0284c7?style=flat-square)](https://github.com/lemon4ksan/foundation)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue?style=flat-square)](LICENSE)

**[English](README.md)** | **[Русский](README_RU.md)** | **[Architecture Concept](docs/CONCEPT.md)**

Command-line application framework, terminal UI toolkit (`tuikit`), Model Context Protocol v1.0 router (`mcpkit`), and Windows UI Automation engine (`uiakit`) for Go 1.27+. Built directly on top of the Go standard library and [`foundation`](https://github.com/lemon4ksan/foundation) with zero third-party dependencies.

## Design tradeoffs

| Standard library / `cobra` / `lipgloss` | What `ctrl` does instead | What you pay |
| :--- | :--- | :--- |
| `flag.FlagSet` halts at the first positional argument, rejects `-la` short-flag clumping, and offers no typo hints | [`argkit`](argkit) pre-normalizes shell tokens and decomposes POSIX short/long flags on the stack before feeding `flag.FlagSet`, with stack-allocated Levenshtein suggestions (`0 allocs/op`) | Levenshtein stack buffers cap typo distance calculation at flag names up to 64 runes |
| Full-screen TUI frameworks force an Elm event loop and alternate screen buffer for simple CLI tools | [`tuikit`](tuikit) streams tables, boxes, progress bars, sparklines, and badges directly to an `io.Writer`, automatically degrading to plain periodic logs when `NO_COLOR` is set or `stdout` is piped | Not an interactive full-screen forms/window manager; designed for streaming CLI output and status surfaces |
| Hand-written MCP servers require nested `map[string]any` JSON schemas and manual JSON-RPC framing | [`mcpkit`](mcpkit) provides a fluent `NewTool` builder, generic `TypedHandler[T]` adapters, and STDIO / HTTP+SSE transports with 16 MiB scanner buffers and token-bucket rate limiting | Targets the MCP `2024-11-05` (`v1.0`) wire contract rather than arbitrary JSON-RPC dialects |
| Accessibility inspection on Windows usually requires C++/C# wrappers or CGO | [`uiakit`](uiakit) invokes `CUIAutomation8` (`IUIAutomation2`) COM vtables directly via `foundation/silicon/comkit` with bulk `IUIAutomationCacheRequest` prefetching | Windows-specific (`UIAutomationCore.dll`); non-Windows targets compile cleanly against a no-op stub |

## Rules

1. **Zero third-party dependencies.** `ctrl` depends only on the Go standard library and [`foundation`](https://github.com/lemon4ksan/foundation). No reflection-heavy struct tag mappers or transitive CLI dependency trees.
2. **Zero-allocation hot paths.** Monospaced width calculation (`VisibleWidth`), ANSI stripping, `LinesSeq` push iteration, and Levenshtein flag suggestions execute with `0 B/op`.
3. **Explicit I/O and styling state.** Commands receive `(ctx, args, stdout, stderr)` explicitly, and `tuikit.Styler` instances carry isolated color policy without mutating global state.
4. **Automatic TTY & CI degradation.** Live progress bars and ANSI sequences probe terminal capabilities (`ProbeTerminal`) and fall back to clean, machine-parsable lines in non-interactive pipes.

## Layout

- [`ctrl` (root)](ctrl.go) — Top-level facade re-exporting `App`, `Command`, `GroupCommand`, `HubCommand`, `ParseInterspersedFlags`, `Table`, `Box`, `Progress`, and status badges.
- [`argkit`](argkit) — POSIX argument engine: interspersed flags and positionals (`app -v src/ -o=bin dst/`), short-flag clumping (`-la` $\to$ `-l -a`), attached values (`-I*.tmp` $\to$ `-I=*.tmp`), `--` terminator, PowerShell/CMD quote repair (`NormalizeArgs`), and fuzzy typo suggestions.
- [`tuikit`](tuikit) — Subcommand router (`App`, `CommandGroup`), terminal capability prober (`ProbeTerminal`), auto-aligned `Table`, bordered `Box` (`BorderSingle`, `BorderRounded`, `BorderHeavy`), live `Progress` with throughput/ETA, inline sparklines (`RenderSparkline`: ` ▂▃▅▆▇█`), and pipeline badges.
- [`mcpkit`](mcpkit) — Model Context Protocol (`2024-11-05`) server over STDIO and HTTP/SSE with Bearer auth, TLS, rate limiting, fluent schema builders (`NewTool`, `Required`, `Min`, `Max`, `Default`), and generic `TypedHandler[T]` / `TextHandler[T]` adapters.
- [`uiakit`](uiakit) — Pure-Go Windows COM UI Automation (`CUIAutomation8`) engine extracting Win32, WPF, WinUI 3, Qt, and Chromium `RootWebArea` trees into Set-of-Marks Markdown digests, plus background `InvokePattern`, `ValuePattern`, `TogglePattern`, and `SelectionItemPattern` execution.

## Installation

```bash
go get github.com/lemon4ksan/ctrl
```

## Usage

### CLI Router with POSIX Flag Clumping & Styled Output

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/lemon4ksan/ctrl"
)

func main() {
	app := ctrl.NewApp("seal", ctrl.Version, "Archive & compression utility",
		ctrl.NewCommand("create", "Create an archive from files", "seal create [flags] <archive> <files...>",
			func(ctx context.Context, args []string, stdout, stderr io.Writer) error {
				fs := flag.NewFlagSet("create", flag.ContinueOnError)
				level := fs.Int("l", 3, "Compression level (1-9)")
				verbose := fs.Bool("v", false, "Verbose output")

				posArgs, err := ctrl.ParseInterspersedFlags(fs, args)
				if err != nil {
					return err
				}
				if len(posArgs) < 2 {
					return fmt.Errorf("archive path and target files required")
				}

				fmt.Fprintf(stdout, "%s Creating %s (level=%d, verbose=%v)\n",
					ctrl.BadgePass(), posArgs[0], *level, *verbose)
				return nil
			},
			"c", "add",
		),
	)

	app.CommandGroups = []ctrl.CommandGroup{
		{Title: "Core Operations", Commands: []string{"create"}},
	}

	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", ctrl.BadgeFail(), err)
		os.Exit(1)
	}
}
```

### Fluent MCP v1.0 Tool Server (`mcpkit`)

```go
package main

import (
	"context"

	"github.com/lemon4ksan/ctrl/mcpkit"
)

type ExecArgs struct {
	Command   string `json:"command"`
	TimeoutMS int    `json:"timeout_ms"`
}

func main() {
	srv := mcpkit.NewServer("agent-tools", "1.0.0")

	tool := mcpkit.NewTool("exec_command", "Execute a shell command").
		String("command", "Command line to run", mcpkit.Required()).
		Int("timeout_ms", "Execution timeout in milliseconds", mcpkit.Min(100), mcpkit.Default(5000)).
		Build()

	srv.RegisterTool(tool, mcpkit.TextHandler(func(ctx context.Context, args ExecArgs) (string, error) {
		return "executed: " + args.Command, nil
	}))

	_ = srv.ServeStdio(context.Background())
}
```

## Verification

```bash
# Run unit tests
make test

# Run test suite with data race detector
make race

# Run benchmarks and verify 0 allocs/op on hot paths
make bench

# Run linters
make lint
```

## License

BSD 3-Clause License. See [LICENSE](LICENSE) for full text.
