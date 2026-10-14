package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/grin/life-dots/internal/calendar"
	"github.com/grin/life-dots/internal/model"
)

// RenderSVG builds a vector wallpaper (preview / wallpaper.svg).
func RenderSVG(cfg model.Config, now time.Time) (string, error) {
	if now.IsZero() {
		now = time.Now()
	}
	grid, err := calendar.Build(cfg, now)
	if err != nil {
		return "", err
	}
	colors := model.ResolveColors(cfg)
	m := ComputeLayout(cfg, grid)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, m.Width, m.Height, m.Width, m.Height)
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="%s"/>`, colors.Background)

	if cfg.ShowCard {
		op := cfg.CardOpacity / 100
		if op < 0 {
			op = 0
		}
		if op > 1 {
			op = 1
		}
		fill := cfg.CardColor
		if fill == "" {
			fill = "#ffffff"
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" fill="%s" fill-opacity="%.2f"/>`,
			m.CardRect.X, m.CardRect.Y, m.CardRect.W, m.CardRect.H, cfg.CardRadius, fill, op)
	}

	var colGap, rowGap *calendar.GroupGap
	if grid.Grouping != nil {
		colGap = grid.Grouping.Columns
		rowGap = grid.Grouping.Rows
	}
	gridAlpha := cfg.GridOpacity / 100
	if gridAlpha == 0 {
		gridAlpha = 1
	}

	for i, cell := range grid.Cells {
		if cell.State == calendar.CellEmpty {
			continue
		}
		col := i % grid.Columns
		row := i / grid.Columns
		x := m.GridX + cellOffset(col, m.DotSize, m.Gap, colGap)
		y := m.GridY + cellOffset(row, m.DotSize, m.Gap, rowGap)
		fill := colors.Future
		switch cell.State {
		case calendar.CellPast:
			fill = colors.Past
		case calendar.CellCurrent:
			fill = colors.Current
		case calendar.CellSpecial:
			if cell.Color != "" {
				fill = cell.Color
			} else {
				fill = colors.Past
			}
		}
		op := cell.Opacity * gridAlpha
		shape := svgShape(cfg, x, y, m.DotSize, fill, op)
		b.WriteString(shape)
		if (cell.State == calendar.CellCurrent || i == grid.Current) && cfg.HighlightToday {
			sw := mathMax(1, m.DotSize*0.1)
			pad := m.DotSize * 0.08
			b.WriteString(svgShapeOutline(cfg, x-pad, y-pad, m.DotSize+pad*2, colors.Current, sw))
		}
	}

	if grid.Title != "" {
		fmt.Fprintf(&b, `<text x="%d" y="%.1f" text-anchor="middle" fill="%s" font-family="system-ui,sans-serif" font-size="%.1f" font-weight="600">%s</text>`,
			m.Width/2, m.TitleY+m.TitleSize, colors.Text, m.TitleSize, xmlEscape(grid.Title))
	}
	footer := buildFooter(cfg, grid, now)
	if footer != "" {
		footerFill := colors.Text
		if cfg.MuteText {
			footerFill = colors.TextMuted
		}
		fmt.Fprintf(&b, `<text x="%d" y="%.1f" text-anchor="middle" fill="%s" font-family="system-ui,sans-serif" font-size="%.1f">%s</text>`,
			m.Width/2, m.FooterY, footerFill, m.FootSize, xmlEscape(strings.ReplaceAll(footer, "\n", " · ")))
	}
	b.WriteString(`</svg>`)
	return b.String(), nil
}

func svgShape(cfg model.Config, x, y, size float64, fill string, op float64) string {
	switch cfg.DotShape {
	case model.ShapeSquare:
		return fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" opacity="%.2f"/>`, x, y, size, size, fill, op)
	case model.ShapeRounded:
		return fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="%.2f" fill="%s" opacity="%.2f"/>`, x, y, size, size, size/4, fill, op)
	default:
		return fmt.Sprintf(`<circle cx="%.2f" cy="%.2f" r="%.2f" fill="%s" opacity="%.2f"/>`, x+size/2, y+size/2, size/2, fill, op)
	}
}

func svgShapeOutline(cfg model.Config, x, y, size float64, stroke string, sw float64) string {
	switch cfg.DotShape {
	case model.ShapeSquare:
		return fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="none" stroke="%s" stroke-width="%.2f"/>`, x, y, size, size, stroke, sw)
	case model.ShapeRounded:
		return fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="%.2f" fill="none" stroke="%s" stroke-width="%.2f"/>`, x, y, size, size, size/4, stroke, sw)
	default:
		return fmt.Sprintf(`<circle cx="%.2f" cy="%.2f" r="%.2f" fill="none" stroke="%s" stroke-width="%.2f"/>`, x+size/2, y+size/2, size/2, stroke, sw)
	}
}

func mathMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
