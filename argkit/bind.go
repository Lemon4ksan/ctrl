// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package argkit

import (
	"flag"
	"strings"
	"time"
)

// StringSliceFlag implements [flag.Value] for multi-valued command line arguments.
type StringSliceFlag []string

// String returns a comma-separated representation of the slice values.
func (s *StringSliceFlag) String() string {
	if s == nil {
		return ""
	}
	return strings.Join(*s, ",")
}

// Set appends value to the underlying string slice.
func (s *StringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// Var binds an arbitrary [flag.Value] implementation with an optional short alias.
func Var(fs *flag.FlagSet, value flag.Value, name, short, usage string) {
	fs.Var(value, name, usage)

	if short != "" && short != name {
		fs.Var(value, short, usage)
	}
}

// StringVar binds a string flag with optional short alias.
func StringVar(fs *flag.FlagSet, p *string, name, short, value, usage string) {
	fs.StringVar(p, name, value, usage)

	if short != "" && short != name {
		fs.StringVar(p, short, value, usage)
	}
}

// BoolVar binds a boolean flag with optional short alias.
func BoolVar(fs *flag.FlagSet, p *bool, name, short string, value bool, usage string) {
	fs.BoolVar(p, name, value, usage)

	if short != "" && short != name {
		fs.BoolVar(p, short, value, usage)
	}
}

// IntVar binds an integer flag with optional short alias.
func IntVar(fs *flag.FlagSet, p *int, name, short string, value int, usage string) {
	fs.IntVar(p, name, value, usage)

	if short != "" && short != name {
		fs.IntVar(p, short, value, usage)
	}
}

// Int64Var binds an int64 flag with optional short alias.
func Int64Var(fs *flag.FlagSet, p *int64, name, short string, value int64, usage string) {
	fs.Int64Var(p, name, value, usage)

	if short != "" && short != name {
		fs.Int64Var(p, short, value, usage)
	}
}

// UintVar binds a uint flag with optional short alias.
func UintVar(fs *flag.FlagSet, p *uint, name, short string, value uint, usage string) {
	fs.UintVar(p, name, value, usage)

	if short != "" && short != name {
		fs.UintVar(p, short, value, usage)
	}
}

// Uint64Var binds a uint64 flag with optional short alias.
func Uint64Var(fs *flag.FlagSet, p *uint64, name, short string, value uint64, usage string) {
	fs.Uint64Var(p, name, value, usage)

	if short != "" && short != name {
		fs.Uint64Var(p, short, value, usage)
	}
}

// Float64Var binds a float64 flag with optional short alias.
func Float64Var(fs *flag.FlagSet, p *float64, name, short string, value float64, usage string) {
	fs.Float64Var(p, name, value, usage)

	if short != "" && short != name {
		fs.Float64Var(p, short, value, usage)
	}
}

// DurationVar binds a time.Duration flag with optional short alias.
func DurationVar(fs *flag.FlagSet, p *time.Duration, name, short string, value time.Duration, usage string) {
	fs.DurationVar(p, name, value, usage)

	if short != "" && short != name {
		fs.DurationVar(p, short, value, usage)
	}
}
