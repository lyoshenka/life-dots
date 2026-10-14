package calendar

import (
	"testing"
	"time"

	"github.com/grin/life-dots/internal/model"
)

func TestBuildLife(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Mode = model.ModeLife
	cfg.Birthday = "1995-01-01"
	cfg.Lifespan = 90
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.Local)
	g, err := Build(cfg, now)
	if err != nil {
		t.Fatal(err)
	}
	if g.Columns != 52 {
		t.Fatalf("cols=%d", g.Columns)
	}
	if g.Rows != 90 {
		t.Fatalf("rows=%d", g.Rows)
	}
	if g.Total != 90*52 {
		t.Fatalf("total=%d", g.Total)
	}
	if g.Filled <= 0 || g.Filled >= g.Total {
		t.Fatalf("filled=%d", g.Filled)
	}
	if g.Current < 0 {
		t.Fatal("expected current week")
	}
}

func TestBuildYearDays(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Mode = model.ModeYear
	cfg.Year = 2024 // leap
	now := time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local)
	g, err := Build(cfg, now)
	if err != nil {
		t.Fatal(err)
	}
	if g.Total != 366 {
		t.Fatalf("leap year days=%d", g.Total)
	}
}

func TestBuildHabits(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Mode = model.ModeHabits
	cfg.Habits = []model.Habit{
		{ID: "1", Icon: "check", Completions: map[string]bool{"2026-10-14": true}},
		{ID: "2", Icon: "book-open"},
	}
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.Local)
	g, err := Build(cfg, now)
	if err != nil {
		t.Fatal(err)
	}
	if g.Columns != 14 {
		t.Fatalf("cols=%d", g.Columns)
	}
	if g.Rows != 16 { // 8*2
		t.Fatalf("rows=%d", g.Rows)
	}
}

func TestBuildCustomAdaptive(t *testing.T) {
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	cases := []struct {
		start, end string
		unit       string
	}{
		{"2026-01-01", "2026-01-20", "day"},   // < 35 days
		{"2026-01-01", "2026-02-10", "week"},  // ~6 weeks
		{"2026-01-01", "2026-12-31", "week"},  // ~1 year
		{"2026-01-01", "2028-01-01", "week"},  // ~2 years
		{"2026-01-01", "2032-01-01", "month"}, // multi-year
	}

	for _, tc := range cases {
		cfg := model.DefaultConfig()
		cfg.Mode = model.ModeCustom
		cfg.StartDate = tc.start
		cfg.EndDate = tc.end
		g, err := Build(cfg, now)
		if err != nil {
			t.Fatalf("%s→%s: %v", tc.start, tc.end, err)
		}
		if g.Unit != tc.unit {
			t.Fatalf("%s→%s: unit=%s want %s", tc.start, tc.end, g.Unit, tc.unit)
		}
		if g.Columns < 1 || g.Rows < 1 || g.Total < 1 {
			t.Fatalf("%s→%s: bad grid %dx%d total=%d", tc.start, tc.end, g.Columns, g.Rows, g.Total)
		}
		if len(g.Cells) != g.Columns*g.Rows {
			t.Fatalf("%s→%s: cells=%d want %d", tc.start, tc.end, len(g.Cells), g.Columns*g.Rows)
		}
		switch g.Unit {
		case "day":
			if g.Columns != 7 {
				t.Fatalf("day grid should be 7-wide, got %d", g.Columns)
			}
		case "week":
			if g.Columns < 4 || g.Columns > 26 {
				t.Fatalf("week cols=%d out of range", g.Columns)
			}
		case "month":
			if g.Columns < 3 || g.Columns > 12 {
				t.Fatalf("month cols=%d out of range", g.Columns)
			}
		}
	}
}

func TestBestPortraitCols(t *testing.T) {
	// ~52 weeks should be a tall grid (not one wide year-row).
	c := weekColumns(52)
	rows := (52 + c - 1) / c
	if c < 4 || c > 10 || rows < 5 {
		t.Fatalf("52 weeks → %d×%d (want portrait)", c, rows)
	}
	// ~12 months → squat 12-wide is avoided
	mc := monthColumns(12)
	if mc > 6 || (12+mc-1)/mc < 2 {
		t.Fatalf("12 months → cols=%d", mc)
	}
	// Tiny week counts stay compact
	if c := weekColumns(6); c > 6 {
		t.Fatalf("6 weeks → cols=%d", c)
	}
}
