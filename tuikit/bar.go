// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tuikit

import (
	"math"
	"strings"
)

// RenderBar generates a proportional ASCII/Unicode filled progress bar (e.g. ████████░░░░░░░░).
func RenderBar(ratio float64, width int) string {
	return DefaultStyler().RenderBar(ratio, width)
}

// RenderBar generates a proportional ASCII/Unicode filled progress bar using s's color policy.
func (s Styler) RenderBar(ratio float64, width int) string {
	if width <= 0 {
		return ""
	}

	if ratio < 0 {
		ratio = 0
	} else if ratio > 1.0 {
		ratio = 1.0
	}

	filled := min(int(math.Round(ratio*float64(width))), width)
	empty := width - filled

	return s.Cyan(strings.Repeat("█", filled)) + s.Gray(strings.Repeat("░", empty))
}

var sparklineGlyphs = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderSparkline generates an inline visual trend graph from a sequence of floating-point values.
func RenderSparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	minVal := values[0]
	maxVal := values[0]

	for _, v := range values {
		if v < minVal {
			minVal = v
		}

		if v > maxVal {
			maxVal = v
		}
	}

	span := maxVal - minVal

	var sb strings.Builder

	for _, v := range values {
		if span == 0 {
			sb.WriteRune(sparklineGlyphs[len(sparklineGlyphs)/2])
			continue
		}

		idx := int((v - minVal) / span * float64(len(sparklineGlyphs)-1))
		if idx < 0 {
			idx = 0
		} else if idx >= len(sparklineGlyphs) {
			idx = len(sparklineGlyphs) - 1
		}

		sb.WriteRune(sparklineGlyphs[idx])
	}

	return sb.String()
}
