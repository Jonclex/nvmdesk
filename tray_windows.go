//go:build windows

package main

import (
	"fmt"
	"sort"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	wmApp           = 0x8000
	wmSize          = 0x0005
	wmCommand       = 0x0111
	wmDestroy       = 0x0002
	wmTrayCallback  = wmApp + 1
	wmRButtonUp     = 0x0205
	wmLButtonUp     = 0x0202
	wmLButtonDblClk = 0x0203
	sizeMinimized   = 1
	gwlpWndProc     = ^uintptr(3)
	wmGetIcon       = 0x007F
	iconSmall       = 0
	iconBig         = 1
	gclpHIcon       = ^uintptr(13)
	gclpHIconSm     = ^uintptr(33)
	nimAdd          = 0x00000000
	nimModify       = 0x00000001
	nimDelete       = 0x00000002
	nifMessage      = 0x00000001
	nifIcon         = 0x00000002
	nifTip          = 0x00000004
	nifInfo         = 0x00000010
	mfString        = 0x00000000
	mfSeparator     = 0x00000800
	mfChecked       = 0x00000008
	mfGrayed        = 0x00000001
	mfPopup         = 0x00000010
	tpmLeftAlign    = 0x0000
	tpmBottomAlign  = 0x0020
	tpmRightButton  = 0x0002
	swHide          = 0
	swRestore       = 9
	cmdShowWindow   = 1001
	cmdExitApp      = 1002
	cmdRefreshList  = 1003
	cmdVersionBase  = 2000
	trayTooltip     = "NVM Desktop Manager"
	niifNone        = 0x00000000
	niifInfo        = 0x00000001
	niifError       = 0x00000003
)

var (
	user32                = windows.NewLazySystemDLL("user32.dll")
	shell32               = windows.NewLazySystemDLL("shell32.dll")
	procFindWindowW       = user32.NewProc("FindWindowW")
	procSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW   = user32.NewProc("CallWindowProcW")
	procShowWindow        = user32.NewProc("ShowWindow")
	procSetForegroundWnd  = user32.NewProc("SetForegroundWindow")
	procCreatePopupMenu   = user32.NewProc("CreatePopupMenu")
	procAppendMenuW       = user32.NewProc("AppendMenuW")
	procTrackPopupMenu    = user32.NewProc("TrackPopupMenu")
	procDestroyMenu       = user32.NewProc("DestroyMenu")
	procGetCursorPos      = user32.NewProc("GetCursorPos")
	procPostMessageW      = user32.NewProc("PostMessageW")
	procSendMessageW      = user32.NewProc("SendMessageW")
	procGetClassLongPtrW  = user32.NewProc("GetClassLongPtrW")
	procShellNotifyIconW  = shell32.NewProc("Shell_NotifyIconW")
	globalTrayProc        = syscall.NewCallback(trayWndProc)
	activeTray            *windowsTray
)

type trayController interface {
	Init(app *App)
	Dispose()
}

type windowsTray struct {
	app          *App
	hwnd         windows.Handle
	originalProc uintptr
	iconAdded    bool
	versionMap   map[uintptr]string
}

type point struct {
	X int32
	Y int32
}

type notifyIconData struct {
	CbSize            uint32
	HWnd              windows.Handle
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             windows.Handle
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          windows.GUID
	HBalloonIcon      windows.Handle
}

func newTrayController() trayController {
	return &windowsTray{
		versionMap: make(map[uintptr]string),
	}
}

func (t *windowsTray) Init(app *App) {
	t.app = app
	go t.waitAndInstall()
}

func (t *windowsTray) Dispose() {
	if t.hwnd == 0 || !t.iconAdded {
		return
	}

	nid := t.newNotifyIconData()
	shellNotifyIcon(nimDelete, &nid)
	t.iconAdded = false
}

func (t *windowsTray) waitAndInstall() {
	for i := 0; i < 50; i++ {
		hwnd, err := findWindowByTitle("NVM Desktop Manager")
		if err == nil && hwnd != 0 {
			t.hwnd = hwnd
			if err := t.install(); err == nil {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (t *windowsTray) install() error {
	if t.hwnd == 0 {
		return fmt.Errorf("window handle not found")
	}

	prev, _, err := procSetWindowLongPtrW.Call(uintptr(t.hwnd), uintptr(gwlpWndProc), globalTrayProc)
	if prev == 0 && err != windows.ERROR_SUCCESS {
		return err
	}
	t.originalProc = prev
	activeTray = t

	nid := t.newNotifyIconData()
	if shellNotifyIcon(nimAdd, &nid) == 0 {
		return fmt.Errorf("add tray icon failed")
	}

	t.iconAdded = true
	return nil
}

func (t *windowsTray) newNotifyIconData() notifyIconData {
	nid := notifyIconData{
		CbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:             t.hwnd,
		UID:              1,
		UFlags:           nifMessage | nifIcon | nifTip,
		UCallbackMessage: wmTrayCallback,
		HIcon:            t.loadIconHandle(),
	}
	copy(nid.SzTip[:], windows.StringToUTF16(trayTooltip))
	return nid
}

func (t *windowsTray) loadIconHandle() windows.Handle {
	icon, _, _ := procSendMessageW.Call(uintptr(t.hwnd), uintptr(wmGetIcon), uintptr(iconSmall), 0)
	if icon == 0 {
		icon, _, _ = procSendMessageW.Call(uintptr(t.hwnd), uintptr(wmGetIcon), uintptr(iconBig), 0)
	}
	if icon == 0 {
		icon, _, _ = procGetClassLongPtrW.Call(uintptr(t.hwnd), uintptr(gclpHIconSm))
	}
	if icon == 0 {
		icon, _, _ = procGetClassLongPtrW.Call(uintptr(t.hwnd), uintptr(gclpHIcon))
	}
	return windows.Handle(icon)
}

func (t *windowsTray) handleWindowMessage(hwnd windows.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmSize:
		if wParam == sizeMinimized {
			showWindow(hwnd, swHide)
			return 0
		}
	case wmTrayCallback:
		switch uint32(lParam) {
		case wmLButtonUp, wmLButtonDblClk:
			t.app.showMainWindow()
			return 0
		case wmRButtonUp:
			t.showContextMenu()
			return 0
		}
	case wmCommand:
		switch loword(wParam) {
		case cmdShowWindow:
			t.app.showMainWindow()
			return 0
		case cmdExitApp:
			t.app.quitFromTray()
			return 0
		case cmdRefreshList:
			go t.refreshVersionList()
			return 0
		}
		if version, ok := t.versionMap[loword(wParam)]; ok {
			go t.switchVersion(version)
			return 0
		}
	case wmDestroy:
		t.Dispose()
	}

	return callWindowProc(t.originalProc, hwnd, msg, wParam, lParam)
}

func (t *windowsTray) showContextMenu() {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)
	t.versionMap = make(map[uintptr]string)

	showText, _ := windows.UTF16PtrFromString("显示主窗口")
	refreshText, _ := windows.UTF16PtrFromString("刷新版本列表")
	exitText, _ := windows.UTF16PtrFromString("退出")
	currentLabel, versionItems := t.buildVersionMenuItems()
	switchMenu, _, _ := procCreatePopupMenu.Call()

	procAppendMenuW.Call(menu, uintptr(mfString), uintptr(cmdShowWindow), uintptr(unsafe.Pointer(showText)))
	if switchMenu != 0 {
		if currentLabel != "" {
			currentText, _ := windows.UTF16PtrFromString(currentLabel)
			procAppendMenuW.Call(switchMenu, uintptr(mfString|mfGrayed), 0, uintptr(unsafe.Pointer(currentText)))
			if len(versionItems) > 0 {
				procAppendMenuW.Call(switchMenu, uintptr(mfSeparator), 0, 0)
			}
		}
		if len(versionItems) == 0 {
			emptyText, _ := windows.UTF16PtrFromString("未检测到已安装版本")
			procAppendMenuW.Call(switchMenu, uintptr(mfString|mfGrayed), 0, uintptr(unsafe.Pointer(emptyText)))
		} else {
			for _, item := range versionItems {
				flags := uintptr(mfString)
				if item.Checked {
					flags |= mfChecked
				}
				t.versionMap[item.CommandID] = item.Version
				label, _ := windows.UTF16PtrFromString(item.Label)
				procAppendMenuW.Call(switchMenu, flags, item.CommandID, uintptr(unsafe.Pointer(label)))
			}
		}
		switchText, _ := windows.UTF16PtrFromString("切换 Node.js")
		procAppendMenuW.Call(menu, uintptr(mfPopup), switchMenu, uintptr(unsafe.Pointer(switchText)))
	}
	procAppendMenuW.Call(menu, uintptr(mfString), uintptr(cmdRefreshList), uintptr(unsafe.Pointer(refreshText)))
	procAppendMenuW.Call(menu, uintptr(mfSeparator), 0, 0)
	procAppendMenuW.Call(menu, uintptr(mfString), uintptr(cmdExitApp), uintptr(unsafe.Pointer(exitText)))

	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWnd.Call(uintptr(t.hwnd))
	procTrackPopupMenu.Call(
		menu,
		uintptr(tpmLeftAlign|tpmBottomAlign|tpmRightButton),
		uintptr(pt.X),
		uintptr(pt.Y),
		0,
		uintptr(t.hwnd),
		0,
	)
	procPostMessageW.Call(uintptr(t.hwnd), uintptr(0), 0, 0)
}

func trayWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	if activeTray != nil && windows.Handle(hwnd) == activeTray.hwnd {
		return activeTray.handleWindowMessage(activeTray.hwnd, msg, wParam, lParam)
	}
	return 0
}

func findWindowByTitle(title string) (windows.Handle, error) {
	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return 0, err
	}

	hwnd, _, callErr := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		if callErr != windows.ERROR_SUCCESS {
			return 0, callErr
		}
		return 0, fmt.Errorf("window not found")
	}
	return windows.Handle(hwnd), nil
}

func shellNotifyIcon(message uint32, data *notifyIconData) uintptr {
	ret, _, _ := procShellNotifyIconW.Call(uintptr(message), uintptr(unsafe.Pointer(data)))
	return ret
}

func callWindowProc(prev uintptr, hwnd windows.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procCallWindowProcW.Call(prev, uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func showWindow(hwnd windows.Handle, cmd int32) {
	procShowWindow.Call(uintptr(hwnd), uintptr(cmd))
}

func loword(v uintptr) uintptr {
	return v & 0xffff
}

type trayVersionItem struct {
	CommandID uintptr
	Version   string
	Label     string
	Checked   bool
}

func (t *windowsTray) buildVersionMenuItems() (string, []trayVersionItem) {
	versions, err := t.app.nvmService.ListInstalled()
	if err != nil {
		return "Node.js: 获取版本失败", nil
	}
	if len(versions) == 0 {
		return "", nil
	}

	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i].Version, versions[j].Version)
	})

	current := ""
	for _, version := range versions {
		if version.IsCurrent {
			current = version.Version
			break
		}
	}
	if current == "" {
		if info, infoErr := t.app.nvmService.GetCurrent(); infoErr == nil {
			current = info.NodeVersion
		}
	}

	items := make([]trayVersionItem, 0, len(versions))
	for index, version := range versions {
		commandID := uintptr(cmdVersionBase + index)
		items = append(items, trayVersionItem{
			CommandID: commandID,
			Version:   version.Version,
			Label:     "切换到 Node.js " + version.Version,
			Checked:   version.Version == current,
		})
	}

	if current == "" {
		return "Node.js: 未知", items
	}
	return "当前 Node.js: " + current, items
}

func (t *windowsTray) switchVersion(version string) {
	if err := t.app.nvmService.Use(version); err != nil {
		t.showBalloon("切换 Node.js 失败", err.Error(), niifError)
		return
	}
	t.showBalloon("切换 Node.js 成功", "已切换到 Node.js "+version, niifInfo)
	t.app.refreshFrontend()
}

func (t *windowsTray) refreshVersionList() {
	versions, err := t.app.nvmService.ListInstalled()
	if err != nil {
		t.showBalloon("刷新版本列表失败", err.Error(), niifError)
		return
	}

	message := "未检测到已安装版本"
	if len(versions) > 0 {
		message = fmt.Sprintf("已检测到 %d 个已安装版本", len(versions))
	}

	t.showBalloon("版本列表已刷新", message, niifInfo)
	t.app.refreshFrontend()
}

func (t *windowsTray) showBalloon(title, message string, iconFlag uint32) {
	if t.hwnd == 0 || !t.iconAdded {
		return
	}

	nid := t.newNotifyIconData()
	nid.UFlags = nifInfo
	nid.DwInfoFlags = iconFlag
	copy(nid.SzInfoTitle[:], windows.StringToUTF16(title))
	copy(nid.SzInfo[:], windows.StringToUTF16(message))
	shellNotifyIcon(nimModify, &nid)
}
