// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uiakit_test

import (
	"strings"
	"testing"

	"github.com/lemon4ksan/ctrl/uiakit"
)

func TestRect_And_Selector(t *testing.T) {
	r := uiakit.Rect{Left: 100, Top: 200, Right: 260, Bottom: 240}
	if r.Width() != 160 || r.Height() != 40 {
		t.Fatalf("Rect dimensions = %dx%d, want 160x40", r.Width(), r.Height())
	}
	cx, cy := r.Center()
	if cx != 180 || cy != 220 {
		t.Fatalf("Rect.Center() = (%d, %d), want (180, 220)", cx, cy)
	}
	if !r.Contains(180, 220) || r.Contains(99, 220) || r.Contains(260, 240) {
		t.Fatal("Rect.Contains returned unexpected result")
	}

	sel := uiakit.Selector{X: 10, Y: 20, Width: 100, Height: 50, ControlType: "button"}
	scx, scy := sel.Center()
	if scx != 60 || scy != 45 {
		t.Fatalf("Selector.Center() = (%d, %d), want (60, 45)", scx, scy)
	}
	if sel.EffectiveRole() != uiakit.RoleButton {
		t.Fatalf("Selector.EffectiveRole() = %v, want RoleButton", sel.EffectiveRole())
	}

	ptSel := uiakit.Selector{X: 42, Y: 84, Role: uiakit.RoleCheckBox}
	px, py := ptSel.Center()
	if px != 42 || py != 84 {
		t.Fatalf("point Selector.Center() = (%d, %d), want (42, 84)", px, py)
	}
	if ptSel.EffectiveRole() != uiakit.RoleCheckBox {
		t.Fatalf("ptSel.EffectiveRole() = %v, want RoleCheckBox", ptSel.EffectiveRole())
	}
}

func TestCrossPlatformRoles_And_Mnemonics(t *testing.T) {
	// Windows UIA mapping
	if got := uiakit.RoleFromWindowsUIA(uiakit.UIAButtonControlTypeID); got != uiakit.RoleButton {
		t.Fatalf("RoleFromWindowsUIA(Button) = %v", got)
	}
	if got := uiakit.ControlTypeToString(uiakit.UIAButtonControlTypeID); got != "Button" {
		t.Fatalf("ControlTypeToString(Button) = %q", got)
	}
	if got := uiakit.ControlTypeToString(99999); got != "Type_99999" {
		t.Fatalf("ControlTypeToString(99999) = %q", got)
	}

	// Linux AT-SPI2 mapping
	if got := uiakit.RoleFromLinuxATSPI(43); got != uiakit.RoleButton {
		t.Fatalf("RoleFromLinuxATSPI(43/PUSH_BUTTON) = %v, want RoleButton", got)
	}
	if got := uiakit.RoleFromLinuxATSPI(19); got != uiakit.RoleEdit {
		t.Fatalf("RoleFromLinuxATSPI(19/ENTRY) = %v, want RoleEdit", got)
	}
	if got := uiakit.RoleFromLinuxATSPI(87); got != uiakit.RoleLink {
		t.Fatalf("RoleFromLinuxATSPI(87/LINK) = %v, want RoleLink", got)
	}

	// macOS Accessibility AXRole mapping
	if got := uiakit.RoleFromDarwinAX("AXButton"); got != uiakit.RoleButton {
		t.Fatalf("RoleFromDarwinAX(AXButton) = %v, want RoleButton", got)
	}
	if got := uiakit.RoleFromDarwinAX("AXTextField"); got != uiakit.RoleEdit {
		t.Fatalf("RoleFromDarwinAX(AXTextField) = %v, want RoleEdit", got)
	}
	if got := uiakit.RoleFromDarwinAX("AXCheckBox"); got != uiakit.RoleCheckBox {
		t.Fatalf("RoleFromDarwinAX(AXCheckBox) = %v, want RoleCheckBox", got)
	}

	// ParseRole & Interactive
	if uiakit.ParseRole("push_button") != uiakit.RoleButton || !uiakit.RoleButton.Interactive() {
		t.Fatal("expected push_button -> interactive RoleButton")
	}
	if uiakit.ParseRole("pane") != uiakit.RolePane || uiakit.RolePane.Interactive() {
		t.Fatal("expected pane -> non-interactive RolePane")
	}

	if got := uiakit.StripMnemonicAmpersands("&Save && &Close"); got != "Save & Close" {
		t.Fatalf("StripMnemonicAmpersands = %q, want %q", got, "Save & Close")
	}
	if got := uiakit.StripMnemonicAmpersands("Plain Text"); got != "Plain Text" {
		t.Fatalf("StripMnemonicAmpersands = %q, want %q", got, "Plain Text")
	}
}

func TestNodeInteractive_And_FormatDigest(t *testing.T) {
	nodes := []uiakit.Node{
		{
			ID:           1,
			Role:         uiakit.RoleWindow,
			Name:         "Settings",
			Enabled:      true,
			BoundingRect: uiakit.Rect{Left: 0, Top: 0, Right: 800, Bottom: 600},
			Depth:        0,
		},
		{
			ID:           2,
			Role:         uiakit.RoleButton,
			Name:         "&Save",
			AutomationID: "btnSave",
			Enabled:      true,
			BoundingRect: uiakit.Rect{Left: 120, Top: 45, Right: 200, Bottom: 73},
			Depth:        1,
		},
		{
			ID:           3,
			Role:         uiakit.RoleEdit,
			Name:         "Username",
			Value:        "alice",
			Enabled:      false,
			BoundingRect: uiakit.Rect{Left: 120, Top: 90, Right: 320, Bottom: 118},
			Depth:        1,
		},
	}

	if nodes[0].Interactive() {
		t.Fatal("Window node should not be interactive")
	}
	if !nodes[1].Interactive() {
		t.Fatal("Enabled Button node should be interactive")
	}
	if nodes[2].Interactive() {
		t.Fatal("Disabled Edit node should not be interactive")
	}

	fullDigest := uiakit.FormatDigest(nodes, false)
	if !strings.Contains(fullDigest, `[#2 Button "Save" id="btnSave" @(120,45 80x28)]`) {
		t.Fatalf("unexpected fullDigest:\n%s", fullDigest)
	}
	if !strings.Contains(fullDigest, `val="alice" disabled`) {
		t.Fatalf("expected disabled value in fullDigest:\n%s", fullDigest)
	}

	interactiveDigest := uiakit.FormatDigest(nodes, true)
	if strings.Contains(interactiveDigest, "Window") || strings.Contains(interactiveDigest, "Username") {
		t.Fatalf("interactiveDigest should filter non-interactive nodes:\n%s", interactiveDigest)
	}
}

type mockProvider struct {
	invalidated bool
	lastText    string
}

func (m *mockProvider) EnumerateWindow(w uiakit.WindowID, _ int) ([]uiakit.Node, error) {
	return []uiakit.Node{{ID: 1, Role: uiakit.RoleButton, Name: "OK", Window: w, Enabled: true}}, nil
}

func (m *mockProvider) InvokeOrSelect(_ uiakit.Selector) (uiakit.ActionKind, error) {
	return uiakit.ActionInvoke, nil
}

func (m *mockProvider) SetValue(_ uiakit.Selector, text string) (uiakit.ActionKind, error) {
	m.lastText = text
	return uiakit.ActionSetValue, nil
}

func (m *mockProvider) Scroll(_ uiakit.Selector, _, _ int) (uiakit.ActionKind, error) {
	return uiakit.ActionScroll, nil
}

func (m *mockProvider) InvalidateCache() {
	m.invalidated = true
}

func TestCustomProviderInjection(t *testing.T) {
	orig := uiakit.DefaultProvider()
	defer uiakit.SetDefaultProvider(orig)

	mock := &mockProvider{}
	uiakit.SetDefaultProvider(mock)

	nodes, err := uiakit.EnumerateWindow(0x1234, 10)
	if err != nil || len(nodes) != 1 || nodes[0].Role != uiakit.RoleButton {
		t.Fatalf("EnumerateWindow via mock provider failed: %v, %+v", err, nodes)
	}
	if act, err := uiakit.InvokeOrSelect(uiakit.Selector{Title: "OK"}); err != nil || act != uiakit.ActionInvoke {
		t.Fatalf("InvokeOrSelect via mock failed: %v, %v", act, err)
	}
	if act, err := uiakit.SetValue(
		uiakit.Selector{Title: "Input"},
		"hello",
	); err != nil || act != uiakit.ActionSetValue ||
		mock.lastText != "hello" {
		t.Fatalf("SetValue via mock failed: %v, %v", act, err)
	}
	if act, err := uiakit.Scroll(uiakit.Selector{Title: "List"}, 0, -120); err != nil || act != uiakit.ActionScroll {
		t.Fatalf("Scroll via mock failed: %v, %v", act, err)
	}
	uiakit.InvalidateCache()
	if !mock.invalidated {
		t.Fatal("expected InvalidateCache to be forwarded to mock provider")
	}
}
