//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type saverWindow struct {
	Scene                                   *Scene
	Store                                   *CatalogStore
	Graphics                                saverGraphics
	HWND                                    uintptr
	Preview                                 bool
	Root, Status, VisibleLink, PurchaseLink string
	Revision                                uint64
	Started                                 time.Time
	Mouse                                   winPoint
	LastMinute                              int
	Dirty                                   bool
	PaintError                              error
}

var activeSaver *saverWindow
var saverCallback = syscall.NewCallback(saverWndProc)

func (s *saverWindow) close() {
	if s.HWND != 0 {
		user32.NewProc("DestroyWindow").Call(s.HWND)
	}
}
func (s *saverWindow) tick() {
	now := time.Now()
	books, revision, status := s.Store.snapshot()
	changed := false
	if revision != s.Revision {
		s.Revision = revision
		if len(books) > 0 {
			s.Scene.queue(books, now)
		}
		if s.Status != status {
			s.Status = status
			changed = true
		}
	}
	if s.Scene.step(now) {
		changed = true
	}
	minute := now.Hour()*60 + now.Minute()
	if minute != s.LastMinute {
		s.LastMinute = minute
		changed = true
	}
	if changed || s.Dirty || s.Scene.animating(now) {
		user32.NewProc("InvalidateRect").Call(s.HWND, 0, 0)
		s.Dirty = false
	}
}
func saverWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	s := activeSaver
	if s == nil {
		r, _, _ := user32.NewProc("DefWindowProcW").Call(hwnd, uintptr(msg), wParam, lParam)
		return r
	}
	switch msg {
	case 0x0014:
		return 1 // WM_ERASEBKGND; the complete frame is double buffered.
	case 0x000f:
		var ps winPaint
		dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		if dc != 0 {
			var rect winRect
			user32.NewProc("GetClientRect").Call(hwnd, uintptr(unsafe.Pointer(&rect)))
			if rect.Right > 0 && rect.Bottom > 0 {
				s.Scene.resize(int(rect.Right), int(rect.Bottom))
				link, e := s.Graphics.paint(dc, s.Scene, time.Now(), s.Status, filepath.Join(s.Root, "logo.png"), filepath.Join(s.Root, "glow.png"))
				if link != "" {
					s.VisibleLink = link
				}
				if e != nil {
					s.PaintError = e
					user32.NewProc("PostMessageW").Call(hwnd, 0x10, 0, 0)
				}
			}
		}
		user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0
	case 0x0113:
		s.tick()
		return 0
	case 0x0005:
		s.Dirty = true
		user32.NewProc("InvalidateRect").Call(hwnd, 0, 0)
		return 0
	case 0x0100, 0x0104:
		// Ignore auto-repeat from the Enter used to start a preview/run.
		if wParam == 13 && lParam&(1<<30) != 0 {
			return 0
		}
		exit, open := keyAction(uint32(wParam), s.Preview)
		if open {
			purchaseTrace("enter", fmt.Sprintf("preview=%t target_valid=%t", s.Preview, webURL(s.VisibleLink)))
		}
		if open && webURL(s.VisibleLink) {
			s.PurchaseLink = s.VisibleLink
		}
		if exit {
			s.close()
			return 0
		}
	case 0x0201, 0x0204, 0x0207, 0x020b, 0x020a, 0x020e:
		if !s.Preview {
			s.close()
			return 0
		}
	case 0x0200:
		if !s.Preview && time.Since(s.Started) <= 700*time.Millisecond {
			user32.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&s.Mouse)))
			return 0
		}
		if !s.Preview && time.Since(s.Started) > 700*time.Millisecond {
			var p winPoint
			if r, _, _ := user32.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&p))); r != 0 {
				dx, dy := int64(p.X)-int64(s.Mouse.X), int64(p.Y)-int64(s.Mouse.Y)
				if dx*dx+dy*dy > 36 {
					s.close()
				}
			}
			return 0
		}
	case 0x0020:
		if !s.Preview {
			user32.NewProc("SetCursor").Call(0)
			return 1
		}
	case 0x0112:
		if wParam&0xfff0 == 0xf140 {
			return 0
		} // Avoid starting a nested screen saver.
	case 0x0010:
		s.close()
		return 0
	case 0x0002:
		user32.NewProc("KillTimer").Call(hwnd, 1)
		s.Graphics.close()
		s.HWND = 0
		user32.NewProc("PostQuitMessage").Call(0)
		return 0
	}
	r, _, _ := user32.NewProc("DefWindowProcW").Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func runSaver(root string, preview bool, parent uintptr) error {
	if e := os.MkdirAll(root, 0700); e != nil {
		return e
	}
	for name, data := range map[string][]byte{"logo.png": logo, "booktrigger.png": booktriggerLogo, "glow.png": glow} {
		if e := atomicWrite(filepath.Join(root, name), data); e != nil {
			return e
		}
	}
	endpoint := strings.TrimSpace(string(defaultEndpoint))
	if b, e := os.ReadFile(filepath.Join(root, "endpoint.txt")); e == nil && len(b) < 4096 {
		endpoint = strings.TrimSpace(string(b))
	}
	if !webURL(endpoint) {
		return errors.New("La dirección del catálogo no es válida")
	}
	var token uintptr
	input := gpStartup{Version: 1}
	if code := gp("GdiplusStartup", uintptr(unsafe.Pointer(&token)), uintptr(unsafe.Pointer(&input)), 0); code != 0 {
		return fmt.Errorf("No se pudo iniciar GDI+ (%d)", code)
	}
	defer gp("GdiplusShutdown", token)
	instance, _, _ := kernel.NewProc("GetModuleHandleW").Call(0)
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	name := wide("BookTriggerSaver.SourceEngine.V11")
	class := winClass{Size: uint32(unsafe.Sizeof(winClass{})), Style: 3, WndProc: saverCallback, Instance: instance, Cursor: cursor, ClassName: name}
	atom, _, e := user32.NewProc("RegisterClassExW").Call(uintptr(unsafe.Pointer(&class)))
	if atom == 0 {
		return fmt.Errorf("No se pudo registrar la ventana: %v", e)
	}
	defer user32.NewProc("UnregisterClassW").Call(uintptr(unsafe.Pointer(name)), instance)
	w, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
	h, _, _ := user32.NewProc("GetSystemMetrics").Call(1)
	style, exStyle := uintptr(0x80000000), uintptr(0x88)
	if preview {
		var r winRect
		valid, _, _ := user32.NewProc("IsWindow").Call(parent)
		if valid == 0 {
			return nil
		}
		if valid, _, _ := user32.NewProc("GetClientRect").Call(parent, uintptr(unsafe.Pointer(&r))); valid == 0 {
			return nil
		}
		if r.Right <= 0 || r.Bottom <= 0 {
			return nil
		}
		w, h = uintptr(r.Right), uintptr(r.Bottom)
		style = 0x50000000
		exStyle = 0
	}
	now := time.Now()
	state := &saverWindow{Scene: newScene(int(w), int(h), preview, now.UnixNano(), now), Store: newCatalogStore(root, endpoint), Root: root, Preview: preview, Started: now, Status: "Conectando con la biblioteca…", LastMinute: -1, Dirty: true}
	user32.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&state.Mouse)))
	activeSaver = state
	defer func() { activeSaver = nil; state.Graphics.close() }()
	hwnd, _, e := user32.NewProc("CreateWindowExW").Call(exStyle, uintptr(unsafe.Pointer(name)), ptr("BookTrigger — J.M. Pocket"), style, 0, 0, w, h, parent, 0, instance, 0)
	if hwnd == 0 {
		return fmt.Errorf("No se pudo crear el salvapantallas: %v", e)
	}
	state.HWND = hwnd
	defer func() {
		if state.HWND != 0 {
			state.close()
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go state.Store.run(ctx)
	if r, _, _ := user32.NewProc("SetTimer").Call(hwnd, 1, 33, 0); r == 0 {
		return errors.New("No se pudo iniciar el temporizador del salvapantallas")
	}
	user32.NewProc("ShowWindow").Call(hwnd, 5)
	if !preview {
		user32.NewProc("SetForegroundWindow").Call(hwnd)
		user32.NewProc("SetFocus").Call(hwnd)
	}
	user32.NewProc("UpdateWindow").Call(hwnd)
	var msg winMessage
	for {
		result, _, e := user32.NewProc("GetMessageW").Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(result) == -1 {
			return fmt.Errorf("El bucle de ventanas ha fallado: %v", e)
		}
		if result == 0 {
			break
		}
		user32.NewProc("TranslateMessage").Call(uintptr(unsafe.Pointer(&msg)))
		user32.NewProc("DispatchMessageW").Call(uintptr(unsafe.Pointer(&msg)))
	}
	if state.PaintError != nil {
		return state.PaintError
	}
	if webURL(state.PurchaseLink) {
		if err := handoffPurchase(state.PurchaseLink); err != nil {
			user32.NewProc("MessageBoxW").Call(0, ptr(err.Error()), ptr("BookTrigger — No se pudo abrir la compra"), 0x10)
			return err
		}
	}
	return nil
}
