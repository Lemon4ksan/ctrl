// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mcpkit

import (
	"context"
	"encoding/json"
)

// TypedHandler adapts a strongly-typed Go function into a [ToolHandlerFunc].
// If rawArgs fails JSON unmarshaling, an [ErrorResult] is returned automatically.
func TypedHandler[T any](fn func(ctx context.Context, args T) CallToolResult) ToolHandlerFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) CallToolResult {
		var args T
		if len(rawArgs) > 0 {
			if err := json.Unmarshal(rawArgs, &args); err != nil {
				return ErrorResult(err)
			}
		}
		return fn(ctx, args)
	}
}

// LenientHandler adapts a strongly-typed Go function into a [ToolHandlerFunc] that
// tolerates empty or partially omitted JSON arguments without failing prior to invocation.
func LenientHandler[T any](fn func(ctx context.Context, args T) CallToolResult) ToolHandlerFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) CallToolResult {
		var args T
		if len(rawArgs) > 0 {
			_ = json.Unmarshal(rawArgs, &args)
		}
		return fn(ctx, args)
	}
}

// TextHandler adapts a function returning `(string, error)` into a [ToolHandlerFunc],
// automatically converting errors to [ErrorResult] and strings to [TextResult].
func TextHandler[T any](fn func(ctx context.Context, args T) (string, error)) ToolHandlerFunc {
	return LenientHandler(func(ctx context.Context, args T) CallToolResult {
		out, err := fn(ctx, args)
		if err != nil {
			return ErrorResult(err)
		}
		return TextResult(out)
	})
}

// StrictTextHandler adapts a function returning `(string, error)` into a strict [ToolHandlerFunc]
// that rejects malformed JSON inputs before invoking `fn`.
func StrictTextHandler[T any](fn func(ctx context.Context, args T) (string, error)) ToolHandlerFunc {
	return TypedHandler(func(ctx context.Context, args T) CallToolResult {
		out, err := fn(ctx, args)
		if err != nil {
			return ErrorResult(err)
		}
		return TextResult(out)
	})
}

// NoArgsHandler adapts a zero-argument function returning `(string, error)` into a [ToolHandlerFunc].
func NoArgsHandler(fn func(ctx context.Context) (string, error)) ToolHandlerFunc {
	return func(ctx context.Context, _ json.RawMessage) CallToolResult {
		out, err := fn(ctx)
		if err != nil {
			return ErrorResult(err)
		}
		return TextResult(out)
	}
}

// JSONResult serializes `v` as indented JSON inside a [TextResult], or returns [ErrorResult] on failure.
func JSONResult(v any) CallToolResult {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ErrorResult(err)
	}
	return TextResult(string(data))
}
