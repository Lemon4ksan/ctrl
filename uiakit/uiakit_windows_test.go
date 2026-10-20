// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package uiakit_test

import (
	"syscall"
	"testing"

	"github.com/lemon4ksan/ctrl/uiakit"
)

var (
	modUser32            = syscall.NewLazyDLL("user32.dll")
	procGetDesktopWindow = modUser32.NewProc("GetDesktopWindow")
)

func TestEnumerateWindow_Desktop(t *testing.T) {
	if _, err := uiakit.EnumerateWindow(0, 10); err == nil {
		t.Fatal("expected error for WindowID 0")
	}

	hwnd, _, _ := procGetDesktopWindow.Call()
	if hwnd == 0 {
		t.Skip("GetDesktopWindow returned 0")
	}

	winID := uiakit.WindowID(hwnd)
	uiakit.InvalidateCache()
	nodes, err := uiakit.EnumerateWindow(winID, 16)
	if err != nil {
		t.Fatalf("EnumerateWindow(Desktop) failed: %v", err)
	}
	if len(nodes) == 0 {
		t.Fatal("expected at least 1 UIA node for Desktop window")
	}

	// Second call should hit the short-lived per-WindowID cache
	cachedNodes, err := uiakit.EnumerateWindow(winID, 8)
	if err != nil {
		t.Fatalf("cached EnumerateWindow failed: %v", err)
	}
	if len(cachedNodes) > 8 {
		t.Fatalf("expected at most 8 nodes, got %d", len(cachedNodes))
	}
}
