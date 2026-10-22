//go:build windows

package main

// Cached alongside the mosaic: deterministic shelf geometry stays aligned
// with coverSlots at all resolutions, with no remote background download.
func drawShelves(g uintptr, w, h int, preview bool) {
	grid := shelfGeometry(w, h, preview)
	top := 0
	for _, base := range shelfBaselines(w, h, preview) {
		height := max(1, base-top)
		for i := 0; i < 64; i++ {
			y := top + height*i/64
			next := top + height*(i+1)/64
			light := i * 15 / 63
			color := uint32(0xff000000) | uint32(14+light)<<16 | uint32(22+light)<<8 | uint32(18+light)
			fillRect(g, Box{0, y, w, max(1, next-y)}, color)
		}
		// Fine vertical grain, shadow under the rear of the shelf, and its face.
		for x := 13; x < w; x += 37 {
			fillRect(g, Box{x, top, 1, height}, 0x05000000)
		}
		fillRect(g, Box{0, base - 2, w, 2}, 0xff738074)
		fillRect(g, Box{0, base, w, max(2, grid.Lip/3)}, 0xff424e44)
		fillRect(g, Box{0, base + max(2, grid.Lip/3), w, max(2, grid.Lip*2/3)}, 0xff1b251e)
		fillRect(g, Box{0, base + grid.Lip - 1, w, 1}, 0xff080e0b)
		top = base + grid.Lip
	}
}
