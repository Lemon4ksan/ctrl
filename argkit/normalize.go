// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package argkit

import "strings"

// NormalizeArgs stitches back arguments that were fragmented by shell tokenizers
// (e.g. PowerShell or CMD splitting -out=pkg/api and .go).
func NormalizeArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}

	result := make([]string, 0, len(args))

	var arg strings.Builder
	for i := 0; i < len(args); i++ {
		arg.WriteString(args[i])
		for i+1 < len(args) && isOrphanedSuffix(args[i+1]) {
			arg.WriteString(args[i+1])
			i++
		}

		result = append(result, arg.String())
		arg.Reset()
	}

	return result
}

func isOrphanedSuffix(arg string) bool {
	return strings.HasPrefix(arg, ".") &&
		arg != "." &&
		arg != ".." &&
		!strings.Contains(arg, "/") &&
		!strings.Contains(arg, "\\")
}
