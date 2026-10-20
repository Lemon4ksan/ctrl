<div align="center">

<img src="assets/logo.svg" width="120" height="120" alt="ctrl" />

# ctrl

### Zero-Dependency CLI, TUI, MCP & UI Automation Framework for Go

_«In the terminal, latency and clutter are the default — let ctrl be your sharp, zero-dependency command surface»_

[![Go Version](https://img.shields.io/badge/go-1.27%2B-007d9c?logo=go&logoColor=white&style=flat-square)](https://go.dev/)
[![Go Reference](https://img.shields.io/badge/godoc-reference-007d9c?style=flat-square)](https://pkg.go.dev/github.com/lemon4ksan/ctrl)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue?style=flat-square)](LICENSE)
[![Zero-Dependencies](https://img.shields.io/badge/dependencies-0%20external-brightgreen?style=flat-square)](#features--architectural-pillars)
[![Zero-Alloc](https://img.shields.io/badge/memory-0%20B%2Fop%20%7C%200%20allocs-success?style=flat-square)](#performance--zero-allocation-guarantees)
[![Ecosystem](https://img.shields.io/badge/ecosystem-foundation-orange?style=flat-square)](https://github.com/lemon4ksan/foundation)
[![POSIX Semantics](https://img.shields.io/badge/POSIX-Flags%20%26%20Clumping-blueviolet?style=flat-square)](argkit)

**ctrl** is a zero-dependency CLI application framework, terminal UI toolkit, Model Context Protocol (`mcpkit`) router, and Windows UI Automation (`uiakit`) engine for Go. It provides full POSIX flag parsing with short-flag clumping and attached values, subcommand routing with command grouping, terminal capability sniffing, terminal UI styling primitives (tables, boxes, progress bars, sparklines, badges), fluent MCP v1.0 tool builders, and zero-allocation COM accessibility traversal.

#### English • [Русский](README_RU.md) • [Architecture Concept](docs/CONCEPT.md)

</div>

## Installation

`ctrl` requires Go version `1.27` or higher.

```bash
go get github.com/lemon4ksan/ctrl
```

## Quickstart

Build a CLI tool with subcommands, POSIX flag clumping, and styled terminal output in minutes:

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/lemon4ksan/ctrl"
	"github.com/lemon4ksan/ctrl/tuikit"
)

func main() {
	app := ctrl.NewApp("seal", ctrl.Version, "High-performance compression & archiving utility",
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

				fmt.Fprintf(stdout, "%s Creating archive %s (level %d, verbose=%v)\n",
					tuikit.BadgePass(), posArgs[0], *level, *verbose)
				return nil
			},
			"c", "add",
		),
	)

	app.CommandGroups = []ctrl.CommandGroup{
		{
			Title:    "Core Operations",
			Commands: []string{"create"},
		},
	}

	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", tuikit.BadgeFail(), err)
		os.Exit(1)
	}
}
```

## Features & Architectural Pillars

| Capability | Standard Library (`flag`) | `spf13/cobra` | `ctrl` |
| :--- | :---: | :---: | :---: |
| **External Dependencies** | **0** | ~15+ transitive | **0** |
| **Interspersed Flags** | ✗ (halts at first positional) | ✓ | **✓ (`argkit`)** |
| **Short Flag Clumping (`-la`)** | ✗ | ✓ | **✓ (`argkit`)** |
| **Attached Values (`-I*.tmp`)** | ✗ | ✓ | **✓ (`argkit`)** |
| **Fuzzy Typo Suggestions** | ✗ | ✓ | **✓ (Levenshtein)** |
| **Subcommand Routing** | ✗ | ✓ | **✓ (`tuikit.App`)** |
| **Terminal Capability Sniffing** | ✗ | ✗ (separate lib) | **✓ (`tuikit.ProbeTerminal`)** |
| **`NO_COLOR` & TTY Awareness** | ✗ | ✗ | **✓ (Built-in)** |
| **Tables, Boxes, Progress Bars** | ✗ | ✗ (needs lipgloss) | **✓ (`tuikit`)** |

## Subpackages Overview

### 1. [`argkit`](argkit) — POSIX Argument & Flag Engine
- **Interspersed Arguments**: Mix options and positional arguments freely (`app -v src/ -o=bin dst/`).
- **Short Flag Clumping**: `-la` automatically unpacks into `-l` and `-a`.
- **Attached Values**: `-I*.tmp` maps directly to `-I=*.tmp`.
- **POSIX Terminator**: `--` treats all subsequent tokens as raw positional arguments.
- **Fuzzy Suggestions**: Suggests corrections for mistyped flags using stack-allocated Levenshtein distance (`"flag provided but not defined: --timout (did you mean: --timeout?)"`).
- **Shell Normalization**: Resolves argument splitting bugs caused by PowerShell and CMD shell tokenizers (`argkit.NormalizeArgs`).

### 2. [`tuikit`](tuikit) — Terminal UI & CLI App Router
- **Subcommand App Router (`tuikit.App`)**: Handles command routing, aliases, grouped help listings, and default fallback commands.
- **Terminal Prober (`tuikit.ProbeTerminal`)**: Sniffs TTY interactive status and enables Virtual Terminal escape processing on Windows 10/11.
- **Auto-Aligned Tables (`tuikit.Table`)**: Pixel-perfect column formatting with left, right, and center alignments.
- **Bordered Boxes (`tuikit.Box`)**: Clean bordered card containers (`BorderSingle`, `BorderRounded`, `BorderHeavy`) with guaranteed straight right borders.
- **Live Progress & Throughput (`tuikit.Progress`)**: Real-time progress bars with throughput rate (MB/s or items/s) and ETA calculations. Gracefully falls back to periodic summary logs in CI/CD or piped output.
- **Sparklines & Trend Graphs (`tuikit.RenderSparkline`)**: Renders inline trend graphs from floating-point sequences (` ▂▃▅▆▇█`).
- **Badges & Pipeline Steps**: Universal status tags (`BadgePass`, `BadgeFail`, `BadgeWarn`, `BadgeActive`) and numbered pipeline stages (`RenderStep`).

### 3. [`mcpkit`](mcpkit) — Model Context Protocol v1.0 Router & Fluent Builders
- **JSON-RPC 2.0 STDIO & HTTP/SSE Transports**: Full MCP `2024-11-05` server with 16 MiB scanner buffers, Bearer auth, TLS termination, and token-bucket rate limiting.
- **Fluent Schema Builder (`mcpkit.NewTool`)**: Construct JSON Schema tool definitions with functional `PropOption` modifiers (`mcpkit.Required()`, `mcpkit.Min()`, `mcpkit.Max()`, `mcpkit.MinItems()`, `mcpkit.Default()`) and zero nested `map[string]any` boilerplate.
- **Typed & Text Handlers**: Generic `TypedHandler[T]`, `LenientHandler[T]`, `TextHandler[T]`, `StrictTextHandler[T]`, and `NoArgsHandler` adapters.

### 4. [`uiakit`](uiakit) — Windows COM UI Automation (`CUIAutomation8`)
- **Zero-Allocation COM Traversal**: Direct `IUIAutomation2` vtable invocation powered by `foundation/silicon/comkit` (`Call0`–`Call5`) with `IUIAutomationCacheRequest` bulk property prefetching.
- **Set-of-Marks & DOM Extraction**: Enumerates Win32, WPF, WinUI 3, Qt, and Chromium/Electron `RootWebArea` accessibility trees into compact Markdown digests.
- **Action Patterns**: Native `IUIAutomationInvokePattern`, `ValuePattern`, `TogglePattern`, and `SelectionItemPattern` execution without cursor theft.

## Performance & Zero-Allocation Guarantees

Tested under parallel CPU load:
- **`VisibleWidth`**: Monospaced width calculation runs without heap allocations.
- **`Levenshtein` & Typo Suggest**: Uses stack-allocated arrays for strings up to 64 runes, guaranteeing **0 allocs/op** during interactive CLI flag validation.
- **`LinesSeq`**: Go 1.27 push iterator over string lines without intermediate slice allocations.

## Contributing & Development

```bash
# Run unit tests
make test

# Run tests with race detector
make race

# Run code linters
make lint

# Run benchmark suite
make bench
```

## 📄 License

This project is licensed under the **BSD 3-Clause License**. See the [LICENSE](LICENSE) file for details.
