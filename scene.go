package main

import (
	"math"
	"math/rand"
	"time"
)

const featuredDuration = 15 * time.Second
const fadeDuration = 500 * time.Millisecond

type Box struct{ X, Y, W, H int }

func (r Box) right() int  { return r.X + r.W }
func (r Box) bottom() int { return r.Y + r.H }
func (r Box) intersects(s Box) bool {
	return r.X < s.right() && r.right() > s.X && r.Y < s.bottom() && r.bottom() > s.Y
}
func (r Box) grow(n int) Box  { return Box{r.X - n, r.Y - n, r.W + 2*n, r.H + 2*n} }
func clamp(x, lo, hi int) int { return max(lo, min(x, hi)) }

type Panel struct {
	Box        Box
	Radius     int
	Logo       Box
	Separator  Box
	Clock      Box
	ClockSize  int
	Footer     Box
	FooterSize int
}

func panelLayout(w, h int, cycle int64) Panel {
	pw, ph, r := clamp(int(float64(w)*.3), 300, 460), clamp(int(float64(h)*.33), 205, 288), 22
	if w <= 340 || h <= 240 {
		pw = max(65, int(float64(w)*.56))
		ph = max(43, int(float64(h)*.54))
		r = 8
	}
	// Extremely small host windows should remain usable rather than overflowing.
	pw = min(pw, max(1, w-4))
	ph = min(ph, max(1, h-4))
	r = min(r, min(pw, ph)/2)
	dx, dy := w-pw, h-ph
	x, y := dx/2, dy/2
	switch max(int64(0), cycle) / 2 % 5 {
	case 1:
		x, y = dx/8, dy/8
	case 2:
		x, y = dx-dx/8, dy/8
	case 3:
		x, y = dx-dx/8, dy-dy/8
	case 4:
		x, y = dx/8, dy-dy/8
	}
	// The original logo's source dimensions are 725 by 516.
	lw := max(1, int(float64(pw)*.65))
	ih := max(1, int(float64(lw)*516/725))
	if limit := max(1, int(float64(ph)*.42)); ih > limit {
		ih = limit
		lw = max(1, int(725*float64(ih)/516))
	}
	logo := Box{x + pw/2 - lw/2, y + int(float64(ph)*.3) - ih/2, lw, ih}
	clock := Box{x + 3, y + int(float64(ph)*.49), max(1, pw-6), max(1, int(float64(ph)*.84)-int(float64(ph)*.49))}
	footer := Box{x + 8, y + int(float64(ph)*.85), max(1, pw-16), max(1, ph-6-int(float64(ph)*.85))}
	return Panel{Box{x, y, pw, ph}, r, logo, Box{x + pw/2 - int(float64(pw)*.05), y + int(float64(ph)*.52), int(float64(pw) * .1), 1}, clock, clamp(int(math.Min(float64(pw)*.18, float64(ph)*.34)), 10, 88), footer, clamp(int(float64(w)*.009), 10, 14)}
}

// Keep the supplied 1419x571 logo below the clock panel, anchored to it.
func booktriggerLogoBox(w, h int, cycle int64) Box {
	return logoBelowPanel(h, panelLayout(w, h, cycle).Box)
}
func logoBelowPanel(h int, p Box) Box {
	gap := clamp(p.H/24, 6, 14)
	available := h - p.bottom() - gap - 8
	if available < 12 {
		return Box{}
	}
	width := min(p.W*65/100, 1419*min(available, p.H/3)/571)
	height := width * 571 / 1419
	if width < 1 || height < 1 {
		return Box{}
	}
	return Box{p.X + (p.W-width)/2, p.bottom() + gap, width, height}
}

type shelfGrid struct {
	Columns, Rows int
	CellW, CellH  float64
	Lip           int
}

func shelfGeometry(w, h int, preview bool) shelfGrid {
	cols, rows := clamp(w/170, 3, 24), clamp(h/235, 2, 12)
	if preview && w < 340 {
		cols, rows = 4, 3
	}
	return shelfGrid{cols, rows, float64(w) / float64(cols), float64(h) / float64(rows), clamp(h/80, 8, 20)}
}
func shelfBaselines(w, h int, preview bool) []int {
	grid := shelfGeometry(w, h, preview)
	out := make([]int, grid.Rows)
	for row := range out {
		out[row] = int(float64(row+1)*grid.CellH) - grid.Lip
	}
	return out
}
func plantsAroundPanel(w, h int, p Box) Box {
	if w <= 340 || h <= 240 {
		return Box{}
	}
	edge := clamp(p.W/10, 8, 48)
	top := clamp(p.H/10, 6, 28)
	return Box{p.X - edge, p.Y - top, p.W + 2*edge, (p.W + 2*edge) * 2 / 3}
}
func coverSlots(w, h int, preview bool, panel Box) []Box {
	grid := shelfGeometry(w, h, preview)
	width := min(grid.CellW*.75, grid.CellH*.76/1.46)
	height := width * 1.46
	margin := clamp(w/100, 8, 24)
	reserved := panel.grow(margin)
	brand := logoBelowPanel(h, panel)
	plants := plantsAroundPanel(w, h, panel)
	var out []Box
	for _, baseline := range shelfBaselines(w, h, preview) {
		for col := 0; col < grid.Columns; col++ {
			r := Box{int(float64(col)*grid.CellW + (grid.CellW-width)/2), baseline - int(height), int(width), int(height)}
			if r.W > 0 && r.H > 0 && !r.intersects(reserved) && (brand.W == 0 || !r.intersects(brand.grow(6))) && (plants.W == 0 || !r.intersects(plants.grow(6))) {
				out = append(out, r)
			}
		}
	}
	return out
}
func enlarged(r Box) Box {
	w, h := int(float64(r.W)*1.45), int(float64(r.H)*1.45)
	// Keep the featured book resting on its shelf during the fade.
	return Box{r.X - (w-r.W)/2, r.bottom() - h, w, h}
}
func (s *Scene) highlightFits(slot int) bool {
	r := enlarged(s.Slots[slot])
	p := panelLayout(s.W, s.H, s.Cycle).Box
	brand := booktriggerLogoBox(s.W, s.H, s.Cycle)
	plants := plantsAroundPanel(s.W, s.H, p)
	return r.X >= 2 && r.Y >= 2 && r.right() <= s.W-2 && r.bottom() <= s.H && !r.intersects(p.grow(10)) && (brand.W == 0 || !r.intersects(brand.grow(6))) && (plants.W == 0 || !r.intersects(plants.grow(6)))
}

func haloBounds(r Box) Box {
	return Box{r.X - r.W*96/272, r.Y - r.H*96/408, r.W * 464 / 272, r.H * 600 / 408}
}
func opacity(elapsed time.Duration) float64 {
	if elapsed < 0 || elapsed >= featuredDuration {
		return 0
	}
	x := 1.0
	if elapsed < fadeDuration {
		x = float64(elapsed) / float64(fadeDuration)
	}
	if featuredDuration-elapsed < fadeDuration {
		x = float64(featuredDuration-elapsed) / float64(fadeDuration)
	}
	return x * x * (3 - 2*x)
}

type Book struct {
	Title string `json:"titulo"`
	Cover string `json:"portada"`
	Link  string `json:"enlace"`
	Path  string `json:"-"`
}

func bookID(b Book) string { return b.Cover + "\n" + b.Link }

type Scene struct {
	W, H             int
	Preview          bool
	Books            []Book
	pending          []Book
	Slots            []Box
	Assignment       []int
	Featured         int
	Previous         string
	Cycle            int64
	Started, Changed time.Time
	Version          uint64
	rng              *rand.Rand
}

func newScene(w, h int, preview bool, seed int64, now time.Time) *Scene {
	s := &Scene{W: w, H: h, Preview: preview, Featured: -1, Started: now, Changed: now, rng: rand.New(rand.NewSource(seed))}
	s.relayout()
	return s
}
func (s *Scene) relayout() {
	s.Slots = coverSlots(s.W, s.H, s.Preview, panelLayout(s.W, s.H, s.Cycle).Box)
	s.Assignment = nil
	s.Version++
}
func (s *Scene) resize(w, h int) {
	if s.W == w && s.H == h {
		return
	}
	visible, ok := s.current()
	s.W, s.H = w, h
	s.relayout()
	s.assign(false)
	if ok && s.Featured >= 0 {
		for i, b := range s.Books {
			if bookID(b) == bookID(visible) {
				s.Assignment[s.Featured] = i
				break
			}
		}
	}
}
func (s *Scene) queue(books []Book, now time.Time) {
	copyBooks := append([]Book(nil), books...)
	if len(copyBooks) == 0 {
		return
	}
	if len(s.Books) == 0 {
		s.Books = copyBooks
		s.Changed = now
		s.Started = now
		s.assign(true)
		return
	}
	s.pending = copyBooks
}
func (s *Scene) assign(next bool) {
	if len(s.Books) == 0 || len(s.Slots) == 0 {
		s.Featured = -1
		s.Version++
		return
	}
	if len(s.Assignment) != len(s.Slots) {
		s.Assignment = make([]int, len(s.Slots))
		perm := s.rng.Perm(len(s.Books))
		for i := range s.Assignment {
			s.Assignment[i] = perm[i%len(perm)]
		}
	}
	var eligible []int
	for i, b := range s.Books {
		if webURL(b.Link) {
			eligible = append(eligible, i)
		}
	}
	if len(eligible) == 0 {
		s.Featured = -1
		s.Version++
		return
	}
	// Select a cover already in the mosaic, without shuffling the background.
	var candidates []int
	for slot, index := range s.Assignment {
		b := s.Books[index]
		if webURL(b.Link) && (!next || bookID(b) != s.Previous) && s.highlightFits(slot) {
			candidates = append(candidates, slot)
		}
	}
	chosen := s.rng.Intn(len(s.Slots))
	for slot := range s.Slots {
		if s.highlightFits(slot) {
			chosen = slot
			break
		}
	}
	if len(candidates) > 0 {
		chosen = candidates[s.rng.Intn(len(candidates))]
	} else {
		book := eligible[s.rng.Intn(len(eligible))]
		for _, i := range eligible {
			if !next || bookID(s.Books[i]) != s.Previous {
				book = i
				break
			}
		}
		s.Assignment[chosen] = book
	}
	s.Featured = chosen
	s.Version++
}
func (s *Scene) step(now time.Time) bool {
	if len(s.Books) == 0 {
		return false
	}
	if now.Sub(s.Changed) < featuredDuration {
		return false
	}
	if b, ok := s.current(); ok {
		s.Previous = bookID(b)
	}
	if len(s.pending) > 0 {
		// Keep already displayed covers in their slots when download progress
		// changes the catalogue order. Only unavailable covers are replaced.
		indices := make(map[string]int, len(s.pending))
		for i, b := range s.pending {
			indices[bookID(b)] = i
		}
		for slot, old := range s.Assignment {
			index, found := indices[bookID(s.Books[old])]
			if !found {
				index = s.rng.Intn(len(s.pending))
			}
			s.Assignment[slot] = index
		}
		s.Books = s.pending
		s.pending = nil
	}
	s.Cycle++
	if s.Cycle%2 == 0 {
		s.relayout()
	}
	s.Changed = now // After a delayed frame, the next visible book still gets 15 s.
	s.assign(true)
	return true
}
func (s *Scene) current() (Book, bool) {
	if s.Featured < 0 || s.Featured >= len(s.Assignment) {
		return Book{}, false
	}
	i := s.Assignment[s.Featured]
	if i < 0 || i >= len(s.Books) {
		return Book{}, false
	}
	return s.Books[i], true
}
func (s *Scene) alpha(now time.Time) float64 { return opacity(now.Sub(s.Changed)) }
func (s *Scene) animating(now time.Time) bool {
	if _, ok := s.current(); !ok {
		return false
	}
	d := now.Sub(s.Changed)
	return d < fadeDuration+50*time.Millisecond || d > featuredDuration-fadeDuration-50*time.Millisecond
}
func keyAction(key uint32, preview bool) (exit, open bool) {
	if preview {
		return false, false
	}
	return true, key == 13
}
