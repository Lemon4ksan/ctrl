// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tuikit

import (
	"fmt"
	"strconv"
	"time"
)

// FormatBytes formats byte counts into human-readable strings (e.g. "4.2 MB", "512 KB").
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatNumber(n int64) string {
	if n < 0 {
		return strconv.FormatInt(n, 10)
	}
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}

	res := make([]byte, 0, len(s)+(len(s)-1)/3)
	rem := len(s) % 3
	if rem > 0 {
		res = append(res, s[:rem]...)
	}
	for i := rem; i < len(s); i += 3 {
		if len(res) > 0 {
			res = append(res, ',')
		}
		res = append(res, s[i:i+3]...)
	}
	return string(res)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	s := int(d.Seconds())
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	m := s / 60
	s %= 60
	if m < 60 {
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	h := m / 60
	m %= 60
	return fmt.Sprintf("%dh %02dm", h, m)
}
