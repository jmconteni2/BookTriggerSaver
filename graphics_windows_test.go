//go:build windows

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

// Real Windows GDI/GDI+ smoke test. It draws offscreen and neither changes
// screen-saver settings nor opens a browser. CI runs it on windows-latest.
func TestNativeGraphicsOffscreen(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var token uintptr
	input := gpStartup{Version: 1}
	if code := gp("GdiplusStartup", uintptr(unsafe.Pointer(&token)), uintptr(unsafe.Pointer(&input)), 0); code != 0 {
		t.Fatalf("GDI+ startup: %d", code)
	}
	defer gp("GdiplusShutdown", token)
	dc, _, _ := user32.NewProc("GetDC").Call(0)
	if dc == 0 {
		t.Fatal("GetDC")
	}
	defer user32.NewProc("ReleaseDC").Call(0, dc)
	dir := t.TempDir()
	for name, data := range map[string][]byte{"logo.png": logo, "booktrigger.png": booktriggerLogo, "glow.png": glow, "book.png": pngSample()} {
		if e := os.WriteFile(filepath.Join(dir, name), data, 0600); e != nil {
			t.Fatal(e)
		}
	}
	now := time.Now()
	scene := newScene(1920, 1080, false, 7, now)
	books := fixtureBooks()
	for i := range books {
		books[i].Path = filepath.Join(dir, "book.png")
	}
	scene.queue(books, now)
	var g saverGraphics
	defer g.close()
	for _, elapsed := range []time.Duration{0, 250 * time.Millisecond, time.Second, 14750 * time.Millisecond} {
		link, e := g.paint(dc, scene, now.Add(elapsed), "Biblioteca online", filepath.Join(dir, "logo.png"), filepath.Join(dir, "glow.png"))
		if e != nil {
			t.Fatal(e)
		}
		b, _ := scene.current()
		if elapsed == 0 && link != "" {
			t.Fatal("invisible cover exposed a purchase link")
		}
		if elapsed > 0 && link != b.Link {
			t.Fatalf("rendered book/link mismatch: %q != %q", link, b.Link)
		}
		p := panelLayout(scene.W, scene.H, scene.Cycle)
		pixel, _, _ := gdi.NewProc("GetPixel").Call(g.front.DC, uintptr(p.Box.X+20), uintptr(p.Box.Y+p.Box.H/2))
		if uint32(pixel) != 0x23140b {
			t.Fatalf("panel interior: %#x", pixel)
		}
	}
}

func TestPurchaseNativeStructureAndInvalidLink(t *testing.T) {
	if unsafe.Sizeof(shellExecuteInfo{}) != 112 {
		t.Fatal("invalid SHELLEXECUTEINFOW layout")
	}
	if unsafe.Offsetof(shellExecuteInfo{}.Instance) != 56 || unsafe.Offsetof(shellExecuteInfo{}.Process) != 104 {
		t.Fatal("invalid shell offsets")
	}
	if openPurchase("file:///test") == nil {
		t.Fatal("non-web purchase URL accepted")
	}
}

func TestNativeEnterUsesPresentedLinkOnce(t *testing.T) {
	old := activeSaver
	defer func() { activeSaver = old }()
	s := &saverWindow{VisibleLink: "https://example.test/presented"}
	activeSaver = s
	saverWndProc(0, 0x0100, 13, 1<<30)
	if s.PurchaseLink != "" {
		t.Fatal("held Enter triggered purchase")
	}
	saverWndProc(0, 0x0100, 13, 0)
	if s.PurchaseLink != s.VisibleLink {
		t.Fatal("Enter did not capture the presented link")
	}
	s.PurchaseLink = ""
	saverWndProc(0, 0x0100, 65, 0)
	if s.PurchaseLink != "" {
		t.Fatal("other key triggered purchase")
	}
	s.Preview = true
	saverWndProc(0, 0x0100, 13, 0)
	if s.PurchaseLink != "" {
		t.Fatal("preview triggered purchase")
	}
}
