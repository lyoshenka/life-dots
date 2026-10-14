package render

import (
	"math"

	"github.com/grin/life-dots/internal/calendar"
	"github.com/grin/life-dots/internal/model"
)

// LayoutMetrics holds computed pixel layout for grid + chrome.
type LayoutMetrics struct {
	Width, Height       int
	ContentX, ContentY  float64
	ContentW, ContentH  float64
	GridX, GridY        float64
	DotSize, Gap        float64
	TitleY, FooterY     float64
	TitleSize, FootSize float64
	AxisSize            float64
	CardRect            struct{ X, Y, W, H float64 }
}

func gapStride(count int, size, gap float64, g *calendar.GroupGap) float64 {
	if count <= 0 {
		return 0
	}
	if g == nil || g.Every <= 0 || g.Every >= count {
		return float64(count)*size + float64(count-1)*gap
	}
	every := g.Every
	extra := g.Gap * gap
	full := count / every
	rem := count % every
	w := float64(full) * (float64(every)*size + float64(every-1)*gap)
	if full > 0 {
		w += float64(full-1) * extra
	}
	if rem > 0 {
		if full > 0 {
			w += extra
		}
		w += float64(rem)*size + float64(rem-1)*gap
	}
	return w
}

func cellOffset(index int, size, gap float64, g *calendar.GroupGap) float64 {
	if g == nil || g.Every <= 0 {
		return float64(index) * (size + gap)
	}
	every := g.Every
	extra := g.Gap * gap
	groups := index / every
	within := index % every
	return float64(groups)*(float64(every)*(size+gap)-gap+extra) + float64(within)*(size+gap)
}

// ComputeLayout derives padding and dot sizes for the wallpaper.
// Side inset is Margin (px); vertical inset is MarginTop (px).
// The grid always fills the available width; title stays above and footer below.
func ComputeLayout(cfg model.Config, grid *calendar.Grid) LayoutMetrics {
	w, h := cfg.Width, cfg.Height
	if w <= 0 {
		w = 1170
	}
	if h <= 0 {
		h = 2532
	}

	padSide := cfg.Margin
	if padSide <= 0 {
		padSide = 48
	}
	padTop := cfg.MarginTop
	if padTop <= 0 {
		padTop = 48
	}
	// Bottom uses side margin so increasing top margin pushes the grid down
	// instead of shrinking both ends equally.
	padBot := padSide

	if padSide >= float64(w)/2 {
		padSide = float64(w)/2 - 8
	}
	if padTop+padBot >= float64(h)-16 {
		padTop = math.Max(0, float64(h)-padBot-16)
	}

	contentW := float64(w) - 2*padSide
	contentH := float64(h) - padTop - padBot
	if contentW < 8 {
		contentW = 8
	}
	if contentH < 8 {
		contentH = 8
	}

	shapeScale := cfg.ShapeScale / 100
	if shapeScale == 0 {
		shapeScale = 1
	}

	titleSize := math.Max(14, contentW*0.045) * (cfg.StatsScale / 100)
	if cfg.StatsScale == 0 {
		titleSize = math.Max(14, contentW*0.045)
	}
	footSize := math.Max(12, contentW*0.032) * (cfg.StatsScale / 100)
	if cfg.StatsScale == 0 {
		footSize = math.Max(12, contentW*0.032)
	}
	axisSize := math.Max(8, contentW*0.022)

	hasTitle := grid.Title != ""
	hasFooter := cfg.FooterMode != model.FooterNone

	chromeTop := titleSize * 0.4
	if hasTitle {
		chromeTop = titleSize * 2.4
	}
	chromeBot := footSize * 0.4
	if hasFooter {
		chromeBot = footSize * 2.6
	}

	axisLeft := 0.0
	axisTop := 0.0
	if grid.Axes != nil && cfg.ShowAxes {
		if len(grid.Axes.Left) > 0 || len(grid.Axes.LeftIcons) > 0 {
			axisLeft = axisSize * 3
		}
		// Top number labels sit at ~1.2×axisSize; TopTitle at ~2.5×.
		if len(grid.Axes.Top) > 0 {
			axisTop = axisSize * 2
		}
		if grid.Axes.TopTitle != "" {
			axisTop = math.Max(axisTop, axisSize*3.4)
		}
		for _, pl := range grid.Axes.Placed {
			if pl.Row == 0 {
				axisTop = math.Max(axisTop, axisSize*2)
				break
			}
		}
	}

	availW := contentW - axisLeft
	availH := contentH - chromeTop - chromeBot - axisTop
	if availW < 8 {
		availW = 8
	}
	if availH < 8 {
		availH = 8
	}

	cols := grid.Columns
	rows := grid.Rows
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}

	var colGap, rowGap *calendar.GroupGap
	if grid.Grouping != nil {
		colGap = grid.Grouping.Columns
		rowGap = grid.Grouping.Rows
	}

	gapRatio := 0.35
	if cfg.NoGap {
		gapRatio = 0
	}

	// Size for full available width, then clamp by height if needed.
	dot := maxDotForWidth(cols, availW, gapRatio, colGap)
	gap := dot * gapRatio
	gridW := gapStride(cols, dot, gap, colGap)
	gridH := gapStride(rows, dot, gap, rowGap)
	if gridH > availH && gridH > 0 {
		fit := availH / gridH
		dot *= fit
		gap *= fit
		gridW = gapStride(cols, dot, gap, colGap)
		gridH = gapStride(rows, dot, gap, rowGap)
	}

	// Shape scale changes mark size relative to gaps (re-fit width after).
	if shapeScale != 1 {
		dot *= shapeScale
		gridW = gapStride(cols, dot, gap, colGap)
		gridH = gapStride(rows, dot, gap, rowGap)
		if gridW > availW || gridH > availH {
			fit := math.Min(availW/gridW, availH/gridH)
			if fit < 1 {
				dot *= fit
				gap *= fit
				gridW = gapStride(cols, dot, gap, colGap)
				gridH = gapStride(rows, dot, gap, rowGap)
			}
		}
	}

	gridRegionTop := padTop + chromeTop + axisTop
	gridRegionBot := padTop + contentH - chromeBot
	gridRegionH := gridRegionBot - gridRegionTop

	gridX := padSide + axisLeft + (availW-gridW)/2
	gridY := gridRegionTop + (gridRegionH-gridH)/2

	// Keep the title baseline above the reserved axis band (and TopTitle).
	titleGap := titleSize * 0.35
	titleY := gridY - axisTop - titleGap - titleSize
	if titleY < padTop {
		titleY = padTop
	}
	// If the grid sits low, don't let a clamped title drop into the axis band.
	maxTitleY := gridY - axisTop - titleGap - titleSize
	if hasTitle && titleY > maxTitleY {
		titleY = maxTitleY
	}
	footerY := gridY + gridH + footSize*1.6
	maxFooter := padTop + contentH - footSize*0.2
	if footerY > maxFooter {
		footerY = maxFooter
	}

	m := LayoutMetrics{
		Width: w, Height: h,
		ContentX: padSide, ContentY: padTop,
		ContentW: contentW, ContentH: contentH,
		GridX: gridX, GridY: gridY,
		DotSize: dot, Gap: gap,
		TitleY: titleY,
		FooterY: footerY,
		TitleSize: titleSize, FootSize: footSize, AxisSize: axisSize,
	}

	cardPadH := cfg.CardPadH
	cardPadV := cfg.CardPadV
	if cardPadH == 0 {
		cardPadH = 24
	}
	if cardPadV == 0 {
		cardPadV = 32
	}
	m.CardRect.X = gridX - cardPadH - axisLeft
	m.CardRect.Y = gridY - cardPadV - axisTop
	m.CardRect.W = gridW + 2*cardPadH + axisLeft
	m.CardRect.H = gridH + 2*cardPadV + axisTop
	return m
}

// maxDotForWidth finds the largest dot size whose grid width equals availW.
func maxDotForWidth(cols int, availW, gapRatio float64, colGap *calendar.GroupGap) float64 {
	lo, hi := 0.5, availW
	dot := lo
	for i := 0; i < 28; i++ {
		mid := (lo + hi) / 2
		g := mid * gapRatio
		gw := gapStride(cols, mid, g, colGap)
		if gw <= availW {
			dot = mid
			lo = mid
		} else {
			hi = mid
		}
	}
	return dot
}
