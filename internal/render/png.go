package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"time"

	"github.com/grin/life-dots/internal/calendar"
	"github.com/grin/life-dots/internal/model"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// RenderPNG builds a wallpaper PNG. photo is optional request-scoped POST body image.
func RenderPNG(cfg model.Config, photo image.Image, now time.Time) ([]byte, error) {
	if now.IsZero() {
		now = time.Now()
	}
	grid, err := calendar.Build(cfg, now)
	if err != nil {
		return nil, err
	}
	colors := model.ResolveColors(cfg)
	m := ComputeLayout(cfg, grid)

	bg, err := PrepareBackground(cfg, m.Width, m.Height, photo)
	if err != nil {
		return nil, err
	}

	if cfg.ShowCard {
		r := image.Rect(int(m.CardRect.X), int(m.CardRect.Y), int(m.CardRect.X+m.CardRect.W), int(m.CardRect.Y+m.CardRect.H))
		DrawGlassCard(bg, r, cfg)
	}

	gridAlpha := cfg.GridOpacity / 100
	if gridAlpha == 0 {
		gridAlpha = 1
	}

	var colGap, rowGap *calendar.GroupGap
	if grid.Grouping != nil {
		colGap = grid.Grouping.Columns
		rowGap = grid.Grouping.Rows
	}

	past := parseHex(colors.Past)
	curr := parseHex(colors.Current)
	future := parseHex(colors.Future)
	textC := parseHex(colors.Text)
	muted := parseHex(colors.TextMuted)

	for i, cell := range grid.Cells {
		if cell.State == calendar.CellEmpty {
			continue
		}
		col := i % grid.Columns
		row := i / grid.Columns
		x := m.GridX + cellOffset(col, m.DotSize, m.Gap, colGap)
		y := m.GridY + cellOffset(row, m.DotSize, m.Gap, rowGap)

		c := future
		switch cell.State {
		case calendar.CellPast:
			c = past
		case calendar.CellCurrent:
			c = curr
		case calendar.CellSpecial:
			if cell.Color != "" {
				c = parseHex(cell.Color)
			} else {
				c = past
			}
		}
		op := cell.Opacity * gridAlpha
		if op <= 0 {
			op = gridAlpha
		}
		c.A = uint8(float64(c.A) * op)

		outlined := cfg.ShapeStyle == model.StyleOutlined
		fillPast := cfg.OutlinedFillPast
		isCurrent := cell.State == calendar.CellCurrent || i == grid.Current

		if outlined {
			if cell.State == calendar.CellPast && fillPast {
				drawShape(bg, cfg, x, y, m.DotSize, c, true)
			} else if isCurrent {
				drawShape(bg, cfg, x, y, m.DotSize, c, false)
				ring := c
				ring.A = 255
				drawShapeRing(bg, cfg, x, y, m.DotSize, ring, math.Max(1.5, m.DotSize*0.12))
			} else {
				drawShape(bg, cfg, x, y, m.DotSize, c, false)
			}
		} else {
			drawShape(bg, cfg, x, y, m.DotSize, c, true)
			if isCurrent && cfg.HighlightToday {
				ring := curr
				ring.A = 255
				pad := m.DotSize * 0.08
				drawShapeRing(bg, cfg, x-pad, y-pad, m.DotSize+pad*2, ring, math.Max(1.2, m.DotSize*0.1))
			}
		}

		if (cfg.ShowNumbers || cfg.ShowBoth) && cell.Label != "" {
			drawCenteredLabel(bg, x, y, m.DotSize, cell.Label, textC)
		}
	}

	// Axis labels
	if grid.Axes != nil && cfg.ShowAxes {
		drawAxes(bg, grid, m, muted)
	}

	// Title
	if grid.Title != "" {
		drawText(bg, m.Width/2, int(m.TitleY+m.TitleSize), grid.Title, textC, true)
	}

	// Footer stats / quote
	footer := buildFooter(cfg, grid, now)
	if footer != "" {
		drawText(bg, m.Width/2, int(m.FooterY), footer, muted, true)
		// Dual-tone: if "left · right", color left with accent
		if strings.Contains(footer, " · ") && cfg.ShowProgress {
			parts := strings.SplitN(footer, " · ", 2)
			// redraw left in accent roughly at center-left — simple approach: already drawn; leave as is for clarity
			_ = parts
			_ = curr
		}
	}

	if cfg.ShowBranding {
		drawText(bg, m.Width/2, m.Height-int(float64(m.Height)*0.04), "life-dots", muted, true)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, bg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildFooter(cfg model.Config, grid *calendar.Grid, now time.Time) string {
	if cfg.FooterMode == model.FooterNone {
		return ""
	}
	if grid.StatsLeft == "" {
		return ""
	}
	if grid.StatsRight != "" {
		return grid.StatsLeft + " · " + grid.StatsRight
	}
	return grid.StatsLeft
}

func drawShape(dst *image.RGBA, cfg model.Config, x, y, size float64, c color.RGBA, filled bool) {
	shape := cfg.DotShape
	switch shape {
	case model.ShapeSquare:
		drawSDFRect(dst, x, y, size, size, 0, c, filled)
	case model.ShapeRounded:
		drawSDFRect(dst, x, y, size, size, size*0.25, c, filled)
	default:
		drawSDFCircle(dst, x+size/2, y+size/2, size/2, c, filled)
	}
}

// drawShapeRing draws an outline matching the configured dot shape (today highlight).
func drawShapeRing(dst *image.RGBA, cfg model.Config, x, y, size float64, c color.RGBA, thickness float64) {
	switch cfg.DotShape {
	case model.ShapeSquare:
		drawSDFOutline(dst, x, y, size, size, 0, c, thickness)
	case model.ShapeRounded:
		drawSDFOutline(dst, x, y, size, size, size*0.25, c, thickness)
	default:
		drawSDFOutline(dst, x, y, size, size, size/2, c, thickness) // circle = fully rounded square
	}
}

func drawSDFOutline(dst *image.RGBA, x, y, w, h, radius float64, c color.RGBA, thickness float64) {
	half := thickness / 2
	pad := half + 1.5
	minX := int(x - pad)
	maxX := int(x + w + pad)
	minY := int(y - pad)
	maxY := int(y + h + pad)
	b := dst.Bounds()
	for py := minY; py <= maxY; py++ {
		for px := minX; px <= maxX; px++ {
			if px < b.Min.X || px >= b.Max.X || py < b.Min.Y || py >= b.Max.Y {
				continue
			}
			d := math.Abs(sdfRoundRect(float64(px)+0.5, float64(py)+0.5, x, y, w, h, radius))
			if d < half+1 {
				a := 1.0
				if d > half {
					a = 1 - (d - half)
				}
				cc := c
				cc.A = uint8(float64(c.A) * a)
				dst.SetRGBA(px, py, alphaOver(dst.RGBAAt(px, py), cc))
			}
		}
	}
}

func drawSDFCircle(dst *image.RGBA, cx, cy, r float64, c color.RGBA, filled bool) {
	minX := int(cx - r - 1)
	maxX := int(cx + r + 1)
	minY := int(cy - r - 1)
	maxY := int(cy + r + 1)
	b := dst.Bounds()
	for py := minY; py <= maxY; py++ {
		for px := minX; px <= maxX; px++ {
			if px < b.Min.X || px >= b.Max.X || py < b.Min.Y || py >= b.Max.Y {
				continue
			}
			dx := float64(px) + 0.5 - cx
			dy := float64(py) + 0.5 - cy
			d := math.Hypot(dx, dy) - r
			if filled {
				if d < 1 {
					a := 1.0
					if d > 0 {
						a = 1 - d
					}
					cc := c
					cc.A = uint8(float64(c.A) * a)
					dst.SetRGBA(px, py, alphaOver(dst.RGBAAt(px, py), cc))
				}
			} else {
				// outline
				ad := math.Abs(d)
				if ad < 1.2 {
					a := 1 - ad/1.2
					cc := c
					cc.A = uint8(float64(c.A) * a)
					dst.SetRGBA(px, py, alphaOver(dst.RGBAAt(px, py), cc))
				}
			}
		}
	}
}

func drawSDFRect(dst *image.RGBA, x, y, w, h, radius float64, c color.RGBA, filled bool) {
	minX := int(x - 1)
	maxX := int(x + w + 1)
	minY := int(y - 1)
	maxY := int(y + h + 1)
	b := dst.Bounds()
	for py := minY; py <= maxY; py++ {
		for px := minX; px <= maxX; px++ {
			if px < b.Min.X || px >= b.Max.X || py < b.Min.Y || py >= b.Max.Y {
				continue
			}
			d := sdfRoundRect(float64(px)+0.5, float64(py)+0.5, x, y, w, h, radius)
			if filled {
				if d < 1 {
					a := 1.0
					if d > 0 {
						a = 1 - d
					}
					cc := c
					cc.A = uint8(float64(c.A) * a)
					dst.SetRGBA(px, py, alphaOver(dst.RGBAAt(px, py), cc))
				}
			} else if math.Abs(d) < 1.2 {
				a := 1 - math.Abs(d)/1.2
				cc := c
				cc.A = uint8(float64(c.A) * a)
				dst.SetRGBA(px, py, alphaOver(dst.RGBAAt(px, py), cc))
			}
		}
	}
}

// drawPathApprox rasterizes a simple normalized SVG path as filled polygon samples.
func drawPathApprox(dst *image.RGBA, path string, x, y, size float64, c color.RGBA) {
	// Sample bounding box with point-in-polygon via SDF-ish: evaluate many points
	// For simplicity, stamp as circle if path parse is heavy — use densified fill of path commands M/L/C/Z
	pts := parsePathPoints(path, x, y, size)
	if len(pts) < 3 {
		drawSDFCircle(dst, x+size/2, y+size/2, size/2, c, true)
		return
	}
	minX, minY := pts[0].X, pts[0].Y
	maxX, maxY := minX, minY
	for _, p := range pts {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	b := dst.Bounds()
	for py := int(minY); py <= int(maxY); py++ {
		for px := int(minX); px <= int(maxX); px++ {
			if px < b.Min.X || px >= b.Max.X || py < b.Min.Y || py >= b.Max.Y {
				continue
			}
			if pointInPoly(float64(px)+0.5, float64(py)+0.5, pts) {
				dst.SetRGBA(px, py, alphaOver(dst.RGBAAt(px, py), c))
			}
		}
	}
}

type pt struct{ X, Y float64 }

func parsePathPoints(path string, ox, oy, scale float64) []pt {
	// Very small parser for M/L/C/Z with absolute 0–1 coords
	tokens := tokenizePath(path)
	var pts []pt
	var cx, cy float64
	i := 0
	for i < len(tokens) {
		cmd := tokens[i]
		i++
		switch cmd {
		case "M", "L":
			if i+1 >= len(tokens) {
				break
			}
			x, y := atof(tokens[i]), atof(tokens[i+1])
			i += 2
			cx, cy = ox+x*scale, oy+y*scale
			pts = append(pts, pt{cx, cy})
		case "C":
			if i+5 >= len(tokens) {
				break
			}
			// approximate cubic with 6 samples
			x1, y1 := atof(tokens[i]), atof(tokens[i+1])
			x2, y2 := atof(tokens[i+2]), atof(tokens[i+3])
			x3, y3 := atof(tokens[i+4]), atof(tokens[i+5])
			i += 6
			x0, y0 := (cx-ox)/scale, (cy-oy)/scale
			for t := 1; t <= 6; t++ {
				tt := float64(t) / 6
				u := 1 - tt
				bx := u*u*u*x0 + 3*u*u*tt*x1 + 3*u*tt*tt*x2 + tt*tt*tt*x3
				by := u*u*u*y0 + 3*u*u*tt*y1 + 3*u*tt*tt*y2 + tt*tt*tt*y3
				cx, cy = ox+bx*scale, oy+by*scale
				pts = append(pts, pt{cx, cy})
			}
		case "Z", "z":
			if len(pts) > 0 {
				pts = append(pts, pts[0])
			}
		case "A":
			// skip arc params (7 numbers) — sample endpoint
			if i+6 < len(tokens) {
				x, y := atof(tokens[i+5]), atof(tokens[i+6])
				i += 7
				cx, cy = ox+x*scale, oy+y*scale
				pts = append(pts, pt{cx, cy})
			}
		default:
			// bare numbers after command sometimes
			i++
		}
	}
	return pts
}

func tokenizePath(s string) []string {
	var out []string
	cur := strings.Builder{}
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		switch {
		case r == ',' || r == ' ' || r == '\t' || r == '\n':
			flush()
		case (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'):
			flush()
			out = append(out, string(r))
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return out
}

func atof(s string) float64 {
	var f float64
	_, _ = parseFloat(s, &f)
	return f
}

func parseFloat(s string, f *float64) (int, error) {
	n, err := fmtSscanf(s, f)
	return n, err
}

// tiny wrapper to avoid importing fmt in hot path file top — use strconv
func fmtSscanf(s string, f *float64) (int, error) {
	var neg bool
	i := 0
	if i < len(s) && s[i] == '-' {
		neg = true
		i++
	}
	var v float64
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		v = v*10 + float64(s[i]-'0')
		i++
	}
	if i < len(s) && s[i] == '.' {
		i++
		base := 0.1
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			v += float64(s[i]-'0') * base
			base *= 0.1
			i++
		}
	}
	if neg {
		v = -v
	}
	*f = v
	return 1, nil
}

func pointInPoly(x, y float64, pts []pt) bool {
	n := len(pts)
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := pts[i].X, pts[i].Y
		xj, yj := pts[j].X, pts[j].Y
		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi+1e-12)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
}

func drawAxes(dst *image.RGBA, grid *calendar.Grid, m LayoutMetrics, c color.RGBA) {
	ax := grid.Axes
	if ax == nil {
		return
	}
	var colGap, rowGap *calendar.GroupGap
	if grid.Grouping != nil {
		colGap = grid.Grouping.Columns
		rowGap = grid.Grouping.Rows
	}
	for i, label := range ax.Top {
		pos := i
		if i < len(ax.TopPos) {
			pos = ax.TopPos[i]
		}
		x := m.GridX + cellOffset(pos, m.DotSize, m.Gap, colGap) + m.DotSize/2
		y := m.GridY - m.AxisSize*1.2
		drawText(dst, int(x), int(y), label, c, true)
	}
	if ax.TopTitle != "" {
		drawText(dst, int(m.GridX+gapStride(grid.Columns, m.DotSize, m.Gap, colGap)/2), int(m.GridY-m.AxisSize*2.5), ax.TopTitle, c, true)
	}
	for _, pl := range ax.Placed {
		x := m.GridX + cellOffset(pl.Col, m.DotSize, m.Gap, colGap) + m.DotSize/2
		y := m.GridY + cellOffset(pl.Row, m.DotSize, m.Gap, rowGap) - m.AxisSize*1.15
		drawText(dst, int(x), int(y), pl.Text, c, true)
	}
	for i, label := range ax.Left {
		pos := i
		if i < len(ax.LeftPos) {
			pos = ax.LeftPos[i]
		}
		x := m.GridX - m.AxisSize*1.5
		y := m.GridY + cellOffset(pos, m.DotSize, m.Gap, rowGap) + m.DotSize/2
		drawText(dst, int(x), int(y), label, c, true)
	}
	for i, icon := range ax.LeftIcons {
		pos := i
		if i < len(ax.LeftIconPos) {
			pos = ax.LeftIconPos[i]
		}
		x := m.GridX - m.DotSize*1.4
		y := m.GridY + cellOffset(pos, m.DotSize, m.Gap, rowGap)
		if sp := model.ShapeByID(icon); sp != nil {
			drawPathApprox(dst, sp.Path, x, y, m.DotSize*0.9, c)
		}
	}
	if ax.LeftTitle != "" {
		drawText(dst, int(m.GridX-m.AxisSize*2.8), int(m.GridY+gapStride(grid.Rows, m.DotSize, m.Gap, rowGap)/2), ax.LeftTitle, c, true)
	}
}

func drawCenteredLabel(dst *image.RGBA, x, y, size float64, label string, c color.RGBA) {
	drawText(dst, int(x+size/2), int(y+size*0.7), label, c, true)
}

func drawText(dst *image.RGBA, x, y int, s string, c color.RGBA, center bool) {
	face := basicfont.Face7x13
	lines := strings.Split(s, "\n")
	for li, line := range lines {
		w := font.MeasureString(face, line).Ceil()
		xx := x
		if center {
			xx = x - w/2
		}
		yy := y + li*14
		point := fixed.Point26_6{X: fixed.I(xx), Y: fixed.I(yy)}
		d := &font.Drawer{
			Dst:  dst,
			Src:  image.NewUniform(c),
			Face: face,
			Dot:  point,
		}
		d.DrawString(line)
	}
}
