package main

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"time"
)

func fixtureBooks() []Book {
	return []Book{{"Uno", "https://example.test/1.jpg", "https://example.test/buy/1", "/1"}, {"Dos", "https://example.test/2.jpg", "https://example.test/buy/2", "/2"}, {"Tres", "https://example.test/3.jpg", "https://example.test/buy/3", "/3"}}
}
func TestVisualReference(t *testing.T) {
	var rows []struct {
		Size   [2]int
		Cycle  int64
		Rounds [][]int
		Fills  [][]int
		Texts  [][]json.RawMessage
		Images [][]int
	}
	data, e := os.ReadFile("testdata/visual-reference.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	for _, row := range rows {
		p := panelLayout(row.Size[0], row.Size[1], row.Cycle)
		rounds := [][]int{{p.Box.X - 10, p.Box.Y - 10, p.Box.W + 20, p.Box.H + 20, p.Radius + 10, 0x2212376f}, {p.Box.X - 4, p.Box.Y - 4, p.Box.W + 8, p.Box.H + 8, p.Radius + 4, 0x5c2255a2}, {p.Box.X - 1, p.Box.Y - 1, p.Box.W + 2, p.Box.H + 2, p.Radius + 1, 0xdc477fcf}, {p.Box.X + 1, p.Box.Y + 1, p.Box.W - 2, p.Box.H - 2, p.Radius, 0xff0b1423}}
		if !reflect.DeepEqual(rounds, row.Rounds) {
			t.Errorf("panel %v cycle %d: %v != %v", row.Size, row.Cycle, rounds, row.Rounds)
		}
		want := row.Images[0][2:]
		got := []int{p.Logo.X, p.Logo.Y, p.Logo.W, p.Logo.H}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("logo %v: %v != %v", row.Size, got, want)
		}
		got = []int{p.Separator.X, p.Separator.Y, p.Separator.W, p.Separator.H, 0xff526b91}
		if !reflect.DeepEqual(got, row.Fills[0]) {
			t.Errorf("separator %v", row.Size)
		}
		for i, field := range []struct {
			R    Box
			Size int
		}{{p.Clock, p.ClockSize}, {p.Footer, p.FooterSize}} {
			got = []int{field.R.X, field.R.Y, field.R.right(), field.R.bottom(), field.Size}
			for j, actual := range got {
				var expected int
				if e = json.Unmarshal(row.Texts[i][j+1], &expected); e != nil {
					t.Fatal(e)
				}
				if actual != expected {
					t.Errorf("text %v cycle %d row%d col%d: %d != %d", row.Size, row.Cycle, i, j, actual, expected)
				}
			}
		}
	}
}
func TestShelfAlignmentWithoutCutout(t *testing.T) {
	for _, size := range [][2]int{{1920, 1080}, {3840, 2160}, {1366, 768}, {800, 600}, {320, 240}, {200, 120}} {
		for cycle := int64(0); cycle < 10; cycle += 2 {
			p := panelLayout(size[0], size[1], cycle).Box
			slots := coverSlots(size[0], size[1], size[0] < 340, p)
			if len(slots) == 0 {
				t.Fatalf("no books at %v cycle%d", size, cycle)
			}
			baselines := shelfBaselines(size[0], size[1], size[0] < 340)
			for i, r := range slots {
				supported := false
				for _, base := range baselines {
					if r.bottom() == base {
						supported = true
					}
				}
				if !supported {
					t.Fatalf("floating book: %+v", r)
				}
				if r.X < 0 || r.Y < 0 || r.right() > size[0] || r.bottom() > size[1] {
					t.Fatalf("book outside screen: %+v", r)
				}
				if enlarged(r).bottom() != r.bottom() {
					t.Fatal("featured book left shelf")
				}
				for j := 0; j < i; j++ {
					if r.intersects(slots[j]) {
						t.Fatal("overlapping books")
					}
				}
			}
		}
	}
}

func TestPanelContainedAndMovesAsOne(t *testing.T) {
	for _, size := range [][2]int{{1920, 1080}, {3840, 2160}, {1366, 768}, {800, 600}, {320, 240}, {200, 120}, {80, 60}} {
		positions := map[Box]bool{}
		for cycle := int64(0); cycle < 10; cycle++ {
			p := panelLayout(size[0], size[1], cycle)
			positions[p.Box] = true
			if p.Box.X < 0 || p.Box.Y < 0 || p.Box.right() > size[0] || p.Box.bottom() > size[1] {
				t.Fatalf("panel outside screen: %+v", p)
			}
			if p.Logo.X < p.Box.X || p.Logo.Y < p.Box.Y || p.Logo.right() > p.Box.right() || p.Logo.bottom() > p.Box.bottom() {
				t.Fatalf("logo outside panel: %+v", p)
			}
			if cycle%2 == 1 && p.Box != panelLayout(size[0], size[1], cycle-1).Box {
				t.Fatal("panel moved before 30 s")
			}
		}
		if len(positions) != 5 {
			t.Errorf("%v: %d positions", size, len(positions))
		}
	}
}
func TestFadeCurveAndDeadline(t *testing.T) {
	for _, v := range []struct {
		D time.Duration
		A float64
	}{{0, 0}, {125 * time.Millisecond, .15625}, {250 * time.Millisecond, .5}, {375 * time.Millisecond, .84375}, {500 * time.Millisecond, 1}, {8 * time.Second, 1}, {14500 * time.Millisecond, 1}, {14750 * time.Millisecond, .5}, {15 * time.Second, 0}, {-1, 0}} {
		if math.Abs(opacity(v.D)-v.A) > 1e-12 {
			t.Errorf("%v: %f", v.D, opacity(v.D))
		}
	}
}
func TestCatalogUpdateDoesNotInterruptFeaturedBook(t *testing.T) {
	now := time.Unix(12345, 0)
	s := newScene(1920, 1080, false, 7, now)
	s.queue(fixtureBooks(), now)
	b, _ := s.current()
	replacement := []Book{{"Nuevo", "https://example.test/new.jpg", "https://example.test/buy/new", "/new"}}
	s.queue(replacement, now.Add(time.Second))
	if s.step(now.Add(featuredDuration - time.Nanosecond)) {
		t.Fatal("early transition")
	}
	if got, _ := s.current(); got != b {
		t.Fatal("new catalogue interrupted the featured cover")
	}
	s.resize(1366, 768)
	if got, _ := s.current(); got != b {
		t.Fatal("resize changed the selected book")
	}
	if !s.step(now.Add(featuredDuration)) {
		t.Fatal("deadline missed")
	}
	if got, _ := s.current(); got != replacement[0] {
		t.Fatal("pending catalogue not applied")
	}
	if s.Cycle != 1 {
		t.Fatal("cycle mismatch")
	}
	if !s.step(now.Add(2*featuredDuration)) || s.Cycle != 2 {
		t.Fatal("30 second move missed")
	}
}
func TestSelectedCoverAndURLStayPairedAndDoNotRepeat(t *testing.T) {
	now := time.Now()
	s := newScene(1920, 1080, false, 42, now)
	books := fixtureBooks()
	s.queue(books, now)
	last := ""
	for i := 0; i < 100; i++ {
		b, ok := s.current()
		if !ok {
			t.Fatal("missing featured cover")
		}
		found := false
		for _, want := range books {
			if want == b {
				found = true
			}
		}
		if !found {
			t.Fatal("cover and URL detached")
		}
		if i > 0 && bookID(b) == last {
			t.Fatal("avoidable repeated book")
		}
		last = bookID(b)
		now = now.Add(featuredDuration)
		s.step(now)
	}
}
func TestOnlyPurchasableBooksCanBeHighlighted(t *testing.T) {
	now := time.Now()
	books := fixtureBooks()
	books[0].Link = ""
	books[1].Link = ""
	s := newScene(1920, 1080, false, 42, now)
	s.queue(books, now)
	for i := 0; i < 10; i++ {
		b, ok := s.current()
		if !ok || b != books[2] {
			t.Fatal("selected a book without a valid link")
		}
		now = now.Add(featuredDuration)
		s.step(now)
	}
	books[2].Link = ""
	s.queue(books, now)
	s.step(now.Add(featuredDuration))
	if _, ok := s.current(); ok {
		t.Fatal("invalid featured link")
	}
}
func TestKeyboardAndPreview(t *testing.T) {
	for key := uint32(1); key < 256; key++ {
		exit, open := keyAction(key, false)
		if !exit || open != (key == 13) {
			t.Fatalf("key%d", key)
		}
		exit, open = keyAction(key, true)
		if exit || open {
			t.Fatal("preview consumed input")
		}
	}
}

func TestHighlightKeepsBackgroundAndDownloadOrder(t *testing.T) {
	now := time.Unix(12345, 0)
	s := newScene(1920, 1080, false, 42, now)
	books := fixtureBooks()
	s.queue(books, now)
	before := append([]int(nil), s.Assignment...)
	s.step(now.Add(featuredDuration))
	if !reflect.DeepEqual(before, s.Assignment) {
		t.Fatal("highlight shuffled the mosaic")
	}
	// A progress update may reorder books but must retain visible identities.
	s.queue([]Book{books[2], books[0], books[1]}, now)
	// Test the next non-layout-changing cycle.
	s.Cycle = 2
	old := make([]string, len(s.Assignment))
	for i, index := range s.Assignment {
		old[i] = bookID(s.Books[index])
	}
	s.step(now.Add(2 * featuredDuration))
	for i, index := range s.Assignment {
		if bookID(s.Books[index]) != old[i] {
			t.Fatal("download order changed the mosaic")
		}
	}
}

func TestBookTriggerLogoTracksPanel(t *testing.T) {
	for _, size := range [][2]int{{1920, 1080}, {1366, 768}, {800, 600}, {3840, 2160}} {
		for cycle := int64(0); cycle < 10; cycle++ {
			p := panelLayout(size[0], size[1], cycle).Box
			logo := booktriggerLogoBox(size[0], size[1], cycle)
			if logo.W <= 0 || logo.Y <= p.bottom() || logo.bottom() > size[1] || logo.X < 0 || logo.right() > size[0] {
				t.Fatalf("logo outside panel/screen: %+v %+v", p, logo)
			}
			if abs := logo.X + logo.W/2 - (p.X + p.W/2); abs < -1 || abs > 1 {
				t.Fatal("logo not centered")
			}
		}
	}
}

func TestClockMovementKeepsEveryShelfBook(t *testing.T) {
	now := time.Now()
	s := newScene(1920, 1080, false, 42, now)
	s.queue(fixtureBooks(), now)
	slots := append([]Box(nil), s.Slots...)
	assigned := append([]int(nil), s.Assignment...)
	for cycle := int64(2); cycle <= 8; cycle += 2 {
		s.Cycle = cycle
		s.relayout()
		s.assign(true)
		if !reflect.DeepEqual(slots, s.Slots) || !reflect.DeepEqual(assigned, s.Assignment) {
			t.Fatal("clock movement removed or shuffled shelf books")
		}
	}
}
