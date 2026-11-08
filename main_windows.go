//go:build windows

package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

//go:embed assets/logo.png
var logo []byte

//go:embed assets/booktrigger.png
var booktriggerLogo []byte

//go:embed assets/ivy.png
var ivy []byte

//go:embed assets/glow.png
var glow []byte

//go:embed assets/endpoint.txt
var defaultEndpoint []byte

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	advapi         = syscall.NewLazyDLL("advapi32.dll")
	kernel         = syscall.NewLazyDLL("kernel32.dll")
	comctl         = syscall.NewLazyDLL("comctl32.dll")
	spi            = user32.NewProc("SystemParametersInfoW")
	regSet         = advapi.NewProc("RegSetValueExW")
	regDeleteValue = advapi.NewProc("RegDeleteValueW")
	regCreate      = advapi.NewProc("RegCreateKeyExW")
	regDeleteKey   = advapi.NewProc("RegDeleteKeyW")
	moveFile       = kernel.NewProc("MoveFileExW")
)

const (
	desktopKey   = `Control Panel\Desktop`
	uninstallKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\BookTriggerSaver`
	saverFile    = "BookTrigger.scr"
)

var wideStrings sync.Map

func wide(s string) *uint16 {
	if value, ok := wideStrings.Load(s); ok {
		return &value.([]uint16)[0]
	}
	u, e := syscall.UTF16FromString(s)
	if e != nil {
		panic(e)
	}
	value, _ := wideStrings.LoadOrStore(s, u)
	return &value.([]uint16)[0]
}
func ptr(s string) uintptr { return uintptr(unsafe.Pointer(wide(s))) }
func appDir() (string, error) {
	d, e := os.UserConfigDir()
	if e != nil {
		return "", e
	}
	return filepath.Join(d, "JMPocketSalvapantallas"), nil
}

type regValue struct {
	Exists bool   `json:"exists"`
	Kind   uint32 `json:"kind"`
	Data   []byte `json:"data,omitempty"`
}
type backup struct {
	Version int                 `json:"version"`
	Values  map[string]regValue `json:"values"`
}

var settings = []string{"SCRNSAVE.EXE", "ScreenSaveActive", "ScreenSaveTimeOut"}

func openKey(root syscall.Handle, path string, write bool) (syscall.Handle, error) {
	var h syscall.Handle
	access := uint32(0x20019)
	if write {
		access = 0x2001f
	}
	if !write {
		e := syscall.RegOpenKeyEx(root, wide(path), 0, access, &h)
		return h, e
	}
	r, _, _ := regCreate.Call(uintptr(root), ptr(path), 0, 0, 0, uintptr(access), 0, uintptr(unsafe.Pointer(&h)), 0)
	if r != 0 {
		return 0, syscall.Errno(r)
	}
	return h, nil
}
func readValue(h syscall.Handle, name string) (regValue, error) {
	var kind, n uint32
	e := syscall.RegQueryValueEx(h, wide(name), nil, &kind, nil, &n)
	if errors.Is(e, syscall.ERROR_FILE_NOT_FOUND) {
		return regValue{}, nil
	}
	if e != nil {
		return regValue{}, e
	}
	if n > 1024*1024 {
		return regValue{}, errors.New("Valor del registro demasiado grande")
	}
	data := make([]byte, n)
	if n > 0 {
		e = syscall.RegQueryValueEx(h, wide(name), nil, &kind, &data[0], &n)
		if e != nil {
			return regValue{}, e
		}
		data = data[:n]
	}
	return regValue{true, kind, data}, nil
}
func writeValue(h syscall.Handle, name string, v regValue) error {
	if !v.Exists {
		r, _, _ := regDeleteValue.Call(uintptr(h), ptr(name))
		if r == 0 || r == 2 {
			return nil
		}
		return syscall.Errno(r)
	}
	var p uintptr
	if len(v.Data) > 0 {
		p = uintptr(unsafe.Pointer(&v.Data[0]))
	}
	r, _, _ := regSet.Call(uintptr(h), ptr(name), 0, uintptr(v.Kind), p, uintptr(len(v.Data)))
	runtime.KeepAlive(v.Data)
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}
func stringValue(s string) regValue {
	u, _ := syscall.UTF16FromString(s)
	b := make([]byte, len(u)*2)
	for i, c := range u {
		b[2*i] = byte(c)
		b[2*i+1] = byte(c >> 8)
	}
	return regValue{true, 1, b}
}
func stringOf(v regValue) string {
	if !v.Exists || (v.Kind != 1 && v.Kind != 2) || len(v.Data)%2 != 0 {
		return ""
	}
	u := make([]uint16, len(v.Data)/2)
	for i := range u {
		u[i] = uint16(v.Data[i*2]) | uint16(v.Data[i*2+1])<<8
	}
	return syscall.UTF16ToString(u)
}
func snapshot(h syscall.Handle) (backup, error) {
	b := backup{Version: 1, Values: make(map[string]regValue)}
	for _, n := range settings {
		v, e := readValue(h, n)
		if e != nil {
			return b, e
		}
		b.Values[n] = v
	}
	return b, nil
}
func systemSet(action, value uint32) error {
	r, _, e := spi.Call(uintptr(action), uintptr(value), 0, 3)
	if r == 0 {
		return fmt.Errorf("Windows no aplicó la configuración (%d): %v", action, e)
	}
	return nil
}
func systemGet(action uint32) (uint32, error) {
	var v uint32
	r, _, e := spi.Call(uintptr(action), 0, uintptr(unsafe.Pointer(&v)), 0)
	if r == 0 {
		return 0, fmt.Errorf("Windows no permitió verificar la configuración: %v", e)
	}
	return v, nil
}
func restore(h syscall.Handle, b backup) error {
	var errs []error
	active := uint32(0)
	if stringOf(b.Values["ScreenSaveActive"]) == "1" {
		active = 1
	}
	timeout, _ := strconv.ParseUint(stringOf(b.Values["ScreenSaveTimeOut"]), 10, 32)
	if e := systemSet(0x11, active); e != nil {
		errs = append(errs, e)
	}
	if timeout > 0 {
		if e := systemSet(0xf, uint32(timeout)); e != nil {
			errs = append(errs, e)
		}
	}
	// SPI writes defaults to the registry: restore the exact original bytes last.
	for _, n := range settings {
		if e := writeValue(h, n, b.Values[n]); e != nil {
			errs = append(errs, e)
		}
	}
	broadcast()
	return errors.Join(errs...)
}
func broadcast() {
	var result uintptr
	user32.NewProc("SendMessageTimeoutW").Call(0xffff, 0x1a, 0, ptr(desktopKey), 2, 1500, uintptr(unsafe.Pointer(&result)))
}
func checkPolicy() error {
	path := `Software\Policies\Microsoft\Windows\Control Panel\Desktop`
	for _, root := range []syscall.Handle{syscall.HKEY_CURRENT_USER, syscall.HKEY_LOCAL_MACHINE} {
		h, e := openKey(root, path, false)
		if errors.Is(e, syscall.ERROR_FILE_NOT_FOUND) {
			continue
		}
		if e != nil {
			return e
		}
		for _, name := range settings {
			v, e := readValue(h, name)
			if e != nil {
				syscall.RegCloseKey(h)
				return e
			}
			if v.Exists {
				syscall.RegCloseKey(h)
				return errors.New("Este equipo tiene una directiva de la organización que controla el salvapantallas. La instalación no cambiará esa directiva.")
			}
		}
		syscall.RegCloseKey(h)
	}
	return nil
}
func apply(h syscall.Handle, path string, timeoutSeconds uint32) error {
	timeoutText := strconv.FormatUint(uint64(timeoutSeconds), 10)
	for n, v := range map[string]string{"SCRNSAVE.EXE": path, "ScreenSaveActive": "1", "ScreenSaveTimeOut": timeoutText} {
		if e := writeValue(h, n, stringValue(v)); e != nil {
			return e
		}
	}
	if e := systemSet(0xf, timeoutSeconds); e != nil {
		return e
	}
	if e := systemSet(0x11, 1); e != nil {
		return e
	}
	broadcast()
	for n, want := range map[string]string{"SCRNSAVE.EXE": path, "ScreenSaveActive": "1", "ScreenSaveTimeOut": timeoutText} {
		got, e := readValue(h, n)
		if e != nil {
			return e
		}
		if !strings.EqualFold(stringOf(got), want) {
			return fmt.Errorf("Windows no conservó el ajuste %s", n)
		}
	}
	active, e := systemGet(0x10)
	if e != nil {
		return e
	}
	if active == 0 {
		return errors.New("Windows mantiene desactivado el salvapantallas")
	}
	timeout, e := systemGet(0xe)
	if e != nil {
		return e
	}
	if timeout != timeoutSeconds {
		return errors.New("Windows no conservó el tiempo de espera")
	}
	return nil
}

// Replace fully written files, instead of truncating an existing installation.
func atomicWrite(path string, data []byte) error {
	if old, e := os.ReadFile(path); e == nil && bytes.Equal(old, data) {
		return nil
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".booktrigger-*")
	if e != nil {
		return e
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, e = f.Write(data); e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	r, _, err := moveFile.Call(ptr(temp), ptr(path), 1|8)
	if r == 0 {
		return fmt.Errorf("No se pudo sustituir %s. Cierra el salvapantallas y vuelve a intentarlo: %v", filepath.Base(path), err)
	}
	return nil
}

type fileChange struct {
	path    string
	old     []byte
	existed bool
}

func replaceFiles(files map[string][]byte) ([]fileChange, error) {
	var changes []fileChange
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		b := files[path]
		old, e := os.ReadFile(path)
		if e != nil && !errors.Is(e, os.ErrNotExist) {
			return changes, e
		}
		changes = append(changes, fileChange{path, old, e == nil})
		if e := atomicWrite(path, b); e != nil {
			return changes, e
		}
	}
	return changes, nil
}
func rollbackFiles(changes []fileChange) error {
	var errs []error
	for i := len(changes) - 1; i >= 0; i-- {
		c := changes[i]
		var e error
		if c.existed {
			e = atomicWrite(c.path, c.old)
		} else {
			e = os.Remove(c.path)
			if errors.Is(e, os.ErrNotExist) {
				e = nil
			}
		}
		if e != nil {
			errs = append(errs, e)
		}
	}
	return errors.Join(errs...)
}
func registerUninstaller(path string) error {
	h, e := openKey(syscall.HKEY_CURRENT_USER, uninstallKey, true)
	if e != nil {
		return e
	}
	defer syscall.RegCloseKey(h)
	for n, v := range map[string]string{
		"DisplayName": "BookTrigger — Salvapantallas", "DisplayVersion": "11.0.5",
		"Publisher": "J.M. Contenidos", "DisplayIcon": path,
		"InstallLocation": filepath.Dir(path), "UninstallString": `"` + path + `" --uninstall`,
	} {
		if e := writeValue(h, n, stringValue(v)); e != nil {
			return e
		}
	}
	return nil
}
func install(timeoutSeconds uint32) error {
	if timeoutSeconds < 60 || timeoutSeconds > 1800 {
		return errors.New("Tiempo de espera no válido")
	}
	if e := checkPolicy(); e != nil {
		return e
	}
	d, e := appDir()
	if e != nil {
		return e
	}
	if e = os.MkdirAll(d, 0755); e != nil {
		return e
	}
	h, e := openKey(syscall.HKEY_CURRENT_USER, desktopKey, true)
	if e != nil {
		return e
	}
	defer syscall.RegCloseKey(h)
	before, e := snapshot(h)
	if e != nil {
		return e
	}
	exe, e := os.Executable()
	if e != nil {
		return e
	}
	self, e := os.ReadFile(exe)
	if e != nil {
		return e
	}
	path := filepath.Join(d, saverFile)
	files := map[string][]byte{path: self, filepath.Join(d, "logo.png"): logo, filepath.Join(d, "ivy.png"): ivy, filepath.Join(d, "booktrigger.png"): booktriggerLogo, filepath.Join(d, "glow.png"): glow}
	endpointPath := filepath.Join(d, "endpoint.txt")
	if _, e = os.Stat(endpointPath); errors.Is(e, os.ErrNotExist) {
		files[endpointPath] = defaultEndpoint
	} else if e != nil {
		return e
	}
	backupPath := filepath.Join(d, "configuracion_anterior.json")
	if raw, e := os.ReadFile(backupPath); errors.Is(e, os.ErrNotExist) {
		b, _ := json.MarshalIndent(before, "", "  ")
		files[backupPath] = b
	} else if e != nil {
		return e
	} else {
		var old backup
		if json.Unmarshal(raw, &old) != nil || old.Version != 1 || len(old.Values) != 3 {
			return errors.New("La copia de configuración anterior está dañada; no se sobrescribirá")
		}
	}
	changes, e := replaceFiles(files)
	if e != nil {
		return errors.Join(e, rollbackFiles(changes))
	}
	if e = apply(h, path, timeoutSeconds); e != nil {
		return errors.Join(e, restore(h, before), rollbackFiles(changes))
	}
	if e = registerUninstaller(path); e != nil {
		return errors.Join(e, restore(h, before), rollbackFiles(changes))
	}
	return nil
}

// TASKDIALOGCONFIG and TASKDIALOG_BUTTON use the Windows SDK's pack(1).
// Build byte arrays explicitly: native Go structs would add incompatible padding.
func put32(b []byte, p int, v uint32) {
	for i := 0; i < 4; i++ {
		b[p+i] = byte(v >> (8 * i))
	}
}
func put64(b []byte, p int, v uintptr) {
	for i := 0; i < 8; i++ {
		b[p+i] = byte(v >> (8 * i))
	}
}
func installOptions() (uint32, bool) {
	if comctl.NewProc("TaskDialogIndirect").Find() != nil {
		message("No se ha podido abrir el instalador de Windows. Comprueba que usas Windows 10 u 11.", 0x10)
		return 0, false
	}
	b := make([]byte, 160)
	buttons := make([]byte, 24)
	put32(buttons, 0, 100)
	put64(buttons, 4, uintptr(unsafe.Pointer(wide("Instalar y activar"))))
	put32(buttons, 12, 2)
	put64(buttons, 16, uintptr(unsafe.Pointer(wide("Cancelar"))))
	minutes := []int{1, 2, 3, 5, 10, 15, 30}
	radios := make([]byte, len(minutes)*12)
	for i, m := range minutes {
		put32(radios, i*12, uint32(200+m))
		label := fmt.Sprintf("%d minutos", m)
		if m == 1 {
			label = "1 minuto"
		}
		put64(radios, i*12+4, uintptr(unsafe.Pointer(wide(label))))
	}
	put32(b, 0, 160)
	put32(b, 20, 8)
	put64(b, 28, ptr("Instalar BookTrigger"))
	put64(b, 44, ptr("Elige cuándo se activa el salvapantallas"))
	put64(b, 52, ptr("Selecciona el tiempo sin actividad. BookTrigger quedará seleccionado y activo al terminar.\n\nSe conservará tu configuración anterior para poder restaurarla.\n\nLa aplicación consulta el catálogo público de BookTrigger y descarga sus portadas. No incorpora analítica ni envía tus documentos."))
	put32(b, 60, 2)
	put64(b, 64, uintptr(unsafe.Pointer(&buttons[0])))
	put32(b, 72, 100)
	put32(b, 76, uint32(len(minutes)))
	put64(b, 80, uintptr(unsafe.Pointer(&radios[0])))
	put32(b, 88, 205)
	var selected, radio int32
	hr, _, _ := comctl.NewProc("TaskDialogIndirect").Call(uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&selected)), uintptr(unsafe.Pointer(&radio)), 0)
	runtime.KeepAlive([]any{buttons, radios, b})
	if int32(hr) < 0 {
		message("Windows no pudo abrir las opciones de instalación.", 0x10)
		return 0, false
	}
	if selected != 100 {
		return 0, false
	}
	for _, m := range minutes {
		if radio == int32(200+m) {
			return uint32(m * 60), true
		}
	}
	return 0, false
}
func dialog(title, heading, body, action string) bool {
	if comctl.NewProc("TaskDialogIndirect").Find() != nil {
		return message(body+"\n\n¿"+action+"?", 4|0x40) == 6
	}
	// Pack(1) SDK layout: 160 bytes on win64, not the natural 176-byte Go layout.
	b := make([]byte, 160)
	titleW, headW, bodyW, actionW, closeW := wide(title), wide(heading), wide(body), wide(action), wide("Cerrar")
	buttons := make([]byte, 24)
	put32(buttons, 0, 100)
	put64(buttons, 4, uintptr(unsafe.Pointer(actionW)))
	put32(buttons, 12, 2)
	put64(buttons, 16, uintptr(unsafe.Pointer(closeW)))
	put32(b, 0, uint32(len(b)))
	put32(b, 20, 8) // TDF_ALLOW_DIALOG_CANCELLATION
	put64(b, 28, uintptr(unsafe.Pointer(titleW)))
	put64(b, 44, uintptr(unsafe.Pointer(headW)))
	put64(b, 52, uintptr(unsafe.Pointer(bodyW)))
	put32(b, 60, 2)
	put64(b, 64, uintptr(unsafe.Pointer(&buttons[0])))
	put32(b, 72, 100)
	var selected int32
	hr, _, _ := comctl.NewProc("TaskDialogIndirect").Call(uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&selected)), 0, 0)
	runtime.KeepAlive([]any{titleW, headW, bodyW, actionW, closeW, buttons, b})
	if int32(hr) < 0 {
		return message(body+"\n\n¿"+action+"?", 4|0x40) == 6
	}
	return selected == 100
}
func message(s string, flags uintptr) uintptr {
	r, _, _ := user32.NewProc("MessageBoxW").Call(0, ptr(s), ptr("BookTrigger"), flags)
	return r
}
func engineDataDir() (string, error) {
	if storeMode { return storeDataDir() }
	return appDir()
}

func logError(err error) {
	if d, e := engineDataDir(); e == nil {
		if f, e := os.OpenFile(filepath.Join(d, "instalacion.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); e == nil {
			fmt.Fprintln(f, time.Now().Format(time.RFC3339), err)
			f.Close()
		}
	}
}
func runEngine(mode string, parent uintptr) error {
	d, e := engineDataDir()
	if e != nil {
		return e
	}
	if mode == "preview" {
		r, _, _ := user32.NewProc("IsWindow").Call(parent)
		if r == 0 {
			return nil
		}
	}
	var handle syscall.Handle
	if mode != "preview" {
		r, _, err := kernel.NewProc("CreateMutexW").Call(0, 0, ptr(`Local\BookTriggerSaverRunning`))
		if r == 0 {
			return err
		}
		handle = syscall.Handle(r)
		defer syscall.CloseHandle(handle)
		if err == syscall.Errno(183) {
			return nil
		}
	}
	return runSaver(d, mode == "preview", parent)
}

func uninstall() error {
	if message("Se desactivará BookTrigger y se restaurará el salvapantallas anterior. ¿Continuar?", 4|0x20) != 6 {
		return nil
	}
	d, e := appDir()
	if e != nil {
		return e
	}
	h, e := openKey(syscall.HKEY_CURRENT_USER, desktopKey, true)
	if e != nil {
		return e
	}
	defer syscall.RegCloseKey(h)
	current, e := readValue(h, "SCRNSAVE.EXE")
	if e != nil {
		return e
	}
	if strings.EqualFold(stringOf(current), filepath.Join(d, saverFile)) {
		raw, e := os.ReadFile(filepath.Join(d, "configuracion_anterior.json"))
		if e != nil {
			return e
		}
		var b backup
		if e = json.Unmarshal(raw, &b); e != nil {
			return e
		}
		if b.Version != 1 || len(b.Values) != 3 {
			return errors.New("La copia de configuración anterior no es válida")
		}
		if e = restore(h, b); e != nil {
			return e
		}
	}
	// The .scr currently in use is retained as a reinstall entry point.
	// Remove only owned files; preserve cached books and previous v9 installations.
	for _, name := range []string{"configuracion_anterior.json"} {
		if e = os.Remove(filepath.Join(d, name)); e != nil && !errors.Is(e, os.ErrNotExist) {
			return e
		}
	}
	r, _, _ := regDeleteKey.Call(uintptr(syscall.HKEY_CURRENT_USER), ptr(uninstallKey))
	if r != 0 && r != 2 {
		return syscall.Errno(r)
	}
	message("BookTrigger se ha desactivado. La configuración anterior se ha restaurado cuando correspondía.", 0x40)
	return nil
}
func main() {
	runtime.LockOSThread()
	if storeMode {
		if handled, err := storeNativeHelper(os.Args[1:]); handled {
			if err != nil { logError(err); os.Exit(1) }
			return
		}
	}
	if storeMode && captureLauncher() { return }
	mode, parent := parseMode(os.Args[1:], strings.EqualFold(filepath.Ext(os.Args[0]), ".scr"))
	if storeMode {
		if err := storeMain(mode, parent); err != nil {
			logError(err)
			if mode != "preview" && mode != "saver" { message("No se pudo completar la operación.\n\n" + err.Error(), 0x10) }
			os.Exit(1)
		}
		return
	}
	var err error
	switch mode {
	case "install":
		timeoutSeconds, ok := installOptions()
		if !ok {
			return
		}
		err = install(timeoutSeconds)
		if err == nil && completedPage(timeoutSeconds) {
			err = runEngine("saver", 0)
		}
	case "saver", "preview":
		err = runEngine(mode, parent)
	case "config":
		if dialog("BookTrigger", "Salvapantallas BookTrigger", "Libro destacado: 15 segundos. Caja del logo y la hora: cambio de posición cada 30 segundos.\n\nIntro abre la compra del libro destacado. Cualquier otra tecla cierra el salvapantallas.", "Ejecutar ahora") {
			err = runEngine("saver", 0)
		}
	case "uninstall":
		err = uninstall()
	case "invalid-preview":
		return
	default:
		err = errors.New("Opción no reconocida")
	}
	if err != nil {
		logError(err)
		if mode != "preview" && mode != "saver" {
			message("No se pudo completar la operación.\n\n"+err.Error(), 0x10)
		}
		os.Exit(1)
	}
}
