// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tuikit

import "fmt"

// TaxStage represents a single phase in latency or throughput breakdown.
type TaxStage struct {
	Name     string
	Duration string
	Share    string
	Ratio    float64
}

// RenderTaxDecomposition formats a clean latency/share decomposition card.
func RenderTaxDecomposition(stages []TaxStage, barWidth int) string {
	if barWidth <= 0 {
		barWidth = 30
	}

	innerWidth := max(75, 45+barWidth)
	box := NewBox("", innerWidth)
	box.AddLine(fmt.Sprintf("%-16s %-14s %-10s %s", "Stage", "Duration", "Share", "Proportional Breakdown"))
	box.AddDivider()

	for _, st := range stages {
		bar := RenderBar(st.Ratio, barWidth)
		line := fmt.Sprintf("%-16s %-14s %-10s %s", st.Name, st.Duration, st.Share, bar)
		box.AddLine(line)
	}

	return box.String()
}
