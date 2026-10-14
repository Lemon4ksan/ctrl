// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctrl

import (
	"context"
	"flag"
	"io"

	"github.com/lemon4ksan/ctrl/argkit"
	"github.com/lemon4ksan/ctrl/tuikit"
)

// App manages the lifecycle, subcommand routing, help formatting, and execution of CLI tools.
type App = tuikit.App

// Command represents an isolated, testable CLI subcommand contract.
type Command = tuikit.Command

// FuncCommand is a lightweight adapter creating a [Command] from functional closures.
type FuncCommand = tuikit.FuncCommand

// CommandGroup organizes subcommands under logical sections in help listings.
type CommandGroup = tuikit.CommandGroup

// Table manages columnar data formatting with guaranteed alignment.
type Table = tuikit.Table

// Box formats enclosed text cards and panels with borders.
type Box = tuikit.Box

// Progress manages live terminal feedback, throughput, ETA calculations, and progress rendering.
type Progress = tuikit.Progress

// NewApp constructs a new CLI application instance.
func NewApp(name, version, description string, commands ...Command) *App {
	return tuikit.NewApp(name, version, description, commands...)
}

// NewCommand creates a functional subcommand adapter.
func NewCommand(name, synopsis, usage string, run func(ctx context.Context, args []string, stdout, stderr io.Writer) error, aliases ...string) Command {
	return &tuikit.FuncCommand{
		CmdName:     name,
		CmdAliases:  aliases,
		CmdSynopsis: synopsis,
		CmdUsage:    usage,
		CmdRun:      run,
	}
}

// ParseInterspersedFlags parses a [flag.FlagSet] correctly with full POSIX semantics:
// clumping (-la), attached values (-I*.tmp), interspersed arguments, and typo suggestions.
func ParseInterspersedFlags(fs *flag.FlagSet, args []string) ([]string, error) {
	return argkit.ParseInterspersedFlags(fs, args)
}

// NewTable constructs a new Table with the specified header titles.
func NewTable(headers ...string) *Table {
	return tuikit.NewTable(headers...)
}

// NewBox creates a new rectangular box card with the given title and inner width constraint.
func NewBox(title string, innerWidth int) *Box {
	return tuikit.NewBox(title, innerWidth)
}

// NewProgress creates and initializes a new Progress widget.
func NewProgress(opts ...tuikit.ProgressOption) *Progress {
	return tuikit.NewProgress(opts...)
}
