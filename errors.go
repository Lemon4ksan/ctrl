// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctrl

import (
	"errors"
)

// Standard error definitions for the ctrl framework.
var (
	// ErrUnknownCommand is returned when an unrecognized command name is invoked.
	ErrUnknownCommand = errors.New("ctrl: unknown command")

	// ErrMissingArgument is returned when required positional arguments are absent.
	ErrMissingArgument = errors.New("ctrl: missing required argument")

	// ErrFlagParsing is returned when flag decoding fails.
	ErrFlagParsing = errors.New("ctrl: flag parsing error")
)
