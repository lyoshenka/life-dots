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
	cfg := model.DefaultConfig()
	cfg.Mode = model.ModeCustom
	cfg.StartDate = "2026-01-01"
	cfg.EndDate = "2026-01-20"
	g, err := Build(cfg, time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if g.Unit != "day" || g.Columns != 7 {
		t.Fatalf("unit=%s cols=%d", g.Unit, g.Columns)
	}
}
