//go:build windows

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

type previewBitmapInfo struct {
	Size                   uint32
	Width, Height          int32
	Planes, Bits           uint16
	Compression, ImageSize uint32
	X, Y                   int32
	Used, Important        uint32
}

func writeSurfacePNG(s *surface, path string) error {
	pixels := make([]byte, s.W*s.H*4)
	info := previewBitmapInfo{Size: 40, Width: int32(s.W), Height: -int32(s.H), Planes: 1, Bits: 32}
	// GetDIBits requires the bitmap to be deselected from its DC.
	gdi.NewProc("SelectObject").Call(s.DC, s.Old)
	rows, _, err := gdi.NewProc("GetDIBits").Call(s.DC, s.Bitmap, 0, uintptr(s.H), uintptr(unsafe.Pointer(&pixels[0])), uintptr(unsafe.Pointer(&info)), 0)
	gdi.NewProc("SelectObject").Call(s.DC, s.Bitmap)
	if rows != uintptr(s.H) {
		return fmt.Errorf("GetDIBits rows=%d: %v", rows, err)
	}
	img := image.NewRGBA(image.Rect(0, 0, s.W, s.H))
	for i := 0; i < len(pixels); i += 4 {
		img.Pix[i] = pixels[i+2]
		img.Pix[i+1] = pixels[i+1]
		img.Pix[i+2] = pixels[i]
		img.Pix[i+3] = 255
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
func TestNativeShelfPreview(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var token uintptr
	input := gpStartup{Version: 1}
	if status := gp("GdiplusStartup", uintptr(unsafe.Pointer(&token)), uintptr(unsafe.Pointer(&input)), 0); status != 0 {
		t.Fatal(status)
	}
	defer gp("GdiplusShutdown", token)
	dc, _, _ := user32.NewProc("GetDC").Call(0)
	if dc == 0 {
		t.Fatal("GetDC")
	}
	defer user32.NewProc("ReleaseDC").Call(0, dc)
	dir := t.TempDir()
	for name, data := range map[string][]byte{"logo.png": logo, "booktrigger.png": booktriggerLogo, "ivy.png": ivy, "glow.png": glow} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var books []Book
	colors := []color.RGBA{{34, 66, 74, 255}, {110, 46, 42, 255}, {172, 140, 84, 255}, {59, 81, 61, 255}, {64, 63, 105, 255}, {157, 96, 62, 255}, {155, 155, 134, 255}, {45, 70, 104, 255}}
	for n, c := range colors {
		img := image.NewRGBA(image.Rect(0, 0, 160, 234))
		for y := 0; y < 234; y++ {
			for x := 0; x < 160; x++ {
				pixel := c
				if x < 5 || (x > 12 && x < 147 && (y == 12 || y == 221)) || ((x == 12 || x == 147) && y > 12 && y < 222) {
					pixel = color.RGBA{209, 183, 122, 255}
				}
				if y > 55 && y < 61 && x > 35 && x < 125 {
					pixel = color.RGBA{233, 218, 177, 255}
				}
				if y > 75 && y < 79 && x > 48 && x < 112 {
					pixel = color.RGBA{233, 218, 177, 255}
				}
				if (x-80)*(x-80)+(y-145)*(y-145) < 625 {
					pixel = color.RGBA{c.R / 2, c.G / 2, c.B / 2, 255}
				}
				img.SetRGBA(x, y, pixel)
			}
		}
		path := filepath.Join(dir, fmt.Sprintf("book%d.png", n))
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, img)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		books = append(books, Book{Title: fmt.Sprintf("Muestra %d", n), Cover: fmt.Sprintf("https://example.test/%d.png", n), Link: fmt.Sprintf("https://example.test/%d", n), Path: path})
	}
	if err := os.MkdirAll("dist", 0700); err != nil {
		t.Fatal(err)
	}
	for _, cycle := range []int64{0, 2, 6} {
		now := time.Date(2026, 10, 22, 12, 38, 0, 0, time.UTC)
		scene := newScene(1920, 1080, false, 42, now)
		scene.Cycle = cycle
		scene.relayout()
		scene.queue(books, now)
		var graphics saverGraphics
		_, err := graphics.paint(dc, scene, now.Add(time.Second), "Biblioteca online", filepath.Join(dir, "logo.png"), filepath.Join(dir, "glow.png"))
		if err == nil && graphics.cache.get(filepath.Join(dir, "ivy.png")) == nil {
			err = fmt.Errorf("ivy asset failed to decode")
		}
		if err == nil {
			err = writeSurfacePNG(&graphics.front, filepath.Join("dist", fmt.Sprintf("preview-estanteria-%d.png", cycle)))
		}
		graphics.close()
		if err != nil {
			t.Fatal(err)
		}
	}
}
