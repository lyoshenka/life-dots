package render

import (
	"testing"
	"time"

	"github.com/grin/life-dots/internal/calendar"
	"github.com/grin/life-dots/internal/model"
)

func TestCardIncludesTitleAndFooter(t *testing.T) {
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.Local)
	cfg := model.DefaultConfig()
	cfg.Width, cfg.Height = 1206, 2622
	cfg.ShowCard = true
	cfg.ShowTitle = true
	cfg.ShowFooter = true
	cfg.CardPadH, cfg.CardPadV = 20, 20
	g, err := calendar.Build(cfg, now)
	if err != nil {
		t.Fatal(err)
	}
	if g.Title == "" {
		t.Fatal("expected title")
	}
	m := ComputeLayout(cfg, g)
	if m.CardRect.Y > m.TitleY+1 {
		t.Fatalf("card top %.1f below title top %.1f", m.CardRect.Y, m.TitleY)
	}
	cardBot := m.CardRect.Y + m.CardRect.H
	if cardBot < m.FooterY-1 {
		t.Fatalf("card bottom %.1f above footer baseline %.1f", cardBot, m.FooterY)
	}
}
