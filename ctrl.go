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

// HelpRenderer is an optional interface implemented by commands with nested subcommands or custom help layouts.
type HelpRenderer = tuikit.HelpRenderer

// FlagProvider is an optional interface implemented by commands that expose their [flag.FlagSet] for help generation.
type FlagProvider = tuikit.FlagProvider

// Renderer represents any terminal UI widget capable of rendering to an [io.Writer] and producing a string.
type Renderer = tuikit.Renderer

// Styler provides isolated ANSI styling operations without relying on mutable global state.
type Styler = tuikit.Styler

// FuncCommand is a lightweight adapter creating a [Command] from functional closures.
type FuncCommand = tuikit.FuncCommand

// CommandGroup organizes subcommands under logical sections in help listings.
type CommandGroup = tuikit.CommandGroup

// HelpSection represents a custom documentation block rendered in the help screen.
type HelpSection = tuikit.HelpSection

// GroupCommand routes declarative 2nd-level subcommands with automatic help and validation.
type GroupCommand = tuikit.GroupCommand

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

// NewGroupCommand constructs a declarative 2nd-level subcommand group router.
func NewGroupCommand(name, synopsis string, subcommands ...Command) *GroupCommand {
	return tuikit.NewGroupCommand(name, synopsis, subcommands...)
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

// FormatFlags renders a [flag.FlagSet] in the unified ctrl HelpSection style with aligned
// "--flag <type>" columns, merged short/long aliases ("-f, --flag <type>"), and default values.
func FormatFlags(fs *flag.FlagSet) string {
	return argkit.FormatFlags(fs)
}

// PrintFlags writes the formatted flag table to w if fs has any registered flags.
func PrintFlags(w io.Writer, fs *flag.FlagSet) {
	argkit.PrintFlags(w, fs)
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

// NewStyler constructs an isolated [Styler] with a fixed color enablement policy.
func NewStyler(enabled bool) Styler {
	return tuikit.NewStyler(enabled)
}

// DefaultStyler returns a [Styler] that tracks the global color enablement setting.
func DefaultStyler() Styler {
	return tuikit.DefaultStyler()
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

	Var         = argkit.Var
	StringVar   = argkit.StringVar
	BoolVar     = argkit.BoolVar
	IntVar      = argkit.IntVar
	Int64Var    = argkit.Int64Var
	UintVar     = argkit.UintVar
	Uint64Var   = argkit.Uint64Var
	Float64Var  = argkit.Float64Var
	DurationVar = argkit.DurationVar
)
