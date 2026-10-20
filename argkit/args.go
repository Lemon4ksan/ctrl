// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package argkit

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"unicode/utf8"
)

type boolFlag interface {
	IsBoolFlag() bool
}

const unknownFlagPrefix = "flag provided but not defined: "

// ParseInterspersedFlags parses a [flag.FlagSet] correctly with full POSIX semantics:
// - Free interspersing of positional arguments and flags anywhere in the command line.
// - Short flag clumping / stacking (e.g. `-la` -> `-l -a`, `-lAF` -> `-l -A -F`).
// - Attached values for short flags (e.g. `-I*.tmp` -> `-I=*.tmp`).
// - Strict POSIX terminator `--` (all subsequent arguments treated as literal positionals).
// - Fuzzy suggestions ("Did you mean: --flag?") for mistyped options.
func ParseInterspersedFlags(fs *flag.FlagSet, args []string) ([]string, error) {
	args = decomposeFlags(fs, NormalizeArgs(args))

	var (
		flagArgs = make([]string, 0, len(args))
		posArgs  []string
	)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			posArgs = append(posArgs, args[i+1:]...)
			break
		}

		if !strings.HasPrefix(arg, "-") || arg == "-" {
			posArgs = append(posArgs, arg)
			continue
		}

		cleanArg := strings.TrimLeft(arg, "-")
		flagName, _, hasEqual := strings.Cut(cleanArg, "=")

		fl := fs.Lookup(flagName)
		if fl == nil || hasEqual {
			flagArgs = append(flagArgs, arg)
			continue
		}

		bf, ok := fl.Value.(boolFlag)
		flagArgs = append(flagArgs, arg)

		if !ok || !bf.IsBoolFlag() {
			if i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		}
	}

	prevUsage := fs.Usage
	fs.Usage = func() {}
	defer func() { fs.Usage = prevUsage }()

	if err := fs.Parse(flagArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			PrintFlags(fs.Output(), fs)
			return nil, err
		}
		return nil, enrichParseError(fs, err)
	}

	return posArgs, nil
}

func enrichParseError(fs *flag.FlagSet, err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if _, after, ok := strings.Cut(msg, unknownFlagPrefix); ok {
		badFlag := strings.TrimSpace(after)
		if match := findClosestFlag(fs, badFlag); match != "" {
			return fmt.Errorf("%s%s (did you mean: --%s?)", unknownFlagPrefix, badFlag, match)
		}
	}
	return err
}

func decomposeFlags(fs *flag.FlagSet, args []string) []string {
	expanded := make([]string, 0, len(args))
	hitDoubleDash := false

	for _, arg := range args {
		if hitDoubleDash {
			expanded = append(expanded, arg)
			continue
		}
		if arg == "--" {
			hitDoubleDash = true
			expanded = append(expanded, arg)
			continue
		}

		if clumped, ok := decomposeFlag(fs, arg); ok {
			expanded = append(expanded, clumped...)
		} else {
			expanded = append(expanded, arg)
		}
	}

	return expanded
}

// decomposeFlag unpacks POSIX short flag clumps (e.g. -la -> [-l, -a]) and attached values (-I*.tmp -> [-I=*.tmp]).
func decomposeFlag(fs *flag.FlagSet, arg string) ([]string, bool) {
	if !strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") || arg == "-" {
		return nil, false
	}

	clean := arg[1:]
	if len(clean) <= 1 {
		return nil, false
	}

	// If the entire string is already registered (e.g. -where, -all, -sort), keep as is
	flagName, _, _ := strings.Cut(clean, "=")
	if fs.Lookup(flagName) != nil {
		return nil, false
	}

	var result []string
	curr := clean

	for len(curr) > 0 {
		r, sz := utf8.DecodeRuneInString(curr)
		if r == utf8.RuneError && sz == 1 {
			return nil, false
		}
		firstChar := curr[:sz]
		fl := fs.Lookup(firstChar)
		if fl == nil {
			// First character is not a known flag, cannot decompose
			return nil, false
		}

		if bf, ok := fl.Value.(boolFlag); ok && bf.IsBoolFlag() {
			result = append(result, "-"+firstChar)
			curr = curr[sz:]
			if strings.HasPrefix(curr, "=") {
				return nil, false
			}
		} else {
			// Non-boolean flag: rest of current string is an attached value
			rest := strings.TrimPrefix(curr[sz:], "=")
			result = append(result, fmt.Sprintf("-%s=%s", firstChar, rest))
			curr = ""
		}
	}

	return result, true
}
