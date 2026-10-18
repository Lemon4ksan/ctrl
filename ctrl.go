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

// HelpSection represents a custom documentation block rendered in the help screen.
type HelpSection = tuikit.HelpSection

// HubCommand routes nested subcommands with an optional default subcommand fallback.
type HubCommand = tuikit.HubCommand

// StringSliceFlag implements [flag.Value] for multi-valued command line arguments.
type StringSliceFlag = argkit.StringSliceFlag

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
func NewCommand(
	name, synopsis, usage string,
	run func(ctx context.Context, args []string, stdout, stderr io.Writer) error,
	aliases ...string,
) Command {
	return &tuikit.FuncCommand{
		CmdName:     name,
		CmdAliases:  aliases,
		CmdSynopsis: synopsis,
		CmdUsage:    usage,
		CmdRun:      run,
	}
}

// NewHubCommand creates a nested subcommand router.
func NewHubCommand(name, synopsis, usage, defaultSubcmd string, subcommands ...Command) *HubCommand {
	return tuikit.NewHubCommand(name, synopsis, usage, defaultSubcmd, subcommands...)
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

// Styling, badge, and flag binding re-exports for rapid zero-dependency console formatting.
var (
	Bold      = tuikit.Bold
	Dim       = tuikit.Dim
	Italic    = tuikit.Italic
	Underline = tuikit.Underline
	Red       = tuikit.Red
	Green     = tuikit.Green
	Yellow    = tuikit.Yellow
	Blue      = tuikit.Blue
	Magenta   = tuikit.Magenta
	Cyan      = tuikit.Cyan
	Gray      = tuikit.Gray
	White     = tuikit.White

	BadgePass   = tuikit.BadgePass
	BadgePassed = tuikit.BadgePassed
	BadgeWarn   = tuikit.BadgeWarn
	BadgeFail   = tuikit.BadgeFail
	BadgeInfo   = tuikit.BadgeInfo
	BadgeActive = tuikit.BadgeActive
	BadgeSync   = tuikit.BadgeSync

	StripANSI    = tuikit.StripANSI
	VisibleWidth = tuikit.VisibleWidth
	RenderStep   = tuikit.RenderStep
	RenderBar    = tuikit.RenderBar

	StringVar   = argkit.StringVar
	BoolVar     = argkit.BoolVar
	IntVar      = argkit.IntVar
	Int64Var    = argkit.Int64Var
	Float64Var  = argkit.Float64Var
	DurationVar = argkit.DurationVar
)
