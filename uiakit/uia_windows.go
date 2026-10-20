// Copyright (c) 2026 Lemon4ksan All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package uiakit

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/lemon4ksan/foundation/silicon/comkit"
)

// Windows UIAutomation TreeScope flags, Pattern IDs, Property IDs, and ScrollAmount constants.
const (
	TreeScopeElement     uintptr = 0x1
	TreeScopeChildren    uintptr = 0x2
	TreeScopeDescendants uintptr = 0x4
	TreeScopeSubtree     uintptr = TreeScopeElement | TreeScopeChildren | TreeScopeDescendants

	UIAInvokePatternID        uintptr = 10000
	UIAValuePatternID         uintptr = 10002
	UIAScrollPatternID        uintptr = 10004
	UIASelectionItemPatternID uintptr = 10010
	UIATogglePatternID        uintptr = 10015

	UIABoundingRectanglePropertyID  uintptr = 30001
	UIAProcessIDPropertyID          uintptr = 30002
	UIAControlTypePropertyID        uintptr = 30003
	UIALocalizedControlTypeProperty uintptr = 30004
	UIANamePropertyID               uintptr = 30005
	UIAIsEnabledPropertyID          uintptr = 30010
	UIAAutomationIDPropertyID       uintptr = 30011
	UIAClassNamePropertyID          uintptr = 30012
	UIANativeWindowHandlePropertyID uintptr = 30020
	UIAIsOffscreenPropertyID        uintptr = 30022
	UIAFrameworkIDPropertyID        uintptr = 30024

	ScrollAmountLargeDecrement uintptr = 0
	ScrollAmountSmallDecrement uintptr = 1
	ScrollAmountNoAmount       uintptr = 2
	ScrollAmountLargeIncrement uintptr = 3
	ScrollAmountSmallIncrement uintptr = 4
)

var (
	// CLSIDCUIAutomation8 is the Windows 8+ / Windows 11 CUIAutomation8 class ID ({e22ad333-b25f-460c-83d0-0581107395c9}).
	CLSIDCUIAutomation8 = comkit.GUID{
		Data1: 0xe22ad333,
		Data2: 0xb25f,
		Data3: 0x460c,
		Data4: [8]byte{0x83, 0xd0, 0x05, 0x81, 0x10, 0x73, 0x95, 0xc9},
	}

	// CLSIDCUIAutomation is the legacy CUIAutomation class ID ({ff48dba4-60ef-4201-aa87-54103eef594e}).
	CLSIDCUIAutomation = comkit.GUID{
		Data1: 0xff48dba4,
		Data2: 0x60ef,
		Data3: 0x4201,
		Data4: [8]byte{0xaa, 0x87, 0x54, 0x10, 0x3e, 0xef, 0x59, 0x4e},
	}

	// IIDIUIAutomation is the base IUIAutomation interface ID ({30cbe57d-d9d0-452a-ab13-7ac5ac4825ee}).
	IIDIUIAutomation = comkit.GUID{
		Data1: 0x30cbe57d,
		Data2: 0xd9d0,
		Data3: 0x452a,
		Data4: [8]byte{0xab, 0x13, 0x7a, 0xc5, 0xac, 0x48, 0x25, 0xee},
	}

	// IIDIUIAutomation2 is the IUIAutomation2 interface ID ({34723aff-0c9d-49d0-9896-7ab52df8cd8a}).
	IIDIUIAutomation2 = comkit.GUID{
		Data1: 0x34723aff,
		Data2: 0x0c9d,
		Data3: 0x49d0,
		Data4: [8]byte{0x98, 0x96, 0x7a, 0xb5, 0x2d, 0xf8, 0xcd, 0x8a},
	}
)

var cachedPropIDs = [...]uintptr{
	UIAProcessIDPropertyID,
	UIAControlTypePropertyID,
	UIALocalizedControlTypeProperty,
	UIANamePropertyID,
	UIAIsEnabledPropertyID,
	UIAAutomationIDPropertyID,
	UIAClassNamePropertyID,
	UIANativeWindowHandlePropertyID,
	UIAIsOffscreenPropertyID,
	UIAFrameworkIDPropertyID,
	UIABoundingRectanglePropertyID,
}

type cacheEntry struct {
	nodes     []Node
	fetchedAt time.Time
}

type windowsUIAProvider struct {
	mu      sync.RWMutex
	cache   map[WindowID]cacheEntry
	ttl     time.Duration
	uiaOnce sync.Once
	uiaPtr  comkit.Ptr
	uiaErr  error
}

func newPlatformProvider() Provider {
	return &windowsUIAProvider{
		cache: make(map[WindowID]cacheEntry),
		ttl:   30 * time.Millisecond,
	}
}

func (p *windowsUIAProvider) InvalidateCache() {
	p.mu.Lock()
	clear(p.cache)
	p.mu.Unlock()
}

func (p *windowsUIAProvider) getAutomation() (comkit.Ptr, error) {
	if err := comkit.CoInitializeEx(comkit.CoInitMultithreaded); err != nil {
		return comkit.Ptr{}, err
	}

	p.uiaOnce.Do(func() {
		uia, err := comkit.CoCreateInstance(&CLSIDCUIAutomation8, 0, comkit.ClsctxInprocServer, &IIDIUIAutomation)
		if err != nil || uia.IsNil() {
			uia, err = comkit.CoCreateInstance(&CLSIDCUIAutomation, 0, comkit.ClsctxInprocServer, &IIDIUIAutomation)
		}
		if err != nil || uia.IsNil() {
			p.uiaErr = fmt.Errorf("uiakit: CoCreateInstance(IUIAutomation) failed: %w", err)
			return
		}

		// Configure IUIAutomation2 for zero-focus background execution and bounded timeouts.
		if uia2, hr := uia.QueryInterface(&IIDIUIAutomation2); hr.Succeeded() && !uia2.IsNil() {
			_ = comkit.VtblCall1(uia2.Raw(), 59, 0)   // IUIAutomation2::put_AutoSetFocus(FALSE)
			_ = comkit.VtblCall1(uia2.Raw(), 61, 150) // IUIAutomation2::put_ConnectionTimeout(150ms)
			_ = comkit.VtblCall1(uia2.Raw(), 63, 500) // IUIAutomation2::put_TransactionTimeout(500ms)
			uia2.Release()
		}

		p.uiaPtr = uia
	})

	if p.uiaErr != nil || p.uiaPtr.IsNil() {
		return comkit.Ptr{}, p.uiaErr
	}
	return p.uiaPtr, nil
}

// inspectCachedElement reads cached properties from an IUIAutomationElement populated via FindAllBuildCache.
// Each get_Cached* vtable slot is located at CurrentIdx + 32 in IUIAutomationElementVtbl.
func inspectCachedElement(elem uintptr, parentWindow WindowID) Node {
	var pid int32
	_ = comkit.VtblCall1(elem, 52, uintptr(unsafe.Pointer(&pid))) // get_CachedProcessId (20 + 32)
	var ctrlType int32
	_ = comkit.VtblCall1(elem, 53, uintptr(unsafe.Pointer(&ctrlType))) // get_CachedControlType (21 + 32)
	var locTypeBSTR comkit.BSTR
	_ = comkit.VtblCall1(elem, 54, uintptr(unsafe.Pointer(&locTypeBSTR))) // get_CachedLocalizedControlType (22 + 32)
	var nameBSTR comkit.BSTR
	_ = comkit.VtblCall1(elem, 55, uintptr(unsafe.Pointer(&nameBSTR))) // get_CachedName (23 + 32)
	var enabled int32
	_ = comkit.VtblCall1(elem, 60, uintptr(unsafe.Pointer(&enabled))) // get_CachedIsEnabled (28 + 32)
	var autoIDBSTR comkit.BSTR
	_ = comkit.VtblCall1(elem, 61, uintptr(unsafe.Pointer(&autoIDBSTR))) // get_CachedAutomationId (29 + 32)
	var classBSTR comkit.BSTR
	_ = comkit.VtblCall1(elem, 62, uintptr(unsafe.Pointer(&classBSTR))) // get_CachedClassName (30 + 32)
	var hwnd uintptr
	_ = comkit.VtblCall1(elem, 68, uintptr(unsafe.Pointer(&hwnd))) // get_CachedNativeWindowHandle (36 + 32)
	var offscreen int32
	_ = comkit.VtblCall1(elem, 70, uintptr(unsafe.Pointer(&offscreen))) // get_CachedIsOffscreen (38 + 32)
	var fwBSTR comkit.BSTR
	_ = comkit.VtblCall1(elem, 72, uintptr(unsafe.Pointer(&fwBSTR))) // get_CachedFrameworkId (40 + 32)
	var rect Rect
	_ = comkit.VtblCall1(elem, 75, uintptr(unsafe.Pointer(&rect))) // get_CachedBoundingRectangle (43 + 32)

	role := RoleFromWindowsUIA(ctrlType)
	var valStr string
	if hwnd == 0 && (role == RoleEdit || role == RoleComboBox) {
		var valPat uintptr
		if comkit.VtblCall2(elem, 16, UIAValuePatternID, uintptr(unsafe.Pointer(&valPat))).Succeeded() && valPat != 0 {
			var valBSTR comkit.BSTR
			if comkit.VtblCall1(valPat, 4, uintptr(unsafe.Pointer(&valBSTR))).Succeeded() {
				valStr = comkit.BSTRToString(valBSTR)
			}
			comkit.ComRelease(valPat)
		}
	}

	return Node{
		Role:            role,
		ControlType:     ctrlType,
		ControlTypeName: ControlTypeToString(ctrlType),
		LocalizedType:   comkit.BSTRToString(locTypeBSTR),
		Name:            comkit.BSTRToString(nameBSTR),
		AutomationID:    comkit.BSTRToString(autoIDBSTR),
		ClassName:       comkit.BSTRToString(classBSTR),
		FrameworkID:     comkit.BSTRToString(fwBSTR),
		Value:           valStr,
		ProcessID:       pid,
		Window:          WindowID(hwnd),
		ParentWindow:    parentWindow,
		Enabled:         enabled != 0,
		IsOffscreen:     offscreen != 0,
		BoundingRect:    rect,
		Depth:           1,
	}
}

func (p *windowsUIAProvider) EnumerateWindow(window WindowID, maxNodes int) ([]Node, error) {
	if window == 0 {
		return nil, errors.New("uiakit: invalid WindowID 0")
	}
	if maxNodes <= 0 {
		maxNodes = 128
	}

	p.mu.RLock()
	if entry, ok := p.cache[window]; ok && time.Since(entry.fetchedAt) < p.ttl {
		cached := entry.nodes
		p.mu.RUnlock()
		if len(cached) > maxNodes {
			return cached[:maxNodes], nil
		}
		return cached, nil
	}
	p.mu.RUnlock()

	unlock, err := comkit.LockThreadMTA()
	if err != nil {
		return nil, err
	}
	defer unlock()

	uia, err := p.getAutomation()
	if err != nil {
		return nil, err
	}

	hwnd := uintptr(window)
	var winElem uintptr
	if hr := comkit.VtblCall2(uia.Raw(), 6, hwnd, uintptr(unsafe.Pointer(&winElem))); hr.Failed() || winElem == 0 {
		return nil, hr.Err("IUIAutomation::ElementFromHandle")
	}
	defer comkit.ComRelease(winElem)

	var cacheReq uintptr
	if hr := comkit.VtblCall1(uia.Raw(), 20, uintptr(unsafe.Pointer(&cacheReq))); hr.Failed() || cacheReq == 0 {
		return nil, hr.Err("IUIAutomation::CreateCacheRequest")
	}
	defer comkit.ComRelease(cacheReq)

	for _, propID := range cachedPropIDs {
		_ = comkit.VtblCall1(cacheReq, 3, propID) // IUIAutomationCacheRequest::AddProperty
	}

	var trueCond uintptr
	if hr := comkit.VtblCall1(uia.Raw(), 21, uintptr(unsafe.Pointer(&trueCond))); hr.Failed() || trueCond == 0 {
		return nil, hr.Err("IUIAutomation::CreateTrueCondition")
	}
	defer comkit.ComRelease(trueCond)

	var cachedArray uintptr
	if hr := comkit.VtblCall4(
		winElem,
		8,
		TreeScopeSubtree,
		trueCond,
		cacheReq,
		uintptr(unsafe.Pointer(&cachedArray)),
	); hr.Failed() ||
		cachedArray == 0 {
		return nil, hr.Err("IUIAutomationElement::FindAllBuildCache")
	}
	defer comkit.ComRelease(cachedArray)

	var length int32
	_ = comkit.VtblCall1(cachedArray, 3, uintptr(unsafe.Pointer(&length)))
	if length <= 0 {
		return nil, nil
	}

	limit := int(length)
	if limit > maxNodes {
		limit = maxNodes
	}
	nodes := make([]Node, 0, limit)
	for i := int32(0); i < length && len(nodes) < maxNodes; i++ {
		var item uintptr
		if comkit.VtblCall2(cachedArray, 4, uintptr(i), uintptr(unsafe.Pointer(&item))).Succeeded() && item != 0 {
			n := inspectCachedElement(item, window)
			comkit.ComRelease(item)
			n.ID = len(nodes) + 1
			if i == 0 && n.Window == window {
				n.Depth = 0
			} else if n.Window == 0 {
				n.Depth = 2
			} else {
				n.Depth = 1
			}
			nodes = append(nodes, n)
		}
	}

	p.mu.Lock()
	p.cache[window] = cacheEntry{
		nodes:     nodes,
		fetchedAt: time.Now(),
	}
	p.mu.Unlock()

	return nodes, nil
}

func findMatchingElement(uia comkit.Ptr, sel Selector) (uintptr, error) {
	if sel.Window != 0 {
		var directElem uintptr
		if comkit.VtblCall2(uia.Raw(), 6, uintptr(sel.Window), uintptr(unsafe.Pointer(&directElem))).Succeeded() &&
			directElem != 0 {
			return directElem, nil
		}
	}

	rootWindow := sel.ParentWindow
	if rootWindow == 0 {
		rootWindow = sel.Window
	}
	if rootWindow != 0 {
		var rootElem uintptr
		if comkit.VtblCall2(uia.Raw(), 6, uintptr(rootWindow), uintptr(unsafe.Pointer(&rootElem))).Succeeded() &&
			rootElem != 0 {
			defer comkit.ComRelease(rootElem)

			var cacheReq uintptr
			if comkit.VtblCall1(uia.Raw(), 20, uintptr(unsafe.Pointer(&cacheReq))).Succeeded() && cacheReq != 0 {
				defer comkit.ComRelease(cacheReq)
				for _, propID := range cachedPropIDs {
					_ = comkit.VtblCall1(cacheReq, 3, propID)
				}

				var trueCond uintptr
				if comkit.VtblCall1(uia.Raw(), 21, uintptr(unsafe.Pointer(&trueCond))).Succeeded() && trueCond != 0 {
					defer comkit.ComRelease(trueCond)
					var arr uintptr
					if comkit.VtblCall4(rootElem, 8, TreeScopeSubtree, trueCond, cacheReq, uintptr(unsafe.Pointer(&arr))).
						Succeeded() &&
						arr != 0 {
						defer comkit.ComRelease(arr)
						var length int32
						_ = comkit.VtblCall1(arr, 3, uintptr(unsafe.Pointer(&length)))

						targetTitle := strings.ToLower(strings.TrimSpace(StripMnemonicAmpersands(sel.Title)))
						wantRole := sel.EffectiveRole()
						var bestMatch uintptr
						for i := int32(0); i < length; i++ {
							var cand uintptr
							if comkit.VtblCall2(arr, 4, uintptr(i), uintptr(unsafe.Pointer(&cand))).Failed() ||
								cand == 0 {
								continue
							}
							node := inspectCachedElement(cand, rootWindow)
							if sel.Window != 0 && node.Window == sel.Window {
								comkit.ComRelease(bestMatch)
								return cand, nil
							}
							if sel.AutomationID != "" && node.AutomationID == sel.AutomationID {
								comkit.ComRelease(bestMatch)
								return cand, nil
							}
							candName := strings.ToLower(strings.TrimSpace(StripMnemonicAmpersands(node.Name)))
							if targetTitle != "" && candName == targetTitle {
								roleMatched := (wantRole == RoleUnknown && sel.ControlType == "") ||
									(wantRole != RoleUnknown && node.Role == wantRole) ||
									strings.EqualFold(node.ControlTypeName, sel.ControlType) ||
									strings.EqualFold(node.ControlTypeName, sel.ClassName)
								if roleMatched {
									comkit.ComRelease(bestMatch)
									return cand, nil
								}
								if bestMatch == 0 {
									bestMatch = cand
									continue
								}
							}
							if int(node.BoundingRect.Left) == sel.X && int(node.BoundingRect.Top) == sel.Y &&
								node.BoundingRect.Width() == sel.Width && node.BoundingRect.Height() == sel.Height && bestMatch == 0 {
								bestMatch = cand
								continue
							}
							comkit.ComRelease(cand)
						}
						if bestMatch != 0 {
							return bestMatch, nil
						}
					}
				}
			}
		}
	}

	cx, cy := sel.Center()
	if cx > 0 && cy > 0 {
		packedPt := uintptr(uint32(int32(cx))) | (uintptr(uint32(int32(cy))) << 32)
		var ptElem uintptr
		if comkit.VtblCall2(uia.Raw(), 7, packedPt, uintptr(unsafe.Pointer(&ptElem))).Succeeded() && ptElem != 0 {
			return ptElem, nil
		}
	}

	return 0, errors.New("uiakit: matching UIA element not found")
}

func (p *windowsUIAProvider) InvokeOrSelect(sel Selector) (ActionKind, error) {
	unlock, err := comkit.LockThreadMTA()
	if err != nil {
		return "", err
	}
	defer unlock()

	uia, err := p.getAutomation()
	if err != nil {
		return "", err
	}

	target, err := findMatchingElement(uia, sel)
	if err != nil || target == 0 {
		return "", fmt.Errorf("uiakit: target lookup failed: %w", err)
	}
	defer comkit.ComRelease(target)

	var ctrlType int32
	_ = comkit.VtblCall1(target, 21, uintptr(unsafe.Pointer(&ctrlType)))

	if RoleFromWindowsUIA(ctrlType) == RoleCheckBox || sel.EffectiveRole() == RoleCheckBox {
		var togPat uintptr
		if comkit.VtblCall2(target, 16, UIATogglePatternID, uintptr(unsafe.Pointer(&togPat))).Succeeded() &&
			togPat != 0 {
			hr := comkit.VtblCall0(togPat, 3)
			comkit.ComRelease(togPat)
			if hr.Succeeded() {
				p.InvalidateCache()
				return ActionToggle, nil
			}
		}
	}

	var selPat uintptr
	if comkit.VtblCall2(target, 16, UIASelectionItemPatternID, uintptr(unsafe.Pointer(&selPat))).Succeeded() &&
		selPat != 0 {
		hr := comkit.VtblCall0(selPat, 3)
		comkit.ComRelease(selPat)
		if hr.Succeeded() {
			p.InvalidateCache()
			return ActionSelect, nil
		}
	}

	var invPat uintptr
	if comkit.VtblCall2(target, 16, UIAInvokePatternID, uintptr(unsafe.Pointer(&invPat))).Succeeded() && invPat != 0 {
		hr := comkit.VtblCall0(invPat, 3)
		comkit.ComRelease(invPat)
		if hr.Succeeded() {
			p.InvalidateCache()
			return ActionInvoke, nil
		}
	}

	var togPat uintptr
	if comkit.VtblCall2(target, 16, UIATogglePatternID, uintptr(unsafe.Pointer(&togPat))).Succeeded() && togPat != 0 {
		hr := comkit.VtblCall0(togPat, 3)
		comkit.ComRelease(togPat)
		if hr.Succeeded() {
			p.InvalidateCache()
			return ActionToggle, nil
		}
	}

	return "", errors.New("uiakit: element does not support Invoke, SelectionItem, or Toggle UIA patterns")
}

func (p *windowsUIAProvider) SetValue(sel Selector, text string) (ActionKind, error) {
	unlock, err := comkit.LockThreadMTA()
	if err != nil {
		return "", err
	}
	defer unlock()

	uia, err := p.getAutomation()
	if err != nil {
		return "", err
	}

	target, err := findMatchingElement(uia, sel)
	if err != nil || target == 0 {
		return "", err
	}
	defer comkit.ComRelease(target)

	var valPat uintptr
	if comkit.VtblCall2(target, 16, UIAValuePatternID, uintptr(unsafe.Pointer(&valPat))).Failed() || valPat == 0 {
		return "", errors.New("uiakit: element does not support UIA_ValuePattern")
	}
	defer comkit.ComRelease(valPat)

	var readOnly int32
	_ = comkit.VtblCall1(valPat, 5, uintptr(unsafe.Pointer(&readOnly)))
	if readOnly != 0 {
		return "", errors.New("uiakit: UIA_ValuePattern is read-only")
	}

	bstr := comkit.AllocBSTR(text)
	if bstr == 0 && text != "" {
		return "", errors.New("uiakit: AllocBSTR failed")
	}
	defer comkit.FreeBSTR(bstr)

	if hr := comkit.VtblCall1(valPat, 3, uintptr(bstr)); hr.Failed() {
		return "", hr.Err("IUIAutomationValuePattern::SetValue")
	}
	p.InvalidateCache()
	return ActionSetValue, nil
}

func (p *windowsUIAProvider) Scroll(sel Selector, deltaX, deltaY int) (ActionKind, error) {
	unlock, err := comkit.LockThreadMTA()
	if err != nil {
		return "", err
	}
	defer unlock()

	uia, err := p.getAutomation()
	if err != nil {
		return "", err
	}

	target, err := findMatchingElement(uia, sel)
	if err != nil || target == 0 {
		return "", err
	}
	defer comkit.ComRelease(target)

	var scrollPat uintptr
	if comkit.VtblCall2(target, 16, UIAScrollPatternID, uintptr(unsafe.Pointer(&scrollPat))).Failed() ||
		scrollPat == 0 {
		return "", errors.New("uiakit: element does not support UIA_ScrollPattern")
	}
	defer comkit.ComRelease(scrollPat)

	hAmount := ScrollAmountNoAmount
	if deltaX > 0 {
		hAmount = ScrollAmountSmallIncrement
	} else if deltaX < 0 {
		hAmount = ScrollAmountSmallDecrement
	}

	vAmount := ScrollAmountNoAmount
	if deltaY < 0 {
		vAmount = ScrollAmountLargeIncrement
	} else if deltaY > 0 {
		vAmount = ScrollAmountLargeDecrement
	}

	if hr := comkit.VtblCall2(scrollPat, 3, hAmount, vAmount); hr.Failed() {
		return "", hr.Err("IUIAutomationScrollPattern::Scroll")
	}
	p.InvalidateCache()
	return ActionScroll, nil
}
