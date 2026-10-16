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

// ConfigPxScale maps studio/device pixel values onto the render canvas.
// Preview (and any downscaled width) must shrink dots/gaps/margins with the image.
func ConfigPxScale(cfg model.Config) float64 {
	if cfg.Width <= 0 {
		return 1
	}
	if d := model.DeviceByID(cfg.DeviceID); d != nil && d.Width > 0 {
		s := float64(cfg.Width) / float64(d.Width)
		if s > 0 {
			return s
		}
	}
	return 1
}

// ComputeLayout derives padding and dot sizes for the wallpaper.
// Side inset is Margin (px); vertical inset is MarginTop (px).
// DotSize and DotGap are independent device pixels (scaled to the canvas).
func ComputeLayout(cfg model.Config, grid *calendar.Grid) LayoutMetrics {
	w, h := cfg.Width, cfg.Height
	if w <= 0 {
		w = 1170
	}
	if h <= 0 {
		h = 2532
	}

	px := ConfigPxScale(cfg)

	padSide := cfg.Margin * px
	if cfg.Margin <= 0 {
		padSide = 48 * px
	}
	padTop := cfg.MarginTop * px
	if cfg.MarginTop <= 0 {
		padTop = 48 * px
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

	titleSize := cfg.TitleSize * px
	if cfg.TitleSize <= 0 {
		titleSize = math.Max(14, contentW*0.045)
		if cfg.StatsScale > 0 {
			titleSize *= cfg.StatsScale / 100
		}
	}
	footSize := cfg.FooterSize * px
	if cfg.FooterSize <= 0 {
		footSize = math.Max(12, contentW*0.032)
		if cfg.StatsScale > 0 {
			footSize *= cfg.StatsScale / 100
		}
	}
	axisSize := math.Max(8*px, contentW*0.022)

	hasTitle := grid.Title != ""
	hasFooter := cfg.ShowFooter

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

	reqGap := cfg.DotGap * px
	if reqGap < 0 {
		reqGap = 0
	}
	if cfg.NoGap {
		reqGap = 0
	}
	reqDot := cfg.DotSize * px

	dot, gap := fitDotGap(cols, rows, availW, availH, reqDot, reqGap, colGap, rowGap)
	gridW := gapStride(cols, dot, gap, colGap)
	gridH := gapStride(rows, dot, gap, rowGap)

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

	cardPadH := cfg.CardPadH * px
	cardPadV := cfg.CardPadV * px
	if cardPadH < 0 {
		cardPadH = 0
	}
	if cardPadV < 0 {
		cardPadV = 0
	}
	// Card wraps the full content block: title, axes, grid, and footer.
	contentLeft := gridX - axisLeft
	contentRight := gridX + gridW
	contentTop := gridY - axisTop
	if hasTitle {
		contentTop = math.Min(contentTop, titleY)
	}
	contentBot := gridY + gridH
	if hasFooter {
		// FooterY is the text baseline; include a little descent.
		contentBot = math.Max(contentBot, footerY+footSize*0.25)
	}
	m.CardRect.X = contentLeft - cardPadH
	m.CardRect.Y = contentTop - cardPadV
	m.CardRect.W = (contentRight - contentLeft) + 2*cardPadH
	m.CardRect.H = (contentBot - contentTop) + 2*cardPadV
	return m
}

func gridFits(cols, rows int, availW, availH, dot, gap float64, colGap, rowGap *calendar.GroupGap) bool {
	return gapStride(cols, dot, gap, colGap) <= availW+1e-6 &&
		gapStride(rows, dot, gap, rowGap) <= availH+1e-6
}

// fitDotGap keeps mark size and gap independent.
// When both won't fit, gap shrinks first; size only shrinks if even gap=0 is too tight.
// reqDot <= 0 means "auto": largest size that still honors reqGap (gap still wins over fill).
func fitDotGap(cols, rows int, availW, availH, reqDot, reqGap float64, colGap, rowGap *calendar.GroupGap) (dot, gap float64) {
	if reqGap < 0 {
		reqGap = 0
	}
	if reqDot <= 0 {
		// Auto size: fill space at the requested gap. Not used by the studio defaults.
		return maxDotForBox(cols, rows, availW, availH, reqGap, colGap, rowGap), reqGap
	}
	dot, gap = reqDot, reqGap
	if gridFits(cols, rows, availW, availH, dot, gap, colGap, rowGap) {
		return dot, gap
	}
	// Prefer keeping mark size; shrink spacing until it fits.
	gap = maxGapForBox(cols, rows, availW, availH, dot, colGap, rowGap)
	if gap < 0 {
		gap = 0
	}
	if gridFits(cols, rows, availW, availH, dot, gap, colGap, rowGap) {
		return dot, gap
	}
	// Size itself is too large even with no gap — cap size, keep gap at 0.
	// (Do not re-introduce gap here: that would make spacing change mark size.)
	return maxDotForBox(cols, rows, availW, availH, 0, colGap, rowGap), 0
}

// maxDotForBox finds the largest mark size that fits availW×availH with a fixed gap.
func maxDotForBox(cols, rows int, availW, availH, gap float64, colGap, rowGap *calendar.GroupGap) float64 {
	lo, hi := 0.5, math.Min(availW, availH)
	if hi < lo {
		hi = lo
	}
	dot := lo
	for i := 0; i < 28; i++ {
		mid := (lo + hi) / 2
		if gridFits(cols, rows, availW, availH, mid, gap, colGap, rowGap) {
			dot = mid
			lo = mid
		} else {
			hi = mid
		}
	}
	return dot
}

// maxGapForBox finds the largest gap that fits with a fixed mark size.
// Returns -1 if even gap=0 does not fit.
func maxGapForBox(cols, rows int, availW, availH, dot float64, colGap, rowGap *calendar.GroupGap) float64 {
	if !gridFits(cols, rows, availW, availH, dot, 0, colGap, rowGap) {
		return -1
	}
	lo, hi := 0.0, math.Max(availW, availH)
	best := 0.0
	for i := 0; i < 28; i++ {
		mid := (lo + hi) / 2
		if gridFits(cols, rows, availW, availH, dot, mid, colGap, rowGap) {
			best = mid
			lo = mid
		} else {
			hi = mid
		}
	}
	return best
}
