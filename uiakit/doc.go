// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package uiakit provides a universal, cross-platform UI Automation and Computer Use
// accessibility engine for Go.
//
// It unifies desktop accessibility inspection and background semantic control across operating systems:
//   - Windows: UI Automation (`CUIAutomation8` / `IUIAutomation2` / `FindAllBuildCache`) built on
//     [github.com/lemon4ksan/foundation/silicon/comkit].
//   - Linux: AT-SPI2 (`org.a11y.atspi.Cache` / `Accessible` / `Action` / `EditableText`) role and node model.
//   - macOS: Accessibility (`AXUIElement` / `AXRole`) role and node model.
//   - Remote / Virtualized: Pluggable [Provider] interface for proxying accessibility operations
//     over VMBus, SSH, or Chrome DevTools Protocol (CDP).
//
// Using [EnumerateWindow] and [FormatDigest], AI agents and CLI tools can capture an entire window's
// accessibility hierarchy in a single batched IPC call, render compact Set-of-Marks (SoM) digests,
// and trigger background actions ([InvokeOrSelect], [SetValue], [Scroll]) without stealing focus
// or moving the physical mouse cursor.
package uiakit
