// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package argkit

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

type formattedFlagItem struct {
	val      flag.Value
	short    string
	long     string
	typeName string
	desc     string
	defVal   string
	rawUsage string
	isBool   bool
}

// sameFlagValue safely compares two flag.Value instances for pointer/value identity
// without panicking if a custom implementation uses an incomparable underlying type.
func sameFlagValue(a, b flag.Value) (same bool) {
	if a == nil || b == nil {
		return false
	}
	defer func() {
		if recover() != nil {
			same = false
		}
	}()
	return a == b
}

// findMatchingFlagItem locates an existing formattedFlagItem that represents an alias
// of fl (either sharing the exact same underlying [flag.Value] pointer or matching non-empty usage and default value).
func findMatchingFlagItem(items []*formattedFlagItem, fl *flag.Flag) *formattedFlagItem {
	isShort := len(fl.Name) == 1

	// Priority 1: Exact same underlying flag.Value target (bound via StringVar, BoolVar, Var, etc.)
	for _, it := range items {
		if sameFlagValue(it.val, fl.Value) && it.rawUsage == fl.Usage && it.defVal == fl.DefValue {
			if isShort && it.short == "" {
				return it
			}
			if !isShort && it.long == "" {
				return it
			}
		}
	}

	// Priority 2: Fallback heuristic for flags with identical non-empty usage and default value
	if fl.Usage != "" {
		for _, it := range items {
			if it.rawUsage == fl.Usage && it.defVal == fl.DefValue {
				if isShort && it.short == "" {
					return it
				}
				if !isShort && it.long == "" {
					return it
				}
			}
		}
	}

	return nil
}

// FormatFlags renders a [flag.FlagSet] in the unified ctrl HelpSection style with aligned
// "--flag <type>" columns, merged short/long aliases ("-f, --flag <type>"), and default values.
func FormatFlags(fs *flag.FlagSet) string {
	if fs == nil {
		return ""
	}

	var raw []*flag.Flag
	fs.VisitAll(func(fl *flag.Flag) {
		raw = append(raw, fl)
	})
	if len(raw) == 0 {
		return ""
	}

	items := make([]*formattedFlagItem, 0, len(raw))

	for _, fl := range raw {
		bf, ok := fl.Value.(boolFlag)
		isB := ok && bf.IsBoolFlag()
		typeName, desc := flag.UnquoteUsage(fl)
		if !isB && typeName == "" {
			typeName = "value"
		}

		if existing := findMatchingFlagItem(items, fl); existing != nil {
			if len(fl.Name) == 1 {
				existing.short = fl.Name
			} else {
				existing.long = fl.Name
			}
			continue
		}

		item := &formattedFlagItem{
			val:      fl.Value,
			typeName: typeName,
			desc:     desc,
			defVal:   fl.DefValue,
			rawUsage: fl.Usage,
			isBool:   isB,
		}
		if len(fl.Name) == 1 {
			item.short = fl.Name
		} else {
			item.long = fl.Name
		}
		items = append(items, item)
	}

	type renderedRow struct {
		spec string
		desc string
	}
	rows := make([]renderedRow, 0, len(items))
	maxLen := 14

	for _, it := range items {
		var spec string
		switch {
		case it.short != "" && it.long != "":
			spec = fmt.Sprintf("-%s, --%s", it.short, it.long)
		case it.long != "":
			spec = "--" + it.long
		default:
			spec = "-" + it.short
		}
		if !it.isBool && it.typeName != "" {
			spec += " <" + it.typeName + ">"
		}
		if len(spec) > maxLen && len(spec) <= 34 {
			maxLen = len(spec)
		}

		desc := it.desc
		if it.defVal != "" && it.defVal != "false" && it.defVal != "0" && it.defVal != "0s" &&
			!strings.Contains(strings.ToLower(desc), "default") {
			desc = fmt.Sprintf("%s (default: %s)", desc, it.defVal)
		}
		rows = append(rows, renderedRow{spec: spec, desc: desc})
	}

	var b strings.Builder
	b.WriteString("Flags:\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "  %-*s  %s\n", maxLen, r.spec, r.desc)
	}
	return b.String()
}

// PrintFlags writes the formatted flag table to w if fs has any registered flags.
func PrintFlags(w io.Writer, fs *flag.FlagSet) {
	if w == nil || fs == nil {
		return
	}
	if out := FormatFlags(fs); out != "" {
		_, _ = io.WriteString(w, out)
	}
}
