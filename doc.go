// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ctrl provides a unified, zero-dependency CLI application framework and
// high-performance terminal UI toolkit for Go.
//
// It consolidates POSIX argument and flag parsing ([argkit]), subcommand routing,
// terminal capability sniffing, and terminal UI styling primitives ([tuikit]) into
// a single, cohesive architecture.
//
// # Architectural Foundations
//
//   - [github.com/lemon4ksan/ctrl/argkit] - Zero-dependency POSIX argument parser,
//     flag clumper (-la), attached value parser (-I*.tmp), and fuzzy typo suggester.
//   - [github.com/lemon4ksan/ctrl/tuikit] - Zero-dependency terminal UI toolkit,
//     providing CLI application management ([tuikit.App]), capability detection ([tuikit.ProbeTerminal]),
//     tables ([tuikit.Table]), boxes ([tuikit.Box]), progress bars ([tuikit.Progress]),
//     sparklines ([tuikit.RenderSparkline]), and status badges.
//
// # Basic Usage
//
//	package main
//
//	import (
//		"context"
//		"fmt"
//		"io"
//		"os"
//
//		"github.com/lemon4ksan/ctrl"
//		"github.com/lemon4ksan/ctrl/tuikit"
//	)
//
//	func main() {
//		app := ctrl.NewApp("mycli", ctrl.Version, "High-performance CLI utility",
//			ctrl.NewCommand("hello", "Print a greeting", "mycli hello [name]",
//				func(ctx context.Context, args []string, stdout, stderr io.Writer) error {
//					name := "World"
//					if len(args) > 0 {
//						name = args[0]
//					}
//					fmt.Fprintf(stdout, "%s Hello, %s!\n", tuikit.BadgePass(), name)
//					return nil
//				},
//			),
//		)
//
//		if err := app.Run(context.Background(), os.Args[1:]); err != nil {
//			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
//			os.Exit(1)
//		}
//	}
package ctrl
