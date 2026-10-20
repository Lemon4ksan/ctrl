# ADR-0001: Universal cross-platform UI Automation and Computer Use engine in `uiakit`

Status: Accepted
Date: 2026-10-20
Deciders: Lemon4ksan

## Context

AI agents, MCP servers (`mcpkit`), and CLI automation tools (such as `stasis`) increasingly require **Computer Use** capabilities: inspecting desktop and application UI hierarchies, assigning Set-of-Marks (SoM) numeric badges, and triggering UI controls without stealing window focus or hijacking the user's physical mouse and keyboard.

Historically, desktop accessibility APIs are fragmented across operating systems:
- **Windows**: Component Object Model (COM) `CUIAutomation8` / `IUIAutomation2` (`UIAutomationCore.dll`).
- **Linux**: **AT-SPI2** (`org.a11y.atspi.Cache`, `Accessible`, `Action`, `EditableText`, `Component`) over a local D-Bus UNIX socket (`org.a11y.Bus`).
- **macOS**: **Accessibility API (`AXUIElement`)** (`kAXRoleAttribute`, `kAXPressAction`, `kAXValueAttribute`) over Mach IPC.

Furthermore, naïve implementations suffer from three severe bottlenecks:
1. **Cross-Process RPC Amplification**: Walking an accessibility tree node-by-node triggers $O(N \times P)$ cross-process IPC roundtrips (often 150–400 ms per window).
2. **Focus Stealing and Hang Timeouts**: Default OS automation settings often focus target controls or block for 2,000+ ms on hung background windows.
3. **LLM Context Bloat**: Dumping raw accessibility XML/JSON trees consumes 8,000–25,000 tokens per screen observation.

## Decision

We introduce **`uiakit`** in `ctrl` as a **universal, cross-platform Computer Use and Accessibility engine**:

1. **OS-Agnostic Domain Model (`Role`, `WindowID`, `ActionKind`, `Provider`)**:
   - Define a canonical `Role` enum (`RoleButton`, `RoleCheckBox`, `RoleEdit`, `RoleWindow`, `RoleTabItem`, `RoleListItem`, `RoleMenuItem`, `RoleLink`, `RoleTree`, etc.) with deterministic translators for Windows UIA (`RoleFromWindowsUIA`), Linux AT-SPI2 (`RoleFromLinuxATSPI`), and macOS Accessibility (`RoleFromDarwinAX`).
   - Represent window and surface handles as portable 64-bit identifiers (`WindowID uint64`) rather than OS-specific pointers (`HWND`, X11 `Window`, or `CGWindowID`).
   - Represent semantic interactions via `ActionKind` (`ActionInvoke`, `ActionToggle`, `ActionSelect`, `ActionSetValue`, `ActionScroll`).
   - Abstract execution behind the `Provider` interface (`EnumerateWindow`, `InvokeOrSelect`, `SetValue`, `Scroll`, `InvalidateCache`) with a swappable `DefaultProvider()`, enabling both native OS backends and remote/virtualized proxies (e.g., `stasis` VMBus guest agents or Chrome CDP).

2. **Batched Single-IPC Subtree Snapshots**:
   - On **Windows** (`uia_windows.go`), built on `foundation/silicon/comkit`: activate a process-wide MTA `CUIAutomation8` singleton tuned via `IUIAutomation2` (`AutoSetFocus = FALSE`, `ConnectionTimeout = 150ms`, `TransactionTimeout = 500ms`) and fetch entire window subtrees in a single COM RPC call via `IUIAutomationElement::FindAllBuildCache` (vtable slot 8).
   - On **Linux** (`atspi_linux.go` / `RoleFromLinuxATSPI`): map AT-SPI2 roles (`AtspiRole`) and batch-fetch subtree items via `org.a11y.atspi.Cache.GetItems`.
   - On **macOS** (`RoleFromDarwinAX`): map `AXRole` strings (`AXButton`, `AXTextField`, `AXCheckBox`, etc.) to canonical `Role` values.

3. **Zero-Input Background Semantic Actions**:
   - Dispatch control actions (`Invoke`, `Toggle`, `Select`, `SetValue`, `Scroll`) directly through accessibility patterns without moving the physical cursor or stealing window focus, supporting both native windows and windowless UI frameworks (WPF, WinUI 3, Chromium/Electron, GTK4, Qt6).

4. **Token-Efficient Set-of-Marks Digest (`FormatDigest`)**:
   - Format interactive and structural `Node` slices into a compact one-line-per-element Set-of-Marks (SoM) digest (`[#1 Button "Save" id="btnSave" @(120,45 80x28)]`), reducing LLM token consumption by ~70%.

## Consequences

- What gets easier:
  - Consumer packages (`stasis`, `ctrl/mcpkit`, CLI tools) write 100% platform-independent Go code using `uiakit.Node`, `uiakit.Role`, `uiakit.Selector`, and `uiakit.Provider`.
  - Virtualized environments (like `stasis`) or unit tests can register custom `uiakit.Provider` implementations (e.g., forwarding `uiakit` calls over a VMBus pipe into a micro-VM).
  - Windows UIA inspection runs in < 5 ms per window with zero heap allocations on vtable dispatch via `foundation/silicon/comkit`.
- What gets harder / what we pay:
  - Platform-specific exotic control properties that do not map to the universal `Node` struct require inspecting `Node.NativeControlType` or `Node.FrameworkID`.
  - Custom owner-drawn canvas controls that expose no accessibility tree still require coordinate-based pointer fallback in the caller.

## Alternatives considered

- **Exposing raw Windows `UIA_*` constants and `HWND` in the public API**: Rejected. Leaking Windows-only `UIA_ButtonControlTypeId = 50000` and `uintptr` window handles into consumer code prevents portable Linux AT-SPI2, macOS `AXUIElement`, and remote VMBus provider implementations.
- **Walking trees node-by-node (`IUIAutomationTreeWalker` / per-node D-Bus queries)**: Rejected. Per-node IPC roundtrips are 20–50× slower than batched cache snapshots (`FindAllBuildCache` on Windows, `Cache.GetItems` on Linux AT-SPI2).
