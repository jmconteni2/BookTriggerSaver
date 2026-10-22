//go:build windows

package main

import (
	"fmt"
	"math"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

var graphicsDLL = syscall.NewLazyDLL("gdiplus.dll")
var graphicsProcs = map[string]*syscall.LazyProc{}

func gp(name string, args ...uintptr) uint32 {
	p := graphicsProcs[name]
	if p == nil {
		p = graphicsDLL.NewProc(name)
		graphicsProcs[name] = p
	}
	r, _, _ := p.Call(args...)
	return uint32(r)
}
func signed(v int) uintptr { return uintptr(int64(v)) }

type gpStartup struct {
	Version                        uint32
	Debug                          uintptr
	SuppressThread, SuppressCodecs int32
}
type nativeImage struct {
	Handle uintptr
	W, H   int
	Used   uint64
}
type imageCache struct {
	items map[string]*nativeImage
	bytes uint64
	tick  uint64
}

func (c *imageCache) close() {
	for _, v := range c.items {
		gp("GdipDisposeImage", v.Handle)
	}
	c.items = nil
	c.bytes = 0
}
func (c *imageCache) get(path string) *nativeImage {
	if path == "" {
		return nil
	}
	c.tick++
	if v := c.items[path]; v != nil {
		v.Used = c.tick
		return v
	}
	var original uintptr
	if gp("GdipLoadImageFromFile", ptr(path), uintptr(unsafe.Pointer(&original))) != 0 || original == 0 {
		return nil
	}
	defer gp("GdipDisposeImage", original)
	var w, h uint32
	if gp("GdipGetImageWidth", original, uintptr(unsafe.Pointer(&w))) != 0 || gp("GdipGetImageHeight", original, uintptr(unsafe.Pointer(&h))) != 0 || w == 0 || h == 0 {
		return nil
	}
	scale := math.Min(1, math.Min(512/float64(w), 768/float64(h)))
	tw, th := max(1, int(float64(w)*scale)), max(1, int(float64(h)*scale))
	var bitmap, graphics uintptr
	if gp("GdipCreateBitmapFromScan0", uintptr(tw), uintptr(th), 0, 0x26200a, 0, uintptr(unsafe.Pointer(&bitmap))) != 0 || bitmap == 0 {
		return nil
	}
	if gp("GdipGetImageGraphicsContext", bitmap, uintptr(unsafe.Pointer(&graphics))) != 0 {
		gp("GdipDisposeImage", bitmap)
		return nil
	}
	gp("GdipSetInterpolationMode", graphics, 7)
	gp("GdipGraphicsClear", graphics, 0)
	status := gp("GdipDrawImageRectI", graphics, original, 0, 0, uintptr(tw), uintptr(th))
	gp("GdipDeleteGraphics", graphics)
	if status != 0 {
		gp("GdipDisposeImage", bitmap)
		return nil
	}
	cost := uint64(tw) * uint64(th) * 4
	if c.items == nil {
		c.items = map[string]*nativeImage{}
	}
	for c.bytes+cost > 64<<20 && len(c.items) > 0 {
		oldest := ""
		var tick uint64 = math.MaxUint64
		for key, v := range c.items {
			if v.Used < tick {
				tick = v.Used
				oldest = key
			}
		}
		v := c.items[oldest]
		gp("GdipDisposeImage", v.Handle)
		c.bytes -= uint64(v.W) * uint64(v.H) * 4
		delete(c.items, oldest)
	}
	v := &nativeImage{bitmap, tw, th, c.tick}
	c.items[path] = v
	c.bytes += cost
	return v
}

type surface struct {
	DC, Bitmap, Old uintptr
	W, H            int
}

func (s *surface) close() {
	if s.DC != 0 {
		if s.Old != 0 {
			gdi.NewProc("SelectObject").Call(s.DC, s.Old)
		}
		if s.Bitmap != 0 {
			gdi.NewProc("DeleteObject").Call(s.Bitmap)
		}
		gdi.NewProc("DeleteDC").Call(s.DC)
	}
	*s = surface{}
}
func (s *surface) ensure(dc uintptr, w, h int) error {
	if s.DC != 0 && s.W == w && s.H == h {
		return nil
	}
	s.close()
	if w <= 0 || h <= 0 || int64(w)*int64(h) > 80_000_000 {
		return fmt.Errorf("Dimensiones de pantalla no válidas")
	}
	mem, _, _ := gdi.NewProc("CreateCompatibleDC").Call(dc)
	if mem == 0 {
		return fmt.Errorf("No se pudo crear la memoria gráfica")
	}
	bmp, _, _ := gdi.NewProc("CreateCompatibleBitmap").Call(dc, uintptr(w), uintptr(h))
	if bmp == 0 {
		gdi.NewProc("DeleteDC").Call(mem)
		return fmt.Errorf("No se pudo crear la imagen de pantalla")
	}
	old, _, _ := gdi.NewProc("SelectObject").Call(mem, bmp)
	if old == 0 || old == ^uintptr(0) {
		gdi.NewProc("DeleteObject").Call(bmp)
		gdi.NewProc("DeleteDC").Call(mem)
		return fmt.Errorf("No se pudo seleccionar la imagen de pantalla")
	}
	*s = surface{mem, bmp, old, w, h}
	return nil
}
func graphicsFor(dc uintptr) (uintptr, error) {
	var g uintptr
	if status := gp("GdipCreateFromHDC", dc, uintptr(unsafe.Pointer(&g))); status != 0 || g == 0 {
		return 0, fmt.Errorf("GDI+ no pudo iniciar el dibujo (%d)", status)
	}
	gp("GdipSetSmoothingMode", g, 4)
	gp("GdipSetInterpolationMode", g, 7)
	return g, nil
}
func fillRect(g uintptr, r Box, color uint32) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	var brush uintptr
	if gp("GdipCreateSolidFill", uintptr(color), uintptr(unsafe.Pointer(&brush))) != 0 {
		return
	}
	defer gp("GdipDeleteBrush", brush)
	gp("GdipFillRectangleI", g, brush, signed(r.X), signed(r.Y), uintptr(r.W), uintptr(r.H))
}
func roundRect(g uintptr, r Box, radius int, color uint32) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	radius = min(radius, min(r.W, r.H)/2)
	if radius <= 0 {
		fillRect(g, r, color)
		return
	}
	var path, brush uintptr
	if gp("GdipCreatePath", 0, uintptr(unsafe.Pointer(&path))) != 0 {
		return
	}
	defer gp("GdipDeletePath", path)
	if gp("GdipCreateSolidFill", uintptr(color), uintptr(unsafe.Pointer(&brush))) != 0 {
		return
	}
	defer gp("GdipDeleteBrush", brush)
	d := radius * 2
	for _, arc := range [][3]int{{r.X, r.Y, 180}, {r.right() - d, r.Y, 270}, {r.right() - d, r.bottom() - d, 0}, {r.X, r.bottom() - d, 90}} {
		gp("GdipAddPathArcI", path, signed(arc[0]), signed(arc[1]), uintptr(d), uintptr(d), uintptr(math.Float32bits(float32(arc[2]))), uintptr(math.Float32bits(90)))
	}
	gp("GdipClosePathFigure", path)
	gp("GdipFillPath", g, brush, path)
}

// GdipAddPathArcI's two REAL angles are stack arguments in the Windows x64
// ABI; they are passed above using their IEEE 754 representations.
func drawImage(g uintptr, v *nativeImage, r Box, alpha float64) bool {
	if v == nil || r.W <= 0 || r.H <= 0 || alpha <= 0 {
		return false
	}
	if alpha >= 1 {
		return gp("GdipDrawImageRectI", g, v.Handle, signed(r.X), signed(r.Y), uintptr(r.W), uintptr(r.H)) == 0
	}
	var attrs uintptr
	if gp("GdipCreateImageAttributes", uintptr(unsafe.Pointer(&attrs))) != 0 {
		return false
	}
	defer gp("GdipDisposeImageAttributes", attrs)
	matrix := [25]float32{1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, float32(alpha), 0, 0, 0, 0, 0, 1}
	if gp("GdipSetImageAttributesColorMatrix", attrs, 1, 1, uintptr(unsafe.Pointer(&matrix[0])), 0, 0) != 0 {
		return false
	}
	status := gp("GdipDrawImageRectRectI", g, v.Handle, signed(r.X), signed(r.Y), uintptr(r.W), uintptr(r.H), 0, 0, uintptr(v.W), uintptr(v.H), 2, attrs, 0, 0)
	runtime.KeepAlive(matrix)
	return status == 0
}
func alphaColor(color uint32, a float64) uint32 {
	return color&0xffffff | uint32(float64(color>>24)*a+.5)<<24
}

type fontKey struct{ Size, Weight int }
type saverGraphics struct {
	front, background surface
	cache             imageCache
	BackgroundVersion uint64
	Fonts             map[fontKey]uintptr
}

func (r *saverGraphics) close() {
	r.front.close()
	r.background.close()
	r.cache.close()
	for _, f := range r.Fonts {
		gdi.NewProc("DeleteObject").Call(f)
	}
	r.Fonts = nil
}
func (r *saverGraphics) text(dc uintptr, text string, box Box, size, weight int, color uint32) {
	if text == "" || box.W <= 0 || box.H <= 0 {
		return
	}
	key := fontKey{size, weight}
	font := r.Fonts[key]
	if font == 0 {
		if r.Fonts == nil {
			r.Fonts = map[fontKey]uintptr{}
		}
		font, _, _ = gdi.NewProc("CreateFontW").Call(signed(-size), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, ptr("Segoe UI"))
		if font != 0 {
			r.Fonts[key] = font
		}
	}
	if font == 0 {
		return
	}
	old, _, _ := gdi.NewProc("SelectObject").Call(dc, font)
	defer gdi.NewProc("SelectObject").Call(dc, old)
	mode, _, _ := gdi.NewProc("SetBkMode").Call(dc, 1)
	defer gdi.NewProc("SetBkMode").Call(dc, mode)
	previous, _, _ := gdi.NewProc("SetTextColor").Call(dc, uintptr(color))
	defer gdi.NewProc("SetTextColor").Call(dc, previous)
	rect := winRect{int32(box.X), int32(box.Y), int32(box.right()), int32(box.bottom())}
	// DT_CENTER | DT_VCENTER | DT_SINGLELINE | DT_END_ELLIPSIS | DT_NOPREFIX.
	user32.NewProc("DrawTextW").Call(dc, ptr(text), ^uintptr(0), uintptr(unsafe.Pointer(&rect)), 0x8825)
}
func (r *saverGraphics) paint(dc uintptr, scene *Scene, now time.Time, status, logoPath, glowPath string) (string, error) {
	if e := r.front.ensure(dc, scene.W, scene.H); e != nil {
		return "", e
	}
	if e := r.background.ensure(dc, scene.W, scene.H); e != nil {
		return "", e
	}
	if r.BackgroundVersion != scene.Version || r.BackgroundVersion == 0 {
		g, e := graphicsFor(r.background.DC)
		if e != nil {
			return "", e
		}
		gp("GdipGraphicsClear", g, 0xff000000)
		for i, slot := range scene.Slots {
			if i == scene.Featured || i >= len(scene.Assignment) {
				continue
			}
			b := scene.Books[scene.Assignment[i]]
			fillRect(g, slot.grow(1), 0xff252935)
			drawImage(g, r.cache.get(b.Path), slot, 1)
		}
		gp("GdipDeleteGraphics", g)
		r.BackgroundVersion = scene.Version
	}
	gdi.NewProc("BitBlt").Call(r.front.DC, 0, 0, uintptr(scene.W), uintptr(scene.H), r.background.DC, 0, 0, 0xcc0020)
	g, e := graphicsFor(r.front.DC)
	if e != nil {
		return "", e
	}
	link := ""
	if book, ok := scene.current(); ok {
		rect := enlarged(scene.Slots[scene.Featured])
		a := scene.alpha(now)
		drawImage(g, r.cache.get(glowPath), haloBounds(rect), a)
		fillRect(g, rect.grow(3), alphaColor(0xff0d61f1, a))
		fillRect(g, rect.grow(1), alphaColor(0xff71c7ff, a))
		cover := r.cache.get(book.Path)
		if drawImage(g, cover, rect, a) {
			link = book.Link
		}
	}
	p := panelLayout(scene.W, scene.H, scene.Cycle)
	for _, v := range []struct {
		Expand, Radius int
		Color          uint32
	}{{10, p.Radius + 10, 0x2212376f}, {4, p.Radius + 4, 0x5c2255a2}, {1, p.Radius + 1, 0xdc477fcf}, {-1, p.Radius, 0xff0b1423}} {
		roundRect(g, p.Box.grow(v.Expand), v.Radius, v.Color)
	}
	drawImage(g, r.cache.get(logoPath), p.Logo, 1)
	fillRect(g, p.Separator, 0xff526b91)
	gp("GdipDeleteGraphics", g) // Flush GDI+ before drawing text on the same HDC.
	r.text(r.front.DC, now.Format("15:04"), p.Clock, p.ClockSize, 300, 0xf7f0eb)
	if !scene.Preview || scene.W > 340 {
		footer := status
		if book, ok := scene.current(); ok && book.Link != "" {
			footer = "¿Te interesa el libro destacado? Pulsa Intro"
		}
		r.text(r.front.DC, footer, p.Footer, p.FooterSize, 400, 0xcaae9b)
	}
	gdi.NewProc("BitBlt").Call(dc, 0, 0, uintptr(scene.W), uintptr(scene.H), r.front.DC, 0, 0, 0xcc0020)
	return link, nil
}
