// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tuikit

import (
	"context"
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

// FuncCommand is a lightweight adapter creating a [Command] from functional closures.
type FuncCommand struct {
	CmdName     string
	CmdAliases  []string
	CmdSynopsis string
	CmdUsage    string
	CmdRun      func(ctx context.Context, args []string, stdout, stderr io.Writer) error
}

func (c *FuncCommand) Name() string      { return c.CmdName }
func (c *FuncCommand) Aliases() []string { return c.CmdAliases }
func (c *FuncCommand) Synopsis() string  { return c.CmdSynopsis }
func (c *FuncCommand) Usage() string     { return c.CmdUsage }
func (c *FuncCommand) Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if c.CmdRun != nil {
		return c.CmdRun(ctx, args, stdout, stderr)
	}
	return nil
}

// HubCommand routes nested subcommands (e.g., "net [status|enable|disable]" or "tool [list|add]")
// with an optional default subcommand when invoked without a subcommand name.
type HubCommand struct {
	CmdName       string
	CmdAliases    []string
	CmdSynopsis   string
	CmdUsage      string
	DefaultSubcmd string
	Subcommands   []Command
	subMap        map[string]Command
}

// NewHubCommand creates a nested subcommand router.
func NewHubCommand(name, synopsis, usage, defaultSubcmd string, subcommands ...Command) *HubCommand {
	h := &HubCommand{
		CmdName:       name,
		CmdSynopsis:   synopsis,
		CmdUsage:      usage,
		DefaultSubcmd: defaultSubcmd,
		Subcommands:   subcommands,
		subMap:        make(map[string]Command, len(subcommands)*2),
	}
	for _, sc := range subcommands {
		h.subMap[sc.Name()] = sc
		for _, alias := range sc.Aliases() {
			h.subMap[alias] = sc
		}
	}
	return h
}

// WithAliases attaches command aliases to the HubCommand and returns it for chaining.
func (h *HubCommand) WithAliases(aliases ...string) *HubCommand {
	h.CmdAliases = aliases
	return h
}

func (h *HubCommand) Name() string      { return h.CmdName }
func (h *HubCommand) Aliases() []string { return h.CmdAliases }
func (h *HubCommand) Synopsis() string  { return h.CmdSynopsis }
func (h *HubCommand) Usage() string {
	if h.CmdUsage != "" {
		return h.CmdUsage
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s <subcommand> [flags]\n\nSubcommands:\n", h.CmdName)
	for _, sc := range h.Subcommands {
		fmt.Fprintf(&b, "  %-14s %s\n", sc.Name(), sc.Synopsis())
	}
	return strings.TrimRight(b.String(), "\n")
}

// Run dispatches execution to the matching nested subcommand or the default subcommand.
func (h *HubCommand) Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if h.subMap == nil {
		h.subMap = make(map[string]Command, len(h.Subcommands)*2)
		for _, sc := range h.Subcommands {
			h.subMap[sc.Name()] = sc
			for _, alias := range sc.Aliases() {
				h.subMap[alias] = sc
			}
		}
	}

	// Locate the first non-flag token before "--" as the subcommand candidate
	subIdx := -1
	for i, arg := range args {
		if arg == "--" {
			break
		}
		if !strings.HasPrefix(arg, "-") {
			if _, ok := h.subMap[strings.ToLower(arg)]; ok {
				subIdx = i
				break
			}
		}
	}

	if subIdx >= 0 {
		subName := strings.ToLower(args[subIdx])
		sc := h.subMap[subName]
		subArgs := make([]string, 0, len(args)-1)
		subArgs = append(subArgs, args[:subIdx]...)
		subArgs = append(subArgs, args[subIdx+1:]...)
		return sc.Run(ctx, subArgs, stdout, stderr)
	}

	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return fmt.Errorf("unknown %s subcommand %q (usage: %s)", h.CmdName, args[0], h.Usage())
	}

	if h.DefaultSubcmd != "" {
		if sc, ok := h.subMap[h.DefaultSubcmd]; ok {
			return sc.Run(ctx, args, stdout, stderr)
		}
	}

	fmt.Fprintf(stdout, "%s — %s\n\nUsage:\n  %s\n", h.Name(), h.Synopsis(), h.Usage())
	return flag.ErrHelp
}
