//go:build windows

package main

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

// Explorer owns the desktop automation object. Delegate to that existing
// process; never spawn a browser/helper from the screen-saver desktop.
// Based on Microsoft's GetDesktopAutomationObject / ShellExecuteFromExplorer.
type automationVariant struct {
	Type         uint16
	Reserved     [3]uint16
	Value, Extra uint64
}
type automationParams struct {
	Args              *automationVariant
	Named             *int32
	Count, NamedCount uint32
}

var oleAutomation = syscall.NewLazyDLL("oleaut32.dll")
var oleCOM = syscall.NewLazyDLL("ole32.dll")

func comCall(object uintptr, slot int, args ...uintptr) uint32 {
	// Native COM pointers refer to reference-counted memory owned by Windows.
	instance := *(*unsafe.Pointer)(unsafe.Pointer(&object))
	table := *(*unsafe.Pointer)(instance)
	method := (*[32]uintptr)(table)[slot]
	parameters := append([]uintptr{object}, args...)
	hr, _, _ := syscall.SyscallN(method, parameters...)
	return uint32(hr)
}
func comRelease(object uintptr) {
	if object != 0 {
		comCall(object, 2)
	}
}
func comError(stage string, hr uint32) error {
	return fmt.Errorf("%s: Windows HRESULT 0x%08x", stage, hr)
}
func comGUID(s string) (syscall.GUID, error) {
	var result syscall.GUID
	hr, _, _ := oleCOM.NewProc("CLSIDFromString").Call(ptr(s), uintptr(unsafe.Pointer(&result)))
	if int32(hr) < 0 {
		return result, comError("Identificador COM", uint32(hr))
	}
	return result, nil
}
func comQuery(object uintptr, iid syscall.GUID) (uintptr, error) {
	var out uintptr
	hr := comCall(object, 0, uintptr(unsafe.Pointer(&iid)), uintptr(unsafe.Pointer(&out)))
	if int32(hr) < 0 || out == 0 {
		return 0, comError("Interfaz del Explorador", hr)
	}
	return out, nil
}
func dispatchInvoke(object uintptr, name string, flags uint16, args []automationVariant) (automationVariant, error) {
	var null syscall.GUID
	member := wide(name)
	var id int32
	hr := comCall(object, 5, uintptr(unsafe.Pointer(&null)), uintptr(unsafe.Pointer(&member)), 1, 0, uintptr(unsafe.Pointer(&id)))
	if int32(hr) < 0 {
		return automationVariant{}, comError("Explorador: "+name, hr)
	}
	params := automationParams{Count: uint32(len(args))}
	if len(args) > 0 {
		params.Args = &args[0]
	}
	var result automationVariant
	var argument uint32
	hr = comCall(object, 6, uintptr(uint32(id)), uintptr(unsafe.Pointer(&null)), 0, uintptr(flags), uintptr(unsafe.Pointer(&params)), uintptr(unsafe.Pointer(&result)), 0, uintptr(unsafe.Pointer(&argument)))
	runtime.KeepAlive(args)
	runtime.KeepAlive(member)
	if int32(hr) < 0 {
		oleAutomation.NewProc("VariantClear").Call(uintptr(unsafe.Pointer(&result)))
		return automationVariant{}, comError("Explorador: "+name, hr)
	}
	return result, nil
}
func withExplorerShell(action func(uintptr) error) error {
	clsid, err := comGUID("{9BA05972-F6A8-11CF-A442-00A0C90A8F39}")
	if err != nil {
		return err
	}
	iid, err := comGUID("{85CB6900-4D95-11CF-960C-0080C7F4EE85}")
	if err != nil {
		return err
	}
	var windows uintptr
	hr, _, _ := oleCOM.NewProc("CoCreateInstance").Call(uintptr(unsafe.Pointer(&clsid)), 0, 4, uintptr(unsafe.Pointer(&iid)), uintptr(unsafe.Pointer(&windows)))
	if int32(hr) < 0 || windows == 0 {
		return comError("Conectar con el Explorador", uint32(hr))
	}
	defer comRelease(windows)
	var empty automationVariant
	var hwnd int32
	var desktop uintptr
	// IShellWindows::FindWindowSW, SWC_DESKTOP=8, SWFO_NEEDDISPATCH=1.
	status := comCall(windows, 15, uintptr(unsafe.Pointer(&empty)), uintptr(unsafe.Pointer(&empty)), 8, uintptr(unsafe.Pointer(&hwnd)), 1, uintptr(unsafe.Pointer(&desktop)))
	if status != 0 || desktop == 0 {
		return comError("Localizar escritorio del Explorador", status)
	}
	defer comRelease(desktop)
	providerID, _ := comGUID("{6D5140C1-7436-11CE-8034-00AA006009FA}")
	provider, err := comQuery(desktop, providerID)
	if err != nil {
		return err
	}
	defer comRelease(provider)
	serviceID, _ := comGUID("{4C96BE40-915C-11CF-99D3-00AA004AE837}")
	browserID, _ := comGUID("{000214E2-0000-0000-C000-000000000046}")
	var browser uintptr
	status = comCall(provider, 3, uintptr(unsafe.Pointer(&serviceID)), uintptr(unsafe.Pointer(&browserID)), uintptr(unsafe.Pointer(&browser)))
	if int32(status) < 0 || browser == 0 {
		return comError("Servicio del escritorio", status)
	}
	defer comRelease(browser)
	var view uintptr
	status = comCall(browser, 15, uintptr(unsafe.Pointer(&view))) // QueryActiveShellView.
	if int32(status) < 0 || view == 0 {
		return comError("Vista del escritorio", status)
	}
	defer comRelease(view)
	dispatchID, _ := comGUID("{00020400-0000-0000-C000-000000000046}")
	var folder uintptr
	status = comCall(view, 15, 0, uintptr(unsafe.Pointer(&dispatchID)), uintptr(unsafe.Pointer(&folder))) // SVGIO_BACKGROUND.
	if int32(status) < 0 || folder == 0 {
		return comError("Automatización del escritorio", status)
	}
	defer comRelease(folder)
	app, err := dispatchInvoke(folder, "Application", 2, nil)
	if err != nil {
		return err
	}
	defer oleAutomation.NewProc("VariantClear").Call(uintptr(unsafe.Pointer(&app)))
	if app.Type != 9 || app.Value == 0 {
		return fmt.Errorf("El Explorador no devolvió su aplicación de escritorio")
	}
	return action(uintptr(app.Value))
}
func explorerShellExecute(shell uintptr, file, arguments string) error {
	// IDispatch arguments are reversed: show, verb, directory, arguments, file.
	args := make([]automationVariant, 5)
	args[0] = automationVariant{Type: 3, Value: 1}
	for i, text := range []string{"open", "", arguments, file} {
		bstr, _, _ := oleAutomation.NewProc("SysAllocString").Call(ptr(text))
		if bstr == 0 {
			return fmt.Errorf("No se pudo preparar el enlace")
		}
		defer oleAutomation.NewProc("SysFreeString").Call(bstr)
		args[i+1] = automationVariant{Type: 8, Value: uint64(bstr)}
	}
	result, err := dispatchInvoke(shell, "ShellExecute", 1, args)
	defer oleAutomation.NewProc("VariantClear").Call(uintptr(unsafe.Pointer(&result)))
	return err
}
func onCOMThread(action func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	hr, _, _ := oleCOM.NewProc("CoInitializeEx").Call(0, 6)
	if int32(hr) < 0 {
		return comError("Inicializar COM", uint32(hr))
	}
	defer oleCOM.NewProc("CoUninitialize").Call()
	return action()
}
func openPurchase(link string) error {
	if !webURL(link) {
		return fmt.Errorf("El enlace de compra no es válido")
	}
	purchaseTrace("explorer_request", purchaseTarget(link))
	result := make(chan error, 1)
	threadID := make(chan uint32, 1)
	go func() {
		result <- onCOMThread(func() error {
			id, _, _ := kernel.NewProc("GetCurrentThreadId").Call()
			threadID <- uint32(id)
			hr, _, _ := oleCOM.NewProc("CoEnableCallCancellation").Call(0)
			if int32(hr) >= 0 {
				defer oleCOM.NewProc("CoDisableCallCancellation").Call(0)
			}
			return withExplorerShell(func(shell uintptr) error { return explorerShellExecute(shell, link, "") })
		})
	}()
	select {
	case err := <-result:
		if err != nil {
			purchaseTrace("explorer_error", err.Error())
		} else {
			purchaseTrace("explorer_accepted", "")
		}
		return err
	case <-time.After(3 * time.Second):
		select {
		case id := <-threadID:
			oleCOM.NewProc("CoCancelCall").Call(uintptr(id), 0)
		default:
		}
		err := fmt.Errorf("El Explorador no confirmó la apertura. La petición no se ha repetido")
		purchaseTrace("explorer_timeout", err.Error())
		return err
	}
}
