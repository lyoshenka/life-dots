package render_test

import (
	"testing"
	"time"

	"github.com/grin/life-dots/internal/calendar"
	"github.com/grin/life-dots/internal/model"
	"github.com/grin/life-dots/internal/render"
)

func TestDotGapIndependentOfSize(t *testing.T) {
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.Local)
	cfg := model.DefaultConfig()
	cfg.Width, cfg.Height = 1206, 2622
	cfg.DeviceID = "iphone-16-pro"
	cfg.DotSize = 24
	cfg.DotGap = 3
	g, err := calendar.Build(cfg, now)
	if err != nil {
		t.Fatal(err)
	}
	m1 := render.ComputeLayout(cfg, g)
	cfg.DotGap = 12
	m2 := render.ComputeLayout(cfg, g)
	if m1.DotSize != 24 || m2.DotSize != 24 {
		t.Fatalf("dot size changed with gap: %.1f -> %.1f", m1.DotSize, m2.DotSize)
	}
	if m2.Gap <= m1.Gap {
		t.Fatalf("gap did not increase: %.1f -> %.1f", m1.Gap, m2.Gap)
	}
}

func TestDotSizeIndependentOfGap(t *testing.T) {
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.Local)
	cfg := model.DefaultConfig()
	cfg.Width, cfg.Height = 1206, 2622
	cfg.DeviceID = "iphone-16-pro"
	cfg.DotGap = 3
	cfg.DotSize = 16
	g, err := calendar.Build(cfg, now)
	if err != nil {
		t.Fatal(err)
	}
	m1 := render.ComputeLayout(cfg, g)
	cfg.DotSize = 28
	m2 := render.ComputeLayout(cfg, g)
	if m2.DotSize <= m1.DotSize {
		t.Fatalf("dot size did not increase: %.1f -> %.1f", m1.DotSize, m2.DotSize)
	}
	if m1.Gap != 3 || m2.Gap != 3 {
		t.Fatalf("gap changed with size: %.1f -> %.1f", m1.Gap, m2.Gap)
	}
}

func TestPreviewScaleKeepsDotProportions(t *testing.T) {
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.Local)
	cfg := model.DefaultConfig()
	cfg.DeviceID = "iphone-16-pro"
	cfg.Width, cfg.Height = 1206, 2622
	cfg.DotSize = 24
	cfg.DotGap = 6
	g, err := calendar.Build(cfg, now)
	if err != nil {
		t.Fatal(err)
	}
	full := render.ComputeLayout(cfg, g)

	cfg.Width, cfg.Height = 302, 656 // ~1/4 scale preview
	preview := render.ComputeLayout(cfg, g)
	ratio := preview.DotSize / full.DotSize
	gapRatio := preview.Gap / full.Gap
	if ratio < 0.2 || ratio > 0.3 {
		t.Fatalf("preview dot scale weird: full=%.1f preview=%.1f ratio=%.3f", full.DotSize, preview.DotSize, ratio)
	}
	if gapRatio < 0.2 || gapRatio > 0.3 {
		t.Fatalf("preview gap scale weird: full=%.1f preview=%.1f ratio=%.3f", full.Gap, preview.Gap, gapRatio)
	}
}
