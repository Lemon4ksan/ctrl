// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !windows

package uiakit

import "errors"

var errNoNativeBackend = errors.New(
	"uiakit: no native accessibility provider is active; register a custom Provider via SetDefaultProvider",
)

type fallbackProvider struct{}

func newPlatformProvider() Provider {
	return fallbackProvider{}
}

func (fallbackProvider) EnumerateWindow(_ WindowID, _ int) ([]Node, error) {
	return nil, errNoNativeBackend
}

func (fallbackProvider) InvokeOrSelect(_ Selector) (ActionKind, error) {
	return "", errNoNativeBackend
}

func (fallbackProvider) SetValue(_ Selector, _ string) (ActionKind, error) {
	return "", errNoNativeBackend
}

func (fallbackProvider) Scroll(_ Selector, _, _ int) (ActionKind, error) {
	return "", errNoNativeBackend
}

func (fallbackProvider) InvalidateCache() {}
