// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uiakit

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// WindowID is a portable 64-bit identifier for a native OS window or top-level surface
// (`HWND` on Windows, X11 Window ID or AT-SPI surface ID on Linux, `CGWindowID` on macOS).
type WindowID uint64

// Role represents a universal, platform-independent semantic accessibility control role.
type Role uint8

// Canonical cross-platform accessibility control roles.
const (
	RoleUnknown Role = iota
	RoleWindow
	RoleTitleBar
	RoleMenuBar
	RoleMenu
	RoleMenuItem
	RoleToolBar
	RoleStatusBar
	RoleTab
	RoleTabItem
	RolePane
	RoleGroup
	RoleDocument
	RoleButton
	RoleCheckBox
	RoleRadioButton
	RoleComboBox
	RoleEdit
	RoleLink
	RoleImage
	RoleText
	RoleList
	RoleListItem
	RoleTree
	RoleTreeItem
	RoleDataGrid
	RoleDataItem
	RoleHeader
	RoleHeaderItem
	RoleScrollBar
	RoleSlider
	RoleSpinner
	RoleProgressBar
	RoleCalendar
	RoleToolTip
	RoleCustom
)

// String returns the canonical PascalCase name of the role (e.g. "Button", "CheckBox", "Edit").
func (r Role) String() string {
	switch r {
	case RoleWindow:
		return "Window"
	case RoleTitleBar:
		return "TitleBar"
	case RoleMenuBar:
		return "MenuBar"
	case RoleMenu:
		return "Menu"
	case RoleMenuItem:
		return "MenuItem"
	case RoleToolBar:
		return "ToolBar"
	case RoleStatusBar:
		return "StatusBar"
	case RoleTab:
		return "Tab"
	case RoleTabItem:
		return "TabItem"
	case RolePane:
		return "Pane"
	case RoleGroup:
		return "Group"
	case RoleDocument:
		return "Document"
	case RoleButton:
		return "Button"
	case RoleCheckBox:
		return "CheckBox"
	case RoleRadioButton:
		return "RadioButton"
	case RoleComboBox:
		return "ComboBox"
	case RoleEdit:
		return "Edit"
	case RoleLink:
		return "Hyperlink"
	case RoleImage:
		return "Image"
	case RoleText:
		return "Text"
	case RoleList:
		return "List"
	case RoleListItem:
		return "ListItem"
	case RoleTree:
		return "Tree"
	case RoleTreeItem:
		return "TreeItem"
	case RoleDataGrid:
		return "DataGrid"
	case RoleDataItem:
		return "DataItem"
	case RoleHeader:
		return "Header"
	case RoleHeaderItem:
		return "HeaderItem"
	case RoleScrollBar:
		return "ScrollBar"
	case RoleSlider:
		return "Slider"
	case RoleSpinner:
		return "Spinner"
	case RoleProgressBar:
		return "ProgressBar"
	case RoleCalendar:
		return "Calendar"
	case RoleToolTip:
		return "ToolTip"
	case RoleCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// Interactive reports whether r is a user-actionable control role suitable for Set-of-Marks indexing.
func (r Role) Interactive() bool {
	switch r {
	case RoleButton,
		RoleCheckBox,
		RoleRadioButton,
		RoleComboBox,
		RoleEdit,
		RoleLink,
		RoleListItem,
		RoleMenuItem,
		RoleSlider,
		RoleSpinner,
		RoleTabItem,
		RoleTreeItem,
		RoleDataItem:
		return true
	default:
		return false
	}
}

// ParseRole parses a case-insensitive role name (e.g. "button", "checkbox", "edit", "hyperlink", "link")
// into its canonical [Role].
func ParseRole(s string) Role {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "window", "frame", "dialog":
		return RoleWindow
	case "titlebar", "title_bar":
		return RoleTitleBar
	case "menubar", "menu_bar":
		return RoleMenuBar
	case "menu", "popup_menu":
		return RoleMenu
	case "menuitem", "menu_item":
		return RoleMenuItem
	case "toolbar", "tool_bar":
		return RoleToolBar
	case "statusbar", "status_bar":
		return RoleStatusBar
	case "tab", "page_tab_list":
		return RoleTab
	case "tabitem", "tab_item", "page_tab":
		return RoleTabItem
	case "pane", "panel", "filler":
		return RolePane
	case "group", "grouping":
		return RoleGroup
	case "document", "document_frame", "document_web":
		return RoleDocument
	case "button", "push_button", "toggle_button":
		return RoleButton
	case "checkbox", "check_box":
		return RoleCheckBox
	case "radiobutton", "radio_button":
		return RoleRadioButton
	case "combobox", "combo_box":
		return RoleComboBox
	case "edit", "entry", "text_field", "password_text":
		return RoleEdit
	case "hyperlink", "link":
		return RoleLink
	case "image", "icon", "graphic":
		return RoleImage
	case "text", "label", "static":
		return RoleText
	case "list", "list_box":
		return RoleList
	case "listitem", "list_item":
		return RoleListItem
	case "tree", "tree_table":
		return RoleTree
	case "treeitem", "tree_item":
		return RoleTreeItem
	case "datagrid", "data_grid", "table":
		return RoleDataGrid
	case "dataitem", "data_item", "table_cell", "row":
		return RoleDataItem
	case "header", "table_column_header":
		return RoleHeader
	case "headeritem", "header_item":
		return RoleHeaderItem
	case "scrollbar", "scroll_bar":
		return RoleScrollBar
	case "slider":
		return RoleSlider
	case "spinner", "spin_button":
		return RoleSpinner
	case "progressbar", "progress_bar":
		return RoleProgressBar
	case "calendar":
		return RoleCalendar
	case "tooltip", "tool_tip":
		return RoleToolTip
	case "custom":
		return RoleCustom
	default:
		return RoleUnknown
	}
}

// ActionKind identifies the semantic accessibility pattern or action executed on a control.
type ActionKind string

// Canonical semantic action kinds returned by [Provider] operations.
const (
	ActionInvoke   ActionKind = "invoke"
	ActionToggle   ActionKind = "toggle"
	ActionSelect   ActionKind = "select"
	ActionSetValue ActionKind = "set_value"
	ActionScroll   ActionKind = "scroll"
)

// Windows UIAutomation ControlType IDs (from `UIAutomationClient.h`), retained for
// low-level diagnostics and cross-platform mapping via [RoleFromWindowsUIA].
const (
	UIAButtonControlTypeID      int32 = 50000
	UIACalendarControlTypeID    int32 = 50001
	UIACheckBoxControlTypeID    int32 = 50002
	UIAComboBoxControlTypeID    int32 = 50003
	UIAEditControlTypeID        int32 = 50004
	UIAHyperlinkControlTypeID   int32 = 50005
	UIAImageControlTypeID       int32 = 50006
	UIAListItemControlTypeID    int32 = 50007
	UIAListControlTypeID        int32 = 50008
	UIAMenuControlTypeID        int32 = 50009
	UIAMenuBarControlTypeID     int32 = 50010
	UIAMenuItemControlTypeID    int32 = 50011
	UIAProgressBarControlTypeID int32 = 50012
	UIARadioButtonControlTypeID int32 = 50013
	UIAScrollBarControlTypeID   int32 = 50014
	UIASliderControlTypeID      int32 = 50015
	UIASpinnerControlTypeID     int32 = 50016
	UIAStatusBarControlTypeID   int32 = 50017
	UIATabControlTypeID         int32 = 50018
	UIATabItemControlTypeID     int32 = 50019
	UIATextControlTypeID        int32 = 50020
	UIAToolBarControlTypeID     int32 = 50021
	UIAToolTipControlTypeID     int32 = 50022
	UIATreeControlTypeID        int32 = 50023
	UIATreeItemControlTypeID    int32 = 50024
	UIACustomControlTypeID      int32 = 50025
	UIAGroupControlTypeID       int32 = 50026
	UIADataGridControlTypeID    int32 = 50028
	UIADataItemControlTypeID    int32 = 50029
	UIADocumentControlTypeID    int32 = 50030
	UIAWindowControlTypeID      int32 = 50032
	UIAPaneControlTypeID        int32 = 50033
	UIAHeaderControlTypeID      int32 = 50034
	UIAHeaderItemControlTypeID  int32 = 50035
	UIATitleBarControlTypeID    int32 = 50037
)

// RoleFromWindowsUIA maps a Windows UIAutomation ControlType ID (`50000`..`50037`) to a universal [Role].
func RoleFromWindowsUIA(ctrlType int32) Role {
	switch ctrlType {
	case UIAButtonControlTypeID:
		return RoleButton
	case UIACalendarControlTypeID:
		return RoleCalendar
	case UIACheckBoxControlTypeID:
		return RoleCheckBox
	case UIAComboBoxControlTypeID:
		return RoleComboBox
	case UIAEditControlTypeID:
		return RoleEdit
	case UIAHyperlinkControlTypeID:
		return RoleLink
	case UIAImageControlTypeID:
		return RoleImage
	case UIAListItemControlTypeID:
		return RoleListItem
	case UIAListControlTypeID:
		return RoleList
	case UIAMenuControlTypeID:
		return RoleMenu
	case UIAMenuBarControlTypeID:
		return RoleMenuBar
	case UIAMenuItemControlTypeID:
		return RoleMenuItem
	case UIAProgressBarControlTypeID:
		return RoleProgressBar
	case UIARadioButtonControlTypeID:
		return RoleRadioButton
	case UIAScrollBarControlTypeID:
		return RoleScrollBar
	case UIASliderControlTypeID:
		return RoleSlider
	case UIASpinnerControlTypeID:
		return RoleSpinner
	case UIAStatusBarControlTypeID:
		return RoleStatusBar
	case UIATabControlTypeID:
		return RoleTab
	case UIATabItemControlTypeID:
		return RoleTabItem
	case UIATextControlTypeID:
		return RoleText
	case UIAToolBarControlTypeID:
		return RoleToolBar
	case UIAToolTipControlTypeID:
		return RoleToolTip
	case UIATreeControlTypeID:
		return RoleTree
	case UIATreeItemControlTypeID:
		return RoleTreeItem
	case UIACustomControlTypeID:
		return RoleCustom
	case UIAGroupControlTypeID:
		return RoleGroup
	case UIADataGridControlTypeID:
		return RoleDataGrid
	case UIADataItemControlTypeID:
		return RoleDataItem
	case UIADocumentControlTypeID:
		return RoleDocument
	case UIAWindowControlTypeID:
		return RoleWindow
	case UIAPaneControlTypeID:
		return RolePane
	case UIAHeaderControlTypeID:
		return RoleHeader
	case UIAHeaderItemControlTypeID:
		return RoleHeaderItem
	case UIATitleBarControlTypeID:
		return RoleTitleBar
	default:
		return RoleUnknown
	}
}

// RoleFromLinuxATSPI maps a Linux AT-SPI2 `AtspiRole` numeric constant (`org.a11y.atspi`) to a universal [Role].
func RoleFromLinuxATSPI(atspiRole uint32) Role {
	switch atspiRole {
	case 4: // ATSPI_ROLE_CALENDAR
		return RoleCalendar
	case 7: // ATSPI_ROLE_CHECK_BOX
		return RoleCheckBox
	case 8: // ATSPI_ROLE_CHECK_MENU_ITEM
		return RoleMenuItem
	case 11: // ATSPI_ROLE_COMBO_BOX
		return RoleComboBox
	case 16, 22, 78: // ATSPI_ROLE_DIALOG, ATSPI_ROLE_FRAME, ATSPI_ROLE_WINDOW
		return RoleWindow
	case 19, 80, 94, 96: // ATSPI_ROLE_ENTRY, ATSPI_ROLE_PASSWORD_TEXT, ATSPI_ROLE_DOCUMENT_TEXT
		return RoleEdit
	case 20, 38, 69: // ATSPI_ROLE_FILLER, ATSPI_ROLE_PANEL, ATSPI_ROLE_SPLIT_PANE
		return RolePane
	case 27, 29: // ATSPI_ROLE_ICON, ATSPI_ROLE_IMAGE
		return RoleImage
	case 30, 70: // ATSPI_ROLE_LABEL, ATSPI_ROLE_STATIC
		return RoleText
	case 31, 32: // ATSPI_ROLE_LIST, ATSPI_ROLE_LIST_BOX
		return RoleList
	case 33: // ATSPI_ROLE_LIST_ITEM
		return RoleListItem
	case 34: // ATSPI_ROLE_MENU
		return RoleMenu
	case 35: // ATSPI_ROLE_MENU_BAR
		return RoleMenuBar
	case 36, 45: // ATSPI_ROLE_MENU_ITEM, ATSPI_ROLE_RADIO_MENU_ITEM
		return RoleMenuItem
	case 39: // ATSPI_ROLE_PAGE_TAB
		return RoleTabItem
	case 40: // ATSPI_ROLE_PAGE_TAB_LIST
		return RoleTab
	case 42: // ATSPI_ROLE_PROGRESS_BAR
		return RoleProgressBar
	case 43, 62: // ATSPI_ROLE_PUSH_BUTTON, ATSPI_ROLE_TOGGLE_BUTTON
		return RoleButton
	case 44: // ATSPI_ROLE_RADIO_BUTTON
		return RoleRadioButton
	case 49: // ATSPI_ROLE_SCROLL_BAR
		return RoleScrollBar
	case 52: // ATSPI_ROLE_SLIDER
		return RoleSlider
	case 53: // ATSPI_ROLE_SPIN_BUTTON
		return RoleSpinner
	case 54: // ATSPI_ROLE_STATUS_BAR
		return RoleStatusBar
	case 55: // ATSPI_ROLE_TABLE
		return RoleDataGrid
	case 56, 89: // ATSPI_ROLE_TABLE_CELL, ATSPI_ROLE_TABLE_ROW
		return RoleDataItem
	case 57, 58: // ATSPI_ROLE_TABLE_COLUMN_HEADER, ATSPI_ROLE_TABLE_ROW_HEADER
		return RoleHeaderItem
	case 63: // ATSPI_ROLE_TOOL_BAR
		return RoleToolBar
	case 64: // ATSPI_ROLE_TOOL_TIP
		return RoleToolTip
	case 65, 90: // ATSPI_ROLE_TREE, ATSPI_ROLE_TREE_TABLE
		return RoleTree
	case 91: // ATSPI_ROLE_TREE_ITEM
		return RoleTreeItem
	case 87: // ATSPI_ROLE_LINK
		return RoleLink
	case 92, 95: // ATSPI_ROLE_DOCUMENT_FRAME, ATSPI_ROLE_DOCUMENT_WEB
		return RoleDocument
	default:
		return RoleUnknown
	}
}

// RoleFromDarwinAX maps a macOS Accessibility `AXRole` identifier (e.g. `"AXButton"`, `"AXTextField"`)
// to a universal [Role].
func RoleFromDarwinAX(axRole string) Role {
	switch axRole {
	case "AXWindow", "AXSheet", "AXDrawer":
		return RoleWindow
	case "AXButton", "AXPopUpButton", "AXMenuButton":
		return RoleButton
	case "AXCheckBox":
		return RoleCheckBox
	case "AXRadioButton":
		return RoleRadioButton
	case "AXComboBox":
		return RoleComboBox
	case "AXTextField", "AXTextArea", "AXSearchField":
		return RoleEdit
	case "AXLink":
		return RoleLink
	case "AXImage":
		return RoleImage
	case "AXStaticText":
		return RoleText
	case "AXList":
		return RoleList
	case "AXRow", "AXCell":
		return RoleDataItem
	case "AXMenu":
		return RoleMenu
	case "AXMenuBar":
		return RoleMenuBar
	case "AXMenuItem", "AXMenuBarItem":
		return RoleMenuItem
	case "AXTabGroup":
		return RoleTab
	case "AXToolbar":
		return RoleToolBar
	case "AXScrollBar":
		return RoleScrollBar
	case "AXSlider":
		return RoleSlider
	case "AXIncrementor":
		return RoleSpinner
	case "AXProgressIndicator":
		return RoleProgressBar
	case "AXOutline":
		return RoleTree
	case "AXTable":
		return RoleDataGrid
	case "AXGroup", "AXSplitGroup", "AXScrollArea":
		return RolePane
	case "AXWebArea":
		return RoleDocument
	case "AXHelpTag":
		return RoleToolTip
	default:
		return RoleUnknown
	}
}

// ControlTypeToString maps a Windows UIAutomation ControlType ID to its canonical semantic name.
func ControlTypeToString(ct int32) string {
	if r := RoleFromWindowsUIA(ct); r != RoleUnknown {
		return r.String()
	}
	return fmt.Sprintf("Type_%d", ct)
}

// IsInteractiveControlType reports whether a Windows UIA ControlType ID is user-actionable.
func IsInteractiveControlType(ct int32) bool {
	return RoleFromWindowsUIA(ct).Interactive()
}

// Rect represents a 16-byte bounding box in screen coordinates (`Left`, `Top`, `Right`, `Bottom`).
type Rect struct {
	Left   int32 `json:"left"`
	Top    int32 `json:"top"`
	Right  int32 `json:"right"`
	Bottom int32 `json:"bottom"`
}

// Width returns the horizontal span of r.
func (r Rect) Width() int {
	return int(r.Right - r.Left)
}

// Height returns the vertical span of r.
func (r Rect) Height() int {
	return int(r.Bottom - r.Top)
}

// Center returns the screen-coordinate midpoint (cx, cy) of r.
func (r Rect) Center() (int, int) {
	return int(r.Left) + r.Width()/2, int(r.Top) + r.Height()/2
}

// Contains reports whether screen point (x, y) lies inside r.
func (r Rect) Contains(x, y int) bool {
	return int32(x) >= r.Left && int32(x) < r.Right && int32(y) >= r.Top && int32(y) < r.Bottom
}

// Node represents a universal semantic accessibility node captured by a [Provider].
type Node struct {
	ID              int      `json:"id,omitempty"`
	Role            Role     `json:"role"`
	ControlType     int32    `json:"control_type_id,omitempty"`
	ControlTypeName string   `json:"control_type"`
	LocalizedType   string   `json:"localized_type,omitempty"`
	Name            string   `json:"name,omitempty"`
	AutomationID    string   `json:"automation_id,omitempty"`
	ClassName       string   `json:"class_name,omitempty"`
	FrameworkID     string   `json:"framework_id,omitempty"`
	Value           string   `json:"value,omitempty"`
	ProcessID       int32    `json:"pid,omitempty"`
	Window          WindowID `json:"window,omitempty"`
	ParentWindow    WindowID `json:"parent_window,omitempty"`
	Enabled         bool     `json:"enabled"`
	IsOffscreen     bool     `json:"is_offscreen,omitempty"`
	BoundingRect    Rect     `json:"rect"`
	Depth           int      `json:"depth"`
}

// Interactive reports whether n represents an actionable UI control suitable for Set-of-Marks indexing.
func (n Node) Interactive() bool {
	if !n.Enabled || n.IsOffscreen {
		return false
	}
	if n.BoundingRect.Width() <= 1 || n.BoundingRect.Height() <= 1 {
		return false
	}
	if n.Role != RoleUnknown {
		return n.Role.Interactive()
	}
	return IsInteractiveControlType(n.ControlType)
}

// Selector specifies universal matching criteria for locating a target UI control
// across native (`Window != 0`) or windowless (`Window == 0`) accessibility trees.
type Selector struct {
	Window       WindowID
	ParentWindow WindowID
	AutomationID string
	Title        string
	Role         Role
	ControlType  string
	ClassName    string
	X            int
	Y            int
	Width        int
	Height       int
}

// Center returns the midpoint of the selector's bounding box if Width and Height are positive,
// or (X, Y) directly if only point coordinates were provided.
func (s Selector) Center() (int, int) {
	if s.Width > 0 && s.Height > 0 {
		return s.X + s.Width/2, s.Y + s.Height/2
	}
	return s.X, s.Y
}

// EffectiveRole resolves the target [Role] of the selector from either `s.Role` or `s.ControlType`.
func (s Selector) EffectiveRole() Role {
	if s.Role != RoleUnknown {
		return s.Role
	}
	if s.ControlType != "" {
		return ParseRole(s.ControlType)
	}
	return RoleUnknown
}

// Provider defines the pluggable cross-platform accessibility and Computer Use backend contract.
//
// Implementations include the Windows `CUIAutomation8` provider, Linux AT-SPI2 D-Bus provider,
// macOS `AXUIElement` provider, and remote VMBus/CDP guest proxies.
type Provider interface {
	// EnumerateWindow captures the accessibility subtree of `window` up to `maxNodes`.
	EnumerateWindow(window WindowID, maxNodes int) ([]Node, error)

	// InvokeOrSelect triggers the primary semantic action (`Invoke`, `Toggle`, or `Select`) on `sel`
	// without moving the physical mouse cursor or stealing window focus.
	InvokeOrSelect(sel Selector) (ActionKind, error)

	// SetValue writes `text` atomically into an editable control matched by `sel`.
	SetValue(sel Selector, text string) (ActionKind, error)

	// Scroll scrolls a container matched by `sel` horizontally (`deltaX`) and vertically (`deltaY`).
	Scroll(sel Selector, deltaX, deltaY int) (ActionKind, error)

	// InvalidateCache clears any cached accessibility subtrees held by the provider.
	InvalidateCache()
}

var (
	providerMu      sync.RWMutex
	defaultProvider Provider = newPlatformProvider()
)

// DefaultProvider returns the active process-wide [Provider].
func DefaultProvider() Provider {
	providerMu.RLock()
	p := defaultProvider
	providerMu.RUnlock()
	return p
}

// SetDefaultProvider overrides the process-wide [Provider] (for example, to route accessibility
// operations over a VMBus micro-VM channel or inject a deterministic test double).
func SetDefaultProvider(p Provider) {
	if p == nil {
		p = newPlatformProvider()
	}
	providerMu.Lock()
	defaultProvider = p
	providerMu.Unlock()
}

// EnumerateWindow queries the accessibility subtree of `window` using [DefaultProvider].
func EnumerateWindow(window WindowID, maxNodes int) ([]Node, error) {
	return DefaultProvider().EnumerateWindow(window, maxNodes)
}

// InvokeOrSelect executes a background semantic action (`Invoke`, `Toggle`, or `Select`) using [DefaultProvider].
func InvokeOrSelect(sel Selector) (ActionKind, error) {
	return DefaultProvider().InvokeOrSelect(sel)
}

// SetValue writes `text` into the control matched by `sel` using [DefaultProvider].
func SetValue(sel Selector, text string) (ActionKind, error) {
	return DefaultProvider().SetValue(sel, text)
}

// Scroll scrolls the container matched by `sel` using [DefaultProvider].
func Scroll(sel Selector, deltaX, deltaY int) (ActionKind, error) {
	return DefaultProvider().Scroll(sel, deltaX, deltaY)
}

// InvalidateCache flushes cached accessibility subtrees in [DefaultProvider].
func InvalidateCache() {
	DefaultProvider().InvalidateCache()
}

// StripMnemonicAmpersands removes menu/button mnemonic accelerator ampersands (e.g. "&Save" -> "Save", "A&&B" -> "A&B").
func StripMnemonicAmpersands(s string) string {
	if strings.IndexByte(s, '&') < 0 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '&' {
			if i+1 < len(s) && s[i+1] == '&' {
				b.WriteByte('&')
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// FormatDigest renders a slice of accessibility nodes into a compact, token-efficient Set-of-Marks (SoM)
// text digest optimized for LLM Computer Use prompts.
func FormatDigest(nodes []Node, onlyInteractive bool) string {
	var b strings.Builder
	markID := 1
	for _, n := range nodes {
		if onlyInteractive && !n.Interactive() {
			continue
		}
		id := n.ID
		if id <= 0 {
			id = markID
			markID++
		}
		for d := 0; d < n.Depth && d < 6; d++ {
			b.WriteString("  ")
		}
		b.WriteString("[#")
		b.WriteString(strconv.Itoa(id))
		b.WriteByte(' ')
		switch {
		case n.Role != RoleUnknown:
			b.WriteString(n.Role.String())
		case n.ControlTypeName != "":
			b.WriteString(n.ControlTypeName)
		default:
			b.WriteString(ControlTypeToString(n.ControlType))
		}
		if n.Name != "" {
			b.WriteString(` "`)
			b.WriteString(StripMnemonicAmpersands(n.Name))
			b.WriteByte('"')
		}
		if n.AutomationID != "" {
			b.WriteString(` id="`)
			b.WriteString(n.AutomationID)
			b.WriteByte('"')
		}
		if n.Value != "" {
			b.WriteString(` val="`)
			b.WriteString(n.Value)
			b.WriteByte('"')
		}
		if !n.Enabled {
			b.WriteString(" disabled")
		}
		b.WriteString(" @(")
		b.WriteString(strconv.Itoa(int(n.BoundingRect.Left)))
		b.WriteByte(',')
		b.WriteString(strconv.Itoa(int(n.BoundingRect.Top)))
		b.WriteByte(' ')
		b.WriteString(strconv.Itoa(n.BoundingRect.Width()))
		b.WriteByte('x')
		b.WriteString(strconv.Itoa(n.BoundingRect.Height()))
		b.WriteString(")]\n")
	}
	return b.String()
}
