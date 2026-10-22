//go:build windows

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/png"
	"runtime"
	"syscall"
	"unsafe"
)

// The selected poster is embedded without editing or re-encoding it.
//
//go:embed assets/bienvenida.png
var welcomePNG []byte

type winRect struct{ Left, Top, Right, Bottom int32 }
type winPoint struct{ X, Y int32 }
type winMessage struct {
	Hwnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Point          winPoint
	Private        uint32
}
type winClass struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClassExtra, WindowExtra            int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	SmallIcon                          uintptr
}
type winPaint struct {
	DC                 uintptr
	Erase              int32
	Rect               winRect
	Restore, IncUpdate int32
	Reserved           [32]byte
}
type bitmapHeader struct {
	Size                   uint32
	Width, Height          int32
	Planes, BitCount       uint16
	Compression, SizeImage uint32
	XPels, YPels           int32
	Used, Important        uint32
}

// Windows SDK structure sizes for the supported x64 build.
var _ [80 - unsafe.Sizeof(winClass{})]byte
var _ [unsafe.Sizeof(winClass{}) - 80]byte
var _ [48 - unsafe.Sizeof(winMessage{})]byte
var _ [unsafe.Sizeof(winMessage{}) - 48]byte
var _ [72 - unsafe.Sizeof(winPaint{})]byte
var _ [unsafe.Sizeof(winPaint{}) - 72]byte
var _ [40 - unsafe.Sizeof(bitmapHeader{})]byte
var _ [unsafe.Sizeof(bitmapHeader{}) - 40]byte

type completionState struct {
	layout welcomeLayout
	pixels []byte
	bitmap bitmapHeader
	run    bool
}

var completion *completionState
var gdi = syscall.NewLazyDLL("gdi32.dll")
var completionProc = syscall.NewCallback(completionWndProc)

func completionWndProc(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case 0xf: // WM_PAINT
		var paint winPaint
		dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&paint)))
		if completion != nil && dc != 0 {
			s := completion
			l := s.layout
			// StretchDIBits consumes the decoded bitmap; the embedded PNG stays intact.
			gdi.NewProc("SetStretchBltMode").Call(dc, 4)
			gdi.NewProc("SetBrushOrgEx").Call(dc, 0, 0, 0)
			gdi.NewProc("StretchDIBits").Call(dc, uintptr(l.Margin), uintptr(l.PosterY), uintptr(l.PosterWidth), uintptr(l.PosterHeight), 0, 0, 751, 475, uintptr(unsafe.Pointer(&s.pixels[0])), uintptr(unsafe.Pointer(&s.bitmap)), 0, 0xcc0020)
			runtime.KeepAlive(s)
		}
		user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&paint)))
		return 0
	case 0x111: // WM_COMMAND
		switch wparam & 0xffff {
		case 100:
			completion.run = true
			user32.NewProc("DestroyWindow").Call(hwnd)
			return 0
		case 2:
			user32.NewProc("DestroyWindow").Call(hwnd)
			return 0
		}
	case 0x10: // WM_CLOSE
		user32.NewProc("DestroyWindow").Call(hwnd)
		return 0
	case 2: // WM_DESTROY
		user32.NewProc("PostQuitMessage").Call(0)
		return 0
	}
	r, _, _ := user32.NewProc("DefWindowProcW").Call(hwnd, uintptr(msg), wparam, lparam)
	return r
}

func completionWindow(timeout uint32) (bool, error) {
	img, e := png.Decode(bytes.NewReader(welcomePNG))
	if e != nil {
		return false, e
	}
	if img.Bounds().Dx() != 751 || img.Bounds().Dy() != 475 {
		return false, fmt.Errorf("Dimensiones del cartel inesperadas")
	}
	pixels := make([]byte, 751*475*4)
	for y := 0; y < 475; y++ {
		for x := 0; x < 751; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			i := (y*751 + x) * 4
			pixels[i] = byte(b >> 8)
			pixels[i+1] = byte(g >> 8)
			pixels[i+2] = byte(r >> 8)
			pixels[i+3] = 255
		}
	}
	dpi := uintptr(96)
	if p := user32.NewProc("GetDpiForSystem"); p.Find() == nil {
		if d, _, _ := p.Call(); d != 0 {
			dpi = d
		}
	}
	scale := func(v int) int { return (v*int(dpi) + 48) / 96 }
	var work winRect
	r, _, _ := spi.Call(0x30, 0, uintptr(unsafe.Pointer(&work)), 0)
	if r == 0 {
		w, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
		h, _, _ := user32.NewProc("GetSystemMetrics").Call(1)
		work.Right = int32(w)
		work.Bottom = int32(h)
	}
	l := completionLayout(int(work.Right-work.Left), int(work.Bottom-work.Top), int(dpi))
	if l.PosterHeight < 100 || l.Width < scale(340) {
		return false, fmt.Errorf("Espacio insuficiente para mostrar el cartel")
	}
	completion = &completionState{layout: l, pixels: pixels, bitmap: bitmapHeader{Size: 40, Width: 751, Height: -475, Planes: 1, BitCount: 32}}
	defer func() { completion = nil }()
	instance, _, _ := kernel.NewProc("GetModuleHandleW").Call(0)
	icon, _, _ := user32.NewProc("LoadIconW").Call(instance, 1)
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	class := winClass{WndProc: completionProc, Instance: instance, Icon: icon, SmallIcon: icon, Cursor: cursor, Background: 6, ClassName: wide("BookTriggerInstallationComplete")}
	class.Size = uint32(unsafe.Sizeof(class))
	atom, _, err := user32.NewProc("RegisterClassExW").Call(uintptr(unsafe.Pointer(&class)))
	if atom == 0 && err != syscall.Errno(1410) {
		return false, fmt.Errorf("No se pudo crear la pantalla final: %v", err)
	}
	const style = 0x00ca0000
	const exStyle = 0x00010001
	rect := winRect{Right: int32(l.Width), Bottom: int32(l.Height)}
	if p := user32.NewProc("AdjustWindowRectExForDpi"); p.Find() == nil {
		p.Call(uintptr(unsafe.Pointer(&rect)), style, 0, exStyle, dpi)
	} else {
		user32.NewProc("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(&rect)), style, 0, exStyle)
	}
	ww, wh := rect.Right-rect.Left, rect.Bottom-rect.Top
	x, y := work.Left+(work.Right-work.Left-ww)/2, work.Top+(work.Bottom-work.Top-wh)/2
	hwnd, _, err := user32.NewProc("CreateWindowExW").Call(exStyle, uintptr(unsafe.Pointer(class.ClassName)), ptr("BookTrigger — Instalación completada"), style, uintptr(x), uintptr(y), uintptr(ww), uintptr(wh), 0, 0, instance, 0)
	if hwnd == 0 {
		return false, fmt.Errorf("No se pudo abrir la pantalla final: %v", err)
	}
	var fonts []uintptr
	defer func() {
		for _, f := range fonts {
			gdi.NewProc("DeleteObject").Call(f)
		}
	}()
	makeFont := func(height, weight int) uintptr {
		h := int32(-scale(height))
		font, _, _ := gdi.NewProc("CreateFontW").Call(uintptr(h), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, ptr("Segoe UI"))
		if font != 0 {
			fonts = append(fonts, font)
		}
		return font
	}
	headingFont, bodyFont, smallFont := makeFont(23, 600), makeFont(14, 400), makeFont(12, 400)
	control := func(kind, text string, style uintptr, x, y, w, h, id int, font uintptr) uintptr {
		c, _, _ := user32.NewProc("CreateWindowExW").Call(0, ptr(kind), ptr(text), style|0x50000000, uintptr(x), uintptr(y), uintptr(w), uintptr(h), hwnd, uintptr(id), instance, 0)
		if c != 0 && font != 0 {
			user32.NewProc("SendMessageW").Call(c, 0x30, font, 1)
		}
		return c
	}
	control("STATIC", "Instalación completada", 0, l.Margin, scale(12), l.PosterWidth, scale(32), 0, headingFont)
	status := fmt.Sprintf("Activación automática: %d minutos de inactividad.", timeout/60)
	if timeout == 60 {
		status = "Activación automática: 1 minuto de inactividad."
	}
	control("STATIC", status, 0, l.Margin, scale(49), l.PosterWidth, scale(28), 0, bodyFont)
	control("STATIC", "Libro destacado: 15 s. Caja del logo y la hora: cambio cada 30 s.", 0, l.Margin, l.FooterY+scale(11), l.PosterWidth, scale(22), 0, smallFont)
	closeW, runW, gap := scale(94), scale(144), scale(10)
	buttonY := l.Height - scale(43)
	closeButton := control("BUTTON", "Cerrar", 0x10000, l.Width-l.Margin-closeW, buttonY, closeW, scale(31), 2, bodyFont)
	runButton := control("BUTTON", "Ejecutar ahora", 0x10001, l.Width-l.Margin-closeW-gap-runW, buttonY, runW, scale(31), 100, bodyFont)
	user32.NewProc("ShowWindow").Call(hwnd, 5)
	user32.NewProc("UpdateWindow").Call(hwnd)
	user32.NewProc("SetForegroundWindow").Call(hwnd)
	user32.NewProc("SetFocus").Call(runButton)
	var message winMessage
	for {
		result, _, err := user32.NewProc("GetMessageW").Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if result == 0 {
			break
		}
		if int32(result) == -1 {
			user32.NewProc("DestroyWindow").Call(hwnd)
			return false, err
		}
		if message.Message == 0x100 && (message.WParam == 13 || message.WParam == 27) {
			focus, _, _ := user32.NewProc("GetFocus").Call()
			completion.run = message.WParam == 13 && focus != closeButton
			user32.NewProc("DestroyWindow").Call(hwnd)
			continue
		}
		handled, _, _ := user32.NewProc("IsDialogMessageW").Call(hwnd, uintptr(unsafe.Pointer(&message)))
		if handled == 0 {
			user32.NewProc("TranslateMessage").Call(uintptr(unsafe.Pointer(&message)))
			user32.NewProc("DispatchMessageW").Call(uintptr(unsafe.Pointer(&message)))
		}
	}
	return completion.run, nil
}

func completedPage(timeout uint32) bool {
	run, e := completionWindow(timeout)
	if e == nil {
		return run
	}
	logError(e)
	return dialog("BookTrigger", "Instalación completada", welcomeText, "Ejecutar ahora")
}
