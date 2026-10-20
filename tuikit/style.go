// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tuikit

import (
	"fmt"
	"io"
	"iter"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"unicode/utf8"
)

// Renderer represents any terminal UI widget capable of rendering to an [io.Writer]
// and producing a formatted string representation.
type Renderer interface {
	fmt.Stringer
	Render(w io.Writer) error
}

var (
	ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	noColor   atomic.Int32
)

func init() {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		noColor.Store(1)
	}
}

// SetColorEnabled toggles ANSI color formatting globally.
func SetColorEnabled(enabled bool) {
	if enabled {
		noColor.Store(0)
	} else {
		noColor.Store(1)
	}
}

// ColorEnabled reports whether ANSI color styling is active globally.
func ColorEnabled() bool {
	return noColor.Load() == 0
}

// Styler provides isolated ANSI styling operations without relying on mutable global state.
// A zero-value Styler defers to the global [ColorEnabled] state; an explicit Styler created
// via [NewStyler] enforces its own deterministic color policy.
type Styler struct {
	explicit bool
	enabled  bool
}

// NewStyler constructs an isolated [Styler] with a fixed color enablement policy.
func NewStyler(enabled bool) Styler {
	return Styler{explicit: true, enabled: enabled}
}

// DefaultStyler returns a [Styler] that tracks the global [ColorEnabled] setting.
func DefaultStyler() Styler {
	return Styler{}
}

// Enabled reports whether ANSI escape sequences are active for this Styler.
func (s Styler) Enabled() bool {
	if s.explicit {
		return s.enabled
	}
	return ColorEnabled()
}

func (s Styler) apply(code, text string) string {
	if !s.Enabled() || text == "" {
		return text
	}
	return code + text + ansiReset
}

// Bold returns text wrapped in ANSI bold using s's color policy.
func (s Styler) Bold(text string) string { return s.apply(ansiBold, text) }

// Dim returns text wrapped in ANSI dim using s's color policy.
func (s Styler) Dim(text string) string { return s.apply(ansiDim, text) }

// Italic returns text wrapped in ANSI italic using s's color policy.
func (s Styler) Italic(text string) string { return s.apply(ansiItalic, text) }

// Underline returns text wrapped in ANSI underline using s's color policy.
func (s Styler) Underline(text string) string { return s.apply(ansiUnder, text) }

// Red returns text colored in Red using s's color policy.
func (s Styler) Red(text string) string { return s.apply(ansiRed, text) }

// Green returns text colored in Green using s's color policy.
func (s Styler) Green(text string) string { return s.apply(ansiGreen, text) }

// Yellow returns text colored in Yellow using s's color policy.
func (s Styler) Yellow(text string) string { return s.apply(ansiYellow, text) }

// Blue returns text colored in Blue using s's color policy.
func (s Styler) Blue(text string) string { return s.apply(ansiBlue, text) }

// Magenta returns text colored in Magenta using s's color policy.
func (s Styler) Magenta(text string) string { return s.apply(ansiMagenta, text) }

// Cyan returns text colored in Cyan using s's color policy.
func (s Styler) Cyan(text string) string { return s.apply(ansiCyan, text) }

// Gray returns text colored in Gray using s's color policy.
func (s Styler) Gray(text string) string { return s.apply(ansiGray, text) }

// White returns text colored in Bright White using s's color policy.
func (s Styler) White(text string) string { return s.apply(ansiWhite, text) }

// StripANSI removes all ANSI escape sequences from the string.
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// VisibleWidth returns the visual character width of a string in a monospaced terminal,
// ignoring any ANSI color codes.
func VisibleWidth(s string) int {
	clean := StripANSI(s)
	return utf8.RuneCountInString(clean)
}

// ANSI Escape Codes
const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiItalic  = "\033[3m"
	ansiUnder   = "\033[4m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
	ansiGray    = "\033[90m"
	ansiWhite   = "\033[97m"
)

// Bold returns text wrapped in ANSI bold.
func Bold(text string) string { return DefaultStyler().Bold(text) }

// Dim returns text wrapped in ANSI dim.
func Dim(text string) string { return DefaultStyler().Dim(text) }

// Italic returns text wrapped in ANSI italic.
func Italic(text string) string { return DefaultStyler().Italic(text) }

// Underline returns text wrapped in ANSI underline.
func Underline(text string) string { return DefaultStyler().Underline(text) }

// Red returns text colored in Red.
func Red(text string) string { return DefaultStyler().Red(text) }

// Green returns text colored in Green.
func Green(text string) string { return DefaultStyler().Green(text) }

// Yellow returns text colored in Yellow.
func Yellow(text string) string { return DefaultStyler().Yellow(text) }

// Blue returns text colored in Blue.
func Blue(text string) string { return DefaultStyler().Blue(text) }

// Magenta returns text colored in Magenta.
func Magenta(text string) string { return DefaultStyler().Magenta(text) }

// Cyan returns text colored in Cyan.
func Cyan(text string) string { return DefaultStyler().Cyan(text) }

// Gray returns text colored in Gray.
func Gray(text string) string { return DefaultStyler().Gray(text) }

// White returns text colored in Bright White.
func White(text string) string { return DefaultStyler().White(text) }

// PadRight pads string with spaces to target visual width.
func PadRight(s string, width int) string {
	w := VisibleWidth(s)
	if w >= width {
		return s
	}

	return s + strings.Repeat(" ", width-w)
}

// PadLeft pads string on the left with spaces to target visual width.
func PadLeft(s string, width int) string {
	w := VisibleWidth(s)
	if w >= width {
		return s
	}

	return strings.Repeat(" ", width-w) + s
}

// PadCenter centers string within target visual width.
func PadCenter(s string, width int) string {
	w := VisibleWidth(s)
	if w >= width {
		return s
	}

	left := (width - w) / 2
	right := width - w - left

	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// LinesSeq returns an iterator over individual lines in text, supporting both LF (\n)
// and CRLF (\r\n) line terminators. Traversal stops immediately if yield returns false.
//
// Concurrency & Zero-Allocation Semantics:
// LinesSeq operates purely on string slices without heap allocations. It is safe for concurrent use.
func LinesSeq(text string) iter.Seq[string] {
	return func(yield func(string) bool) {
		rem := text
		for len(rem) > 0 {
			idx := strings.IndexByte(rem, '\n')
			var line string
			if idx == -1 {
				line = rem
				rem = ""
			} else {
				line = rem[:idx]
				rem = rem[idx+1:]
			}
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			if !yield(line) {
				return
			}
		}
	}
}
