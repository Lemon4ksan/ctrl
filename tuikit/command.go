// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tuikit

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// Command represents an isolated, testable CLI subcommand contract.
type Command interface {
	// Name returns the primary subcommand invocation name (e.g. "ls", "scan", "pack").
	Name() string
	// Aliases returns alternative command names (e.g. "dir" or "list" for "ls").
	Aliases() []string
	// Synopsis returns a one-line summary for help listings.
	Synopsis() string
	// Usage returns detailed syntax instructions.
	Usage() string
	// Run executes the command logic with isolated arguments and output streams.
	Run(ctx context.Context, args []string, stdout, stderr io.Writer) error
}

// HelpRenderer is an optional interface implemented by commands with nested subcommands
// or custom help layouts.
type HelpRenderer interface {
	RenderHelp(ctx context.Context, args []string, w io.Writer) error
}

// FuncCommand is a lightweight adapter creating a [Command] from functional closures,
// optionally supporting declarative 2nd-level Subcommands.
type FuncCommand struct {
	CmdName        string
	CmdAliases     []string
	CmdSynopsis    string
	CmdUsage       string
	CmdDefaultSub  string
	CmdSubcommands []Command
	CmdRun         func(ctx context.Context, args []string, stdout, stderr io.Writer) error
}

func (c *FuncCommand) Name() string      { return c.CmdName }
func (c *FuncCommand) Aliases() []string { return c.CmdAliases }
func (c *FuncCommand) Synopsis() string  { return c.CmdSynopsis }
func (c *FuncCommand) Usage() string     { return c.CmdUsage }
func (c *FuncCommand) Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(c.CmdSubcommands) > 0 {
		grp := NewGroupCommand(c.CmdName, c.CmdSynopsis, c.CmdSubcommands...).
			WithUsage(c.CmdUsage).
			WithDefault(c.CmdDefaultSub).
			WithAliases(c.CmdAliases...)
		return grp.Run(ctx, args, stdout, stderr)
	}
	if c.CmdRun != nil {
		return c.CmdRun(ctx, args, stdout, stderr)
	}
	return nil
}

// RenderHelp renders help for a FuncCommand, delegating to GroupCommand if CmdSubcommands are defined.
func (c *FuncCommand) RenderHelp(ctx context.Context, args []string, w io.Writer) error {
	if len(c.CmdSubcommands) > 0 {
		grp := NewGroupCommand(c.CmdName, c.CmdSynopsis, c.CmdSubcommands...).
			WithUsage(c.CmdUsage).
			WithDefault(c.CmdDefaultSub).
			WithAliases(c.CmdAliases...)
		return grp.RenderHelp(ctx, args, w)
	}
	return renderSingleCommandHelp(ctx, "", c, w)
}

// GroupCommand routes declarative 2nd-level subcommands (e.g., "stss net [status|enable|disable]",
// "stss hook [pre-tool|session-start]", "stss tool [list|add]") with automatic unknown-subcommand
// validation and unified "--help" generation.
type GroupCommand struct {
	CmdName       string
	CmdAliases    []string
	CmdSynopsis   string
	CmdUsage      string
	DefaultSubcmd string
	Subcommands   []Command
	subMap        map[string]Command
}

// HubCommand is an alias for [GroupCommand].
type HubCommand = GroupCommand

// NewGroupCommand constructs a declarative 2nd-level subcommand group router.
func NewGroupCommand(name, synopsis string, subcommands ...Command) *GroupCommand {
	g := &GroupCommand{
		CmdName:     name,
		CmdSynopsis: synopsis,
		Subcommands: subcommands,
	}
	g.initMap()
	return g
}

// NewHubCommand creates a nested subcommand router (alias for [NewGroupCommand] with usage and defaultSubcmd).
func NewHubCommand(name, synopsis, usage, defaultSubcmd string, subcommands ...Command) *GroupCommand {
	return NewGroupCommand(name, synopsis, subcommands...).
		WithUsage(usage).
		WithDefault(defaultSubcmd)
}

// WithUsage sets a custom usage line on the GroupCommand.
func (g *GroupCommand) WithUsage(usage string) *GroupCommand {
	g.CmdUsage = usage
	return g
}

// WithDefault sets the default subcommand executed when no subcommand name is provided.
func (g *GroupCommand) WithDefault(defaultSubcmd string) *GroupCommand {
	g.DefaultSubcmd = defaultSubcmd
	return g
}

// WithAliases attaches command aliases to the GroupCommand.
func (g *GroupCommand) WithAliases(aliases ...string) *GroupCommand {
	g.CmdAliases = aliases
	return g
}

func (g *GroupCommand) initMap() {
	if g.subMap != nil {
		return
	}
	g.subMap = make(map[string]Command, len(g.Subcommands)*2)
	for _, sc := range g.Subcommands {
		g.subMap[strings.ToLower(sc.Name())] = sc
		for _, alias := range sc.Aliases() {
			g.subMap[strings.ToLower(alias)] = sc
		}
	}
}

func (g *GroupCommand) subcommandNames() string {
	names := make([]string, 0, len(g.Subcommands))
	for _, sc := range g.Subcommands {
		names = append(names, sc.Name())
	}
	return strings.Join(names, ", ")
}

func (g *GroupCommand) Name() string      { return g.CmdName }
func (g *GroupCommand) Aliases() []string { return g.CmdAliases }
func (g *GroupCommand) Synopsis() string  { return g.CmdSynopsis }
func (g *GroupCommand) Usage() string {
	if g.CmdUsage != "" {
		return g.CmdUsage
	}
	return fmt.Sprintf("%s <%s> [flags]", g.CmdName, strings.Join(strings.Split(g.subcommandNames(), ", "), "|"))
}

// RenderHelp renders either the group overview (with aligned Subcommands table) or the specific
// 2nd-level subcommand's help and aligned FlagSet.
func (g *GroupCommand) RenderHelp(ctx context.Context, args []string, w io.Writer) error {
	g.initMap()

	for _, arg := range args {
		if arg == "--" {
			break
		}
		if !strings.HasPrefix(arg, "-") {
			if sc, ok := g.subMap[strings.ToLower(arg)]; ok {
				return renderSingleCommandHelp(ctx, g.CmdName, sc, w)
			}
		}
	}

	fmt.Fprintf(w, "%s — %s\n\nUsage:\n  %s\n\nSubcommands:\n", g.Name(), g.Synopsis(), g.Usage())
	maxLen := 14
	for _, sc := range g.Subcommands {
		d := sc.Name()
		if len(sc.Aliases()) > 0 {
			d = fmt.Sprintf("%s, %s", sc.Name(), strings.Join(sc.Aliases(), ", "))
		}
		if len(d) > maxLen && len(d) <= 30 {
			maxLen = len(d)
		}
	}
	for _, sc := range g.Subcommands {
		d := sc.Name()
		if len(sc.Aliases()) > 0 {
			d = fmt.Sprintf("%s, %s", sc.Name(), strings.Join(sc.Aliases(), ", "))
		}
		syn := sc.Synopsis()
		if strings.EqualFold(sc.Name(), g.DefaultSubcmd) {
			syn += " (default)"
		}
		fmt.Fprintf(w, "  %-*s  %s\n", maxLen, d, syn)
	}

	if g.DefaultSubcmd != "" {
		if sc, ok := g.subMap[strings.ToLower(g.DefaultSubcmd)]; ok {
			var flagBuf bytes.Buffer
			if err := sc.Run(
				ctx,
				[]string{"-h"},
				io.Discard,
				&flagBuf,
			); errors.Is(err, flag.ErrHelp) &&
				flagBuf.Len() > 0 {
				fmt.Fprintf(w, "\n%s", flagBuf.String())
			}
		}
	}
	return nil
}

// Run dispatches execution to the matching nested subcommand or the default subcommand.
func (g *GroupCommand) Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	g.initMap()

	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "-h" || arg == "--help" || arg == "-help" {
			return g.RenderHelp(ctx, args, stdout)
		}
	}

	// Locate the first non-flag token before "--" that matches a registered subcommand
	subIdx := -1
	for i, arg := range args {
		if arg == "--" {
			break
		}
		if !strings.HasPrefix(arg, "-") {
			if _, ok := g.subMap[strings.ToLower(arg)]; ok {
				subIdx = i
				break
			}
		}
	}

	if subIdx >= 0 {
		subName := strings.ToLower(args[subIdx])
		sc := g.subMap[subName]
		subArgs := make([]string, 0, len(args)-1)
		subArgs = append(subArgs, args[:subIdx]...)
		subArgs = append(subArgs, args[subIdx+1:]...)
		err := sc.Run(ctx, subArgs, stdout, stderr)
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return fmt.Errorf(
			"unknown %s subcommand %q (expected one of: %s)",
			g.CmdName,
			args[0],
			g.subcommandNames(),
		)
	}

	if g.DefaultSubcmd != "" {
		if sc, ok := g.subMap[strings.ToLower(g.DefaultSubcmd)]; ok {
			err := sc.Run(ctx, args, stdout, stderr)
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
	}

	_ = g.RenderHelp(ctx, args, stdout)
	return flag.ErrHelp
}

func renderSingleCommandHelp(ctx context.Context, parentName string, cmd Command, w io.Writer) error {
	fullName := cmd.Name()
	if parentName != "" {
		fullName = parentName + " " + cmd.Name()
	}
	usage := cmd.Usage()
	if usage == "" {
		usage = fullName + " [flags]"
	}
	fmt.Fprintf(w, "%s — %s\n\nUsage:\n  %s\n", fullName, cmd.Synopsis(), usage)

	var flagBuf bytes.Buffer
	if err := cmd.Run(ctx, []string{"-h"}, io.Discard, &flagBuf); errors.Is(err, flag.ErrHelp) && flagBuf.Len() > 0 {
		fmt.Fprintf(w, "\n%s", flagBuf.String())
	}
	return nil
}
