// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package argkit

import (
	"flag"
	"strings"
)

// Suggest finds the flag name in fs closest in Levenshtein distance to name.
func Suggest(fs *flag.FlagSet, name string) string {
	return findClosestFlag(fs, name)
}

// findClosestFlag calculates Levenshtein distance to suggest the closest defined flag for typos.
func findClosestFlag(fs *flag.FlagSet, unknown string) string {
	if fs == nil {
		return ""
	}
	clean, _, _ := strings.Cut(strings.TrimLeft(unknown, "-"), "=")
	if len(clean) == 0 {
		return ""
	}

	bestMatch := ""
	minDist := 3 // max allowed edit distance threshold
	cleanLower := strings.ToLower(clean)

	fs.VisitAll(func(fl *flag.Flag) {
		dist := Levenshtein(cleanLower, strings.ToLower(fl.Name))
		if dist < minDist {
			minDist = dist
			bestMatch = fl.Name
		}
	})

	return bestMatch
}

// Levenshtein computes the Levenshtein distance between s1 and s2.
// It uses stack-allocated buffers for typical strings up to 64 runes to achieve
// 0 heap allocations.
func Levenshtein(s1, s2 string) int {
	if isASCII(s1) && isASCII(s2) {
		n1, n2 := len(s1), len(s2)
		if n1 == 0 {
			return n2
		}
		if n2 == 0 {
			return n1
		}

		if n1 < n2 {
			s1, s2 = s2, s1
			n1, n2 = n2, n1
		}

		var stackRow [65]int
		var row []int
		if n2+1 <= len(stackRow) {
			row = stackRow[:n2+1]
		} else {
			row = make([]int, n2+1)
		}

		for j := 0; j <= n2; j++ {
			row[j] = j
		}

		for i := 1; i <= n1; i++ {
			prev := i
			b1 := s1[i-1]
			for j := 1; j <= n2; j++ {
				cost := 0
				if b1 != s2[j-1] {
					cost = 1
				}
				cur := min(row[j-1]+cost, min(prev+1, row[j]+1))
				row[j-1] = prev
				prev = cur
			}
			row[n2] = prev
		}
		return row[n2]
	}

	var stackR1 [64]rune
	var stackR2 [64]rune
	r1 := decodeRunes(s1, stackR1[:0])
	r2 := decodeRunes(s2, stackR2[:0])

	n1, n2 := len(r1), len(r2)
	if n1 == 0 {
		return n2
	}
	if n2 == 0 {
		return n1
	}

	if n1 < n2 {
		r1, r2 = r2, r1
		n1, n2 = n2, n1
	}

	var stackRow [65]int
	var row []int
	if n2+1 <= len(stackRow) {
		row = stackRow[:n2+1]
	} else {
		row = make([]int, n2+1)
	}

	for j := 0; j <= n2; j++ {
		row[j] = j
	}

	for i := 1; i <= n1; i++ {
		prev := i
		r1Val := r1[i-1]
		for j := 1; j <= n2; j++ {
			cost := 0
			if r1Val != r2[j-1] {
				cost = 1
			}
			cur := min(row[j-1]+cost, min(prev+1, row[j]+1))
			row[j-1] = prev
			prev = cur
		}
		row[n2] = prev
	}
	return row[n2]
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

func decodeRunes(s string, buf []rune) []rune {
	for _, r := range s {
		if len(buf) < cap(buf) {
			buf = append(buf, r)
		} else {
			return []rune(s)
		}
	}
	return buf
}
