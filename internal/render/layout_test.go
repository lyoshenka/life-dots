package render

import (
	"testing"
	"time"

	"github.com/grin/life-dots/internal/calendar"
	"github.com/grin/life-dots/internal/model"
)

func TestTitleClearsTopAxisBand(t *testing.T) {
	cases := []struct {
		name  string
		mode  model.Mode
		title string
	}{
		{"year custom", model.ModeYear, "MY YEAR IN DAYS"},
		{"year default", model.ModeYear, ""},
		{"life custom", model.ModeLife, "MY LIFE IN WEEKS"},
		{"monthly", model.ModeMonthly, "OCTOBER"},
	}
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.Local)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := model.DefaultConfig()
			cfg.Width, cfg.Height = 1206, 2622
			cfg.Mode = tc.mode
			cfg.ShowAxes = true
			cfg.ShowTitle = true
			cfg.CustomTitle = tc.title
			cfg.Birthday = "1995-01-01"
			cfg.Year = 2026
			g, err := calendar.Build(cfg, now)
			if err != nil {
				t.Fatal(err)
			}
			if g.Title == "" {
				t.Skip("no title")
			}
			m := ComputeLayout(cfg, g)
			titleBottom := m.TitleY + m.TitleSize
			axisBandTop := m.GridY
			if g.Axes != nil && cfg.ShowAxes {
				band := 0.0
				if len(g.Axes.Top) > 0 {
					band = m.AxisSize * 2
				}
				if g.Axes.TopTitle != "" {
					if v := m.AxisSize * 3.4; v > band {
						band = v
					}
				}
				axisBandTop = m.GridY - band
			}
			if titleBottom > axisBandTop-1 {
				t.Fatalf("title overlaps axis band: titleBottom=%.1f axisBandTop=%.1f GridY=%.1f",
					titleBottom, axisBandTop, m.GridY)
			}
			if g.Axes != nil && g.Axes.TopTitle != "" {
				topTitleY := m.GridY - m.AxisSize*2.5
				if titleBottom > topTitleY-m.AxisSize*0.5 {
					t.Fatalf("title overlaps TopTitle: titleBottom=%.1f topTitleY=%.1f",
						titleBottom, topTitleY)
				}
			}
		})
	}
}
