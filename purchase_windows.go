//go:build windows

package main

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

// Native x64 SHELLEXECUTEINFOW. NOASYNC waits for browser activation before
// this process exits, including when the default browser is already running.
type shellExecuteInfo struct {
	Size, Mask                        uint32
	HWND                              uintptr
	Verb, File, Parameters, Directory *uint16
	Show                              int32
	Instance                          uintptr
	IDList                            uintptr
	Class                             *uint16
	ClassKey                          uintptr
	HotKey                            uint32
	Icon, Process                     uintptr
}

func openPurchase(link string) error {
	if !webURL(link) {
		return fmt.Errorf("El enlace de compra no es válido")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ole := syscall.NewLazyDLL("ole32.dll")
	hr, _, _ := ole.NewProc("CoInitializeEx").Call(0, 6)
	if int32(hr) >= 0 {
		defer ole.NewProc("CoUninitialize").Call()
	} else if uint32(hr) != 0x80010106 {
		return fmt.Errorf("No se pudo preparar la apertura del navegador (0x%x)", uint32(hr))
	}
	info := shellExecuteInfo{Mask: 0x100 | 0x400, Verb: wide("open"), File: wide(link), Show: 1}
	info.Size = uint32(unsafe.Sizeof(info))
	ok, _, err := syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteExW").Call(uintptr(unsafe.Pointer(&info)))
	runtime.KeepAlive(info)
	if ok == 0 {
		return fmt.Errorf("Windows no pudo abrir el enlace de compra: %v", err)
	}
	return nil
}
