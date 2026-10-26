//go:build windows && capture

package main

import (
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "time"
    "unsafe"
)

var captureOutput string
var captureStarted time.Time
var captureCount int
var captureDelays = [...]time.Duration{4*time.Second, 19*time.Second, 34*time.Second, 49*time.Second}

func captureLauncher() bool {
    if len(os.Args) < 2 || os.Args[1] != "--capturar" { return false }
    home, err := os.UserHomeDir()
    if err != nil { os.Exit(2) }
    captureOutput = filepath.Join(home, "Pictures", "Capturas BookTriggerSaver")
    if err := os.MkdirAll(captureOutput, 0700); err != nil { os.Exit(2) }
    for i := 1; i <= len(captureDelays); i++ {
        _ = os.Remove(filepath.Join(captureOutput, fmt.Sprintf("BookTriggerSaver-%02d.png", i)))
    }
    _ = os.Remove(filepath.Join(captureOutput, "ERROR.txt"))
    local := os.Getenv("LOCALAPPDATA")
    if local == "" {
        captureFailure(fmt.Errorf("La carpeta local de Windows no esta disponible"))
    }
    err = runSaver(filepath.Join(local, "BookTriggerSaverCapturas"), false, 0)
    if err != nil { captureFailure(err) }
    if captureCount != len(captureDelays) { captureFailure(fmt.Errorf("Solo se guardaron %d de %d capturas", captureCount, len(captureDelays))) }
    return true
}

func captureFailure(err error) {
    if captureOutput != "" {
        _ = os.WriteFile(filepath.Join(captureOutput,"ERROR.txt"), []byte(err.Error()+"\r\n"), 0600)
    }
    os.Exit(1)
}

func captureWantsFrame(s *saverWindow) bool {
    if captureOutput == "" || s.Preview || len(s.Scene.Books) == 0 || captureCount >= len(captureDelays) { return false }
    if captureStarted.IsZero() {
        captureStarted = time.Now()
        return false
    }
    return time.Since(captureStarted) >= captureDelays[captureCount]
}

func captureAfterPaint(s *saverWindow) (bool, error) {
    if !captureWantsFrame(s) { return false, nil }
    filename := filepath.Join(captureOutput,fmt.Sprintf("BookTriggerSaver-%02d.png",captureCount+1))
    if err := saveBitmapPNG(s.Graphics.front.Bitmap, filename); err != nil { return false, err }
    captureCount++
    return captureCount == len(captureDelays), nil
}

type pngGuid struct {
    Data1 uint32
    Data2 uint16
    Data3 uint16
    Data4 [8]byte
}

func saveBitmapPNG(hbitmap uintptr, filename string) error {
    if hbitmap == 0 { return fmt.Errorf("La imagen de pantalla no esta disponible") }
    var bitmap uintptr
    if status := gp("GdipCreateBitmapFromHBITMAP", hbitmap, 0, uintptr(unsafe.Pointer(&bitmap))); status != 0 || bitmap == 0 {
        return fmt.Errorf("No se pudo leer el fotograma (%d)", status)
    }
    defer gp("GdipDisposeImage",bitmap)
    png := pngGuid{0x557cf406, 0x1a04, 0x11d3, [8]byte{0x9a,0x73,0x00,0x00,0xf8,0x1e,0xf3,0x2e}}
    status := gp("GdipSaveImageToFile",bitmap,ptr(filename),uintptr(unsafe.Pointer(&png)),0)
    runtime.KeepAlive(png)
    if status != 0 { return fmt.Errorf("No se pudo guardar el PNG (%d)", status) }
    return nil
}
