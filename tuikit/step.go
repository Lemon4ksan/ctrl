// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tuikit

import (
	"fmt"
	"strings"
)

// RenderStep formats a numbered pipeline stage with dotted connector dots.
// Example: "[1/3] Pre-flight Contract Audit ........... ✔ PASSED (100% clean)"
func RenderStep(current, total int, label, status string, totalLabelWidth int) string {
	return DefaultStyler().RenderStep(current, total, label, status, totalLabelWidth)
}

// RenderStep formats a numbered pipeline stage with dotted connector dots using s's color policy.
func (s Styler) RenderStep(current, total int, label, status string, totalLabelWidth int) string {
	prefix := fmt.Sprintf("[%d/%d] %s ", current, total, label)
	vis := VisibleWidth(prefix)

	if totalLabelWidth <= 0 {
		totalLabelWidth = 44
	}

	dots := max(totalLabelWidth-vis, 3)

	return prefix + s.Gray(strings.Repeat(".", dots)) + " " + status
}

// RenderDivider returns a horizontal divider line of the given width.
func RenderDivider(width int) string {
	return DefaultStyler().RenderDivider(width)
}

// RenderDivider returns a horizontal divider line of the given width using s's color policy.
func (s Styler) RenderDivider(width int) string {
	if width <= 0 {
		width = 77
	}

	return s.Gray(strings.Repeat("─", width))
}

// RenderHeader formats an emphasized section header.
func RenderHeader(title string) string {
	return DefaultStyler().RenderHeader(title)
}

// RenderHeader formats an emphasized section header using s's color policy.
func (s Styler) RenderHeader(title string) string {
	return s.Bold(title)
}
