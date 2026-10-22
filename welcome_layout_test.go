package main

import "testing"

func TestWelcomePosterFitsUsableScreen(t *testing.T) {
	for _, size := range [][3]int{{1920, 1040, 96}, {1366, 728, 96}, {1366, 728, 144}, {3840, 2080, 192}, {800, 600, 96}} {
		w, h, dpi := size[0], size[1], size[2]
		l := completionLayout(w, h, dpi)
		if l.Width >= w || l.Height >= h || l.PosterHeight <= 0 || l.PosterWidth <= 0 {
			t.Fatalf("poster outside screen %v: %+v", size, l)
		}
		if l.PosterY+l.PosterHeight != l.FooterY || l.Width != l.PosterWidth+2*l.Margin {
			t.Fatalf("poster or footer overlap: %+v", l)
		}
		// At most one pixel of rounding when fitting the supplied 751 × 475 poster.
		delta := l.PosterWidth*475 - l.PosterHeight*751
		if delta < -751 || delta > 751 {
			t.Fatalf("poster aspect ratio changed: %+v", l)
		}
	}
}
