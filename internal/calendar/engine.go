package calendar

import (
	"fmt"
	"math"
	"time"

	"github.com/grin/life-dots/internal/model"
)

// CellState describes a single dot.
type CellState int

const (
	CellEmpty  CellState = iota // out of range / padding
	CellPast
	CellCurrent
	CellFuture
	CellSpecial
)

// Cell is one grid position.
type Cell struct {
	State   CellState
	Date    time.Time
	Index   int
	Color   string  // override for special dates
	Opacity float64 // 0–1 for heatmap; 1 default
	Label   string  // number or weekday
	Filled  bool    // habit completion etc.
}

// PlacedLabel is a text label at a grid cell position (drawn above that row).
type PlacedLabel struct {
	Text string
	Col  int
	Row  int
}

// AxisLabels describe optional row/column labels.
type AxisLabels struct {
	Top         []string
	TopPos      []int
	TopTitle    string
	Left        []string
	LeftPos     []int
	LeftTitle   string
	LeftIcons   []string // habit icon IDs
	LeftIconPos []int
	Placed      []PlacedLabel // free-position labels (e.g. month titles)
}

// GroupGap describes periodic gaps.
type GroupGap struct {
	Every int
	Gap   float64 // relative multiplier of gap size
}

// GroupingSpec is column/row gap rules.
type GroupingSpec struct {
	Columns *GroupGap
	Rows    *GroupGap
}

// Grid is the computed calendar layout ready for rendering.
type Grid struct {
	Columns   int
	Rows      int
	Cells     []Cell
	Current   int // index of current cell, -1 if none
	Filled    int
	Total     int
	Title     string
	Subtitle  string
	StatsLeft string
	StatsRight string
	StatsPct  float64
	Axes      *AxisLabels
	Grouping  *GroupingSpec
	Unit      string // "day" or "week"
}

// Build computes a grid from config as of `now`.
func Build(cfg model.Config, now time.Time) (*Grid, error) {
	if now.IsZero() {
		now = time.Now()
	}
	now = truncateDay(now)

	var (
		g   *Grid
		err error
	)
	switch cfg.Mode {
	case model.ModeLife:
		g, err = buildLife(cfg, now)
	case model.ModeYear:
		g, err = buildYearDays(cfg, now)
	case model.ModeYearMonths:
		g, err = buildYearMonths(cfg, now)
	case model.ModeQuarters:
		g, err = buildQuarters(cfg, now)
	case model.ModeMonthly:
		g, err = buildMonthly(cfg, now)
	case model.ModeWeekly:
		g, err = buildWeekly(cfg, now)
	case model.ModeCustom, model.ModeGoal:
		g, err = buildCustom(cfg, now)
	case model.ModeHabits:
		g, err = buildHabits(cfg, now)
	default:
		g, err = buildLife(cfg, now)
	}
	if err != nil || g == nil {
		return g, err
	}
	if !cfg.ShowTitle {
		g.Title = ""
	}
	return g, nil
}

func truncateDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

func weeksBetween(a, b time.Time) int {
	a = startOfWeek(a, time.Monday)
	b = startOfWeek(b, time.Monday)
	days := int(b.Sub(a).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days / 7
}

func startOfWeek(t time.Time, weekStart time.Weekday) time.Time {
	t = truncateDay(t)
	wd := t.Weekday()
	offset := (int(wd) - int(weekStart) + 7) % 7
	return t.AddDate(0, 0, -offset)
}

func weekStartDay(cfg model.Config) time.Weekday {
	if cfg.WeekStart == model.WeekStartSunday {
		return time.Sunday
	}
	return time.Monday
}

func applySpecialAndHeat(cfg model.Config, cells []Cell) {
	for i := range cells {
		if cells[i].State == CellEmpty {
			continue
		}
		d := cells[i].Date
		if d.IsZero() {
			continue
		}
		key := d.Format("2006-01-02")
		for _, sd := range cfg.SpecialDates {
			start, err1 := parseDate(sd.StartDate)
			end, err2 := parseDate(sd.EndDate)
			if err1 != nil || err2 != nil {
				continue
			}
			if !d.Before(start) && !d.After(end) {
				cells[i].Color = sd.Color
				if cells[i].State != CellCurrent {
					cells[i].State = CellSpecial
				}
			}
		}
		if cfg.DailyCounts != nil {
			if c, ok := cfg.DailyCounts[key]; ok && c > 0 {
				// Map count to opacity 0.25–1.0
				op := 0.25 + math.Min(float64(c)/10.0, 0.75)
				cells[i].Opacity = op
			}
		}
		if cells[i].Opacity == 0 {
			cells[i].Opacity = 1
		}
	}
}

func buildLife(cfg model.Config, now time.Time) (*Grid, error) {
	dob, err := parseDate(cfg.Birthday)
	if err != nil {
		return nil, fmt.Errorf("birthday: %w", err)
	}
	lifespan := cfg.Lifespan
	if lifespan < 1 {
		lifespan = 90
	}
	if lifespan > 120 {
		lifespan = 120
	}
	ws := weekStartDay(cfg)
	birthWeek := startOfWeek(dob, ws)
	totalWeeks := lifespan * 52
	lived := weeksBetween(birthWeek, now)
	if lived > totalWeeks {
		lived = totalWeeks
	}
	if lived < 0 {
		lived = 0
	}

	cols, rows := 52, lifespan
	cells := make([]Cell, cols*rows)
	current := -1
	for i := 0; i < totalWeeks; i++ {
		d := birthWeek.AddDate(0, 0, i*7)
		st := CellFuture
		if i < lived {
			st = CellPast
		} else if i == lived {
			st = CellCurrent
			current = i
		}
		cells[i] = Cell{State: st, Date: d, Index: i, Opacity: 1, Filled: st == CellPast || st == CellCurrent}
		if cfg.ShowNumbers || cfg.ShowBoth {
			cells[i].Label = fmt.Sprintf("%d", i+1)
		}
	}

	age := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		age--
	}
	pct := float64(lived) / float64(totalWeeks) * 100
	remaining := totalWeeks - lived

	title := cfg.CustomTitle
	if title == "" {
		title = "My life in weeks"
	}

	g := &Grid{
		Columns:    cols,
		Rows:       rows,
		Cells:      cells,
		Current:    current,
		Filled:     lived,
		Total:      totalWeeks,
		Title:      title,
		Subtitle:   fmt.Sprintf("%d of %d weeks lived (%.0f%%) — %d remaining", lived, totalWeeks, pct, remaining),
		StatsLeft:  fmt.Sprintf("%d lived", lived),
		StatsRight: fmt.Sprintf("%d remaining", remaining),
		StatsPct:   pct,
		Unit:       "week",
	}

	if cfg.ShowAxes {
		left := make([]string, 0)
		leftPos := make([]int, 0)
		for a := 10; a <= lifespan; a += 10 {
			left = append(left, fmt.Sprintf("%d", a))
			leftPos = append(leftPos, a-1)
		}
		g.Axes = &AxisLabels{
			Top: []string{"1", "13", "26", "39", "52"}, TopPos: []int{0, 12, 25, 38, 51}, TopTitle: "WEEK >",
			Left: left, LeftPos: leftPos, LeftTitle: "< AGE",
		}
	}

	g.Grouping = lifeGrouping(cfg)
	applySpecialAndHeat(cfg, cells)
	return g, nil
}

func lifeGrouping(cfg model.Config) *GroupingSpec {
	switch cfg.Grouping {
	case model.GroupingNone:
		return nil
	case model.GroupingMonths, model.GroupingMonth:
		return &GroupingSpec{Columns: &GroupGap{Every: 4, Gap: 1.5}}
	case model.GroupingDecades, model.GroupingDecade:
		return &GroupingSpec{Rows: &GroupGap{Every: 10, Gap: 2}}
	case model.GroupingFull:
		return &GroupingSpec{
			Columns: &GroupGap{Every: 4, Gap: 1.5},
			Rows:    &GroupGap{Every: 10, Gap: 2},
		}
	default:
		return &GroupingSpec{
			Columns: &GroupGap{Every: 4, Gap: 1.5},
			Rows:    &GroupGap{Every: 10, Gap: 2},
		}
	}
}

func buildYearDays(cfg model.Config, now time.Time) (*Grid, error) {
	year := cfg.Year
	if year == 0 {
		year = now.Year()
	}
	start := time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
	end := time.Date(year, 12, 31, 0, 0, 0, 0, now.Location())
	days := int(end.Sub(start).Hours()/24) + 1
	cols := 14
	rows := int(math.Ceil(float64(days) / float64(cols)))
	cells := make([]Cell, cols*rows)
	current := -1
	filled := 0
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		st := CellFuture
		if d.Before(now) {
			st = CellPast
			filled++
		} else if d.Equal(now) {
			st = CellCurrent
			current = i
			filled++
		}
		cells[i] = Cell{State: st, Date: d, Index: i, Opacity: 1, Filled: st == CellPast || st == CellCurrent}
		if cfg.ShowNumbers || cfg.ShowBoth {
			cells[i].Label = fmt.Sprintf("%d", d.Day())
		}
	}
	remaining := days - filled
	if remaining < 0 {
		remaining = 0
	}
	pct := float64(filled) / float64(days) * 100
	title := cfg.CustomTitle
	if title == "" {
		title = fmt.Sprintf("%d", year)
	}

	g := &Grid{
		Columns: cols, Rows: rows, Cells: cells, Current: current,
		Filled: filled, Total: days, Title: title, Unit: "day",
		Subtitle:   fmt.Sprintf("%d days · %d remaining", filled, remaining),
		StatsLeft:  fmt.Sprintf("%dd left", remaining),
		StatsRight: fmt.Sprintf("%.0f%%", pct),
		StatsPct:   pct,
	}

	if cfg.ShowAxes {
		months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		leftPos := make([]int, 12)
		for m := 0; m < 12; m++ {
			md := time.Date(year, time.Month(m+1), 1, 0, 0, 0, 0, now.Location())
			doy := md.YearDay() - 1
			leftPos[m] = doy / cols
		}
		g.Axes = &AxisLabels{
			Top: []string{"1", "7", "14"}, TopPos: []int{0, 6, 13}, TopTitle: "DAY >",
			Left: months, LeftPos: leftPos,
		}
	}
	switch cfg.Grouping {
	case model.GroupingWeek, model.GroupingWeeks, model.GroupingFortnight:
		g.Grouping = &GroupingSpec{Columns: &GroupGap{Every: 7, Gap: 2}}
	}
	applySpecialAndHeat(cfg, cells)
	return g, nil
}

// MonthBlock is one month in year-months view.
type MonthBlock struct {
	Month   time.Month
	Label   string
	Offset  int // leading empty cells
	Days    int
	Cells   []Cell
}

func buildYearMonths(cfg model.Config, now time.Time) (*Grid, error) {
	year := cfg.Year
	if year == 0 {
		year = now.Year()
	}
	ws := weekStartDay(cfg)
	// Flatten into a wide grid: 3 months per row × 4 rows, each month is 7 cols × ~6 rows
	monthW, monthH := 7, 6
	hGap, vGap := 1, 1
	if cfg.ShowAxes {
		vGap = 2 // room for month titles between rows
	}
	cols := 3*monthW + 2*hGap
	rows := 4*monthH + 3*vGap
	cells := make([]Cell, cols*rows)
	for i := range cells {
		cells[i].State = CellEmpty
		cells[i].Opacity = 1
	}
	current := -1
	filled := 0
	totalDays := 0

	for m := 0; m < 12; m++ {
		mx := m % 3
		my := m / 3
		ox := mx * (monthW + hGap)
		oy := my * (monthH + vGap)
		first := time.Date(year, time.Month(m+1), 1, 0, 0, 0, 0, now.Location())
		offset := (int(first.Weekday()) - int(ws) + 7) % 7
		daysInMonth := time.Date(year, time.Month(m+2), 0, 0, 0, 0, 0, now.Location()).Day()
		totalDays += daysInMonth
		for d := 1; d <= daysInMonth; d++ {
			idx := offset + d - 1
			r, c := idx/7, idx%7
			if r >= monthH {
				continue
			}
			gi := (oy+r)*cols + (ox + c)
			dt := time.Date(year, time.Month(m+1), d, 0, 0, 0, 0, now.Location())
			st := CellFuture
			if dt.Before(now) {
				st = CellPast
				filled++
			} else if dt.Equal(now) {
				st = CellCurrent
				current = gi
				filled++
			}
			cells[gi] = Cell{State: st, Date: dt, Index: gi, Opacity: 1, Filled: st == CellPast || st == CellCurrent}
			if cfg.ShowNumbers || cfg.ShowBoth {
				cells[gi].Label = fmt.Sprintf("%d", d)
			}
		}
	}
	remaining := totalDays - filled
	pct := float64(filled) / float64(totalDays) * 100
	title := cfg.CustomTitle
	if title == "" {
		title = fmt.Sprintf("%d", year)
	}
	g := &Grid{
		Columns: cols, Rows: rows, Cells: cells, Current: current,
		Filled: filled, Total: totalDays, Title: title, Unit: "day",
		StatsLeft: fmt.Sprintf("%dd left", remaining), StatsRight: fmt.Sprintf("%.0f%%", pct), StatsPct: pct,
	}
	if cfg.ShowAxes {
		names := []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
		placed := make([]PlacedLabel, 12)
		for m := 0; m < 12; m++ {
			mx, my := m%3, m/3
			ox := mx * (monthW + hGap)
			oy := my * (monthH + vGap)
			placed[m] = PlacedLabel{Text: names[m], Col: ox + monthW/2, Row: oy}
		}
		g.Axes = &AxisLabels{Placed: placed}
	}
	applySpecialAndHeat(cfg, cells)
	return g, nil
}

func buildQuarters(cfg model.Config, now time.Time) (*Grid, error) {
	year := cfg.Year
	if year == 0 {
		year = now.Year()
	}
	q := (int(now.Month())-1)/3 + 1
	if cfg.Scope == model.ScopeYearly {
		// All 4 quarters as continuous days with gaps — reuse year-days with fortnight grouping
		cfg2 := cfg
		cfg2.Mode = model.ModeYear
		cfg2.Grouping = model.GroupingFortnight
		return buildYearDays(cfg2, now)
	}
	// Current quarter only
	startMonth := time.Month((q-1)*3 + 1)
	start := time.Date(year, startMonth, 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 3, -1)
	days := int(end.Sub(start).Hours()/24) + 1
	cols := 7
	if cfg.Orientation == model.OrientHorizontal {
		cols = 21 // 3 months side by side conceptually as continuous
		cols = 14
	}
	rows := int(math.Ceil(float64(days) / float64(cols)))
	cells := make([]Cell, cols*rows)
	current := -1
	filled := 0
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		st := CellFuture
		if d.Before(now) {
			st = CellPast
			filled++
		} else if d.Equal(now) {
			st = CellCurrent
			current = i
			filled++
		}
		cells[i] = Cell{State: st, Date: d, Index: i, Opacity: 1, Filled: st != CellFuture}
	}
	remaining := days - filled
	pct := float64(filled) / float64(days) * 100
	title := cfg.CustomTitle
	if title == "" {
		title = fmt.Sprintf("Q%d %d", q, year)
	}
	g := &Grid{
		Columns: cols, Rows: rows, Cells: cells, Current: current,
		Filled: filled, Total: days, Title: title, Unit: "day",
		StatsLeft: fmt.Sprintf("%dd left", remaining), StatsRight: fmt.Sprintf("%.0f%%", pct), StatsPct: pct,
	}
	applySpecialAndHeat(cfg, cells)
	return g, nil
}

func buildMonthly(cfg model.Config, now time.Time) (*Grid, error) {
	year, month := now.Year(), now.Month()
	if cfg.Year != 0 {
		year = cfg.Year
	}
	ws := weekStartDay(cfg)
	first := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
	offset := (int(first.Weekday()) - int(ws) + 7) % 7
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, now.Location()).Day()
	cols, rows := 7, 6
	cells := make([]Cell, cols*rows)
	for i := range cells {
		cells[i].State = CellEmpty
		cells[i].Opacity = 1
	}
	current := -1
	filled := 0
	for d := 1; d <= daysInMonth; d++ {
		idx := offset + d - 1
		dt := time.Date(year, month, d, 0, 0, 0, 0, now.Location())
		st := CellFuture
		if dt.Before(now) {
			st = CellPast
			filled++
		} else if dt.Equal(now) {
			st = CellCurrent
			current = idx
			filled++
		}
		cells[idx] = Cell{State: st, Date: dt, Index: idx, Opacity: 1, Filled: st != CellFuture}
		if cfg.ShowNumbers || cfg.ShowBoth {
			cells[idx].Label = fmt.Sprintf("%d", d)
		}
	}
	title := cfg.CustomTitle
	if title == "" {
		title = first.Format("January 2006")
	}
	remaining := daysInMonth - filled
	pct := float64(filled) / float64(daysInMonth) * 100
	labels := []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}
	if cfg.WeekStart == model.WeekStartSunday {
		labels = []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	}
	g := &Grid{
		Columns: cols, Rows: rows, Cells: cells, Current: current,
		Filled: filled, Total: daysInMonth, Title: title, Unit: "day",
		StatsLeft: fmt.Sprintf("%dd left", remaining), StatsRight: fmt.Sprintf("%.0f%%", pct), StatsPct: pct,
	}
	if cfg.ShowAxes {
		g.Axes = &AxisLabels{Top: labels, TopPos: []int{0, 1, 2, 3, 4, 5, 6}}
	}
	applySpecialAndHeat(cfg, cells)
	return g, nil
}

func buildWeekly(cfg model.Config, now time.Time) (*Grid, error) {
	ws := weekStartDay(cfg)
	start := startOfWeek(now, ws)
	if cfg.Grouping == model.GroupingRows {
		// Vertical stack with full weekday names
		cols, rows := 1, 7
		cells := make([]Cell, 7)
		current := -1
		names := []string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"}
		if cfg.WeekStart == model.WeekStartSunday {
			names = []string{"SUNDAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY"}
		}
		for i := 0; i < 7; i++ {
			d := start.AddDate(0, 0, i)
			st := CellFuture
			if d.Before(now) {
				st = CellPast
			} else if d.Equal(now) {
				st = CellCurrent
				current = i
			}
			cells[i] = Cell{State: st, Date: d, Index: i, Opacity: 1, Filled: st != CellFuture, Label: names[i]}
		}
		title := cfg.CustomTitle
		if title == "" {
			title = "THIS WEEK"
		}
		g := &Grid{Columns: cols, Rows: rows, Cells: cells, Current: current, Filled: current + 1, Total: 7, Title: title, Unit: "day"}
		if g.Filled < 0 {
			g.Filled = 0
		}
		return g, nil
	}
	cols, rows := 7, 1
	cells := make([]Cell, 7)
	current := -1
	short := []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}
	if cfg.WeekStart == model.WeekStartSunday {
		short = []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	}
	filled := 0
	for i := 0; i < 7; i++ {
		d := start.AddDate(0, 0, i)
		st := CellFuture
		if d.Before(now) {
			st = CellPast
			filled++
		} else if d.Equal(now) {
			st = CellCurrent
			current = i
			filled++
		}
		cells[i] = Cell{State: st, Date: d, Index: i, Opacity: 1, Filled: st != CellFuture, Label: short[i]}
	}
	title := cfg.CustomTitle
	if title == "" {
		title = "THIS WEEK"
	}
	g := &Grid{
		Columns: cols, Rows: rows, Cells: cells, Current: current,
		Filled: filled, Total: 7, Title: title, Unit: "day",
		StatsLeft: fmt.Sprintf("%dd left", 7-filled), StatsRight: fmt.Sprintf("%.0f%%", float64(filled)/7*100), StatsPct: float64(filled) / 7 * 100,
	}
	if cfg.ShowAxes {
		g.Axes = &AxisLabels{Top: short, TopPos: []int{0, 1, 2, 3, 4, 5, 6}}
	}
	return g, nil
}

func customGoalTitle(cfg model.Config, fallback string) string {
	if cfg.CustomTitle != "" {
		return cfg.CustomTitle
	}
	return fallback
}

// bestPortraitCols picks a column count for a tall phone-friendly grid,
// preferring calendar-meaningful widths and a tidy last row.
func bestPortraitCols(n int, candidates []int) int {
	if n < 1 {
		n = 1
	}
	const targetAspect = 2.0 // rows/cols — lock-screen wallpaper
	bestC := candidates[0]
	bestScore := math.MaxFloat64
	for _, c := range candidates {
		if c < 1 {
			continue
		}
		// Avoid a single sparse row unless the goal is tiny.
		if c > n && n > 4 {
			continue
		}
		cols := c
		if cols > n {
			cols = n
		}
		rows := int(math.Ceil(float64(n) / float64(cols)))
		if rows < 1 {
			rows = 1
		}
		aspect := float64(rows) / float64(cols)
		empty := 0
		if rem := n % cols; rem != 0 {
			empty = cols - rem
		}
		score := math.Abs(aspect-targetAspect) + float64(empty)/float64(cols)*0.7
		// Heavy penalty for short/wide grids on phones.
		if aspect < 1.15 && n >= cols*2 {
			score += (1.15 - aspect) * 4
		}
		// Prefer familiar calendar widths when scores are close.
		switch cols {
		case 4, 7: // month-of-weeks / weekdays metaphor
			score -= 0.18
		case 3, 6: // quarters / half-year
			score -= 0.1
		}
		if score < bestScore {
			bestScore = score
			bestC = cols
		}
	}
	if bestC > n {
		bestC = n
	}
	if bestC < 1 {
		bestC = 1
	}
	return bestC
}

func weekColumns(weeks int) int {
	// 4 ≈ month of weeks, 7 ≈ week strip — keep portrait (avoid ultra-wide year rows).
	return bestPortraitCols(weeks, []int{4, 5, 6, 7, 8, 10})
}

func monthColumns(months int) int {
	// 3 ≈ quarter, 4 ≈ seasons, 6 ≈ half-year (12-wide is too squat on a phone).
	return bestPortraitCols(months, []int{3, 4, 6})
}

func buildCustom(cfg model.Config, now time.Time) (*Grid, error) {
	start, err := parseDate(cfg.StartDate)
	if err != nil {
		return nil, fmt.Errorf("startDate: %w", err)
	}
	end, err := parseDate(cfg.EndDate)
	if err != nil {
		return nil, fmt.Errorf("endDate: %w", err)
	}
	if end.Before(start) {
		start, end = end, start
	}
	now = truncateDay(now)
	dayCount := int(end.Sub(start).Hours()/24) + 1
	if dayCount < 1 {
		dayCount = 1
	}

	// Short goals: one dot per day in a 7-wide week calendar.
	if dayCount < 35 {
		return buildCustomDays(cfg, start, end, now, dayCount, 7)
	}

	ws := weekStartDay(cfg)
	weeks := weeksBetween(startOfWeek(start, ws), end) + 1
	if weeks < 1 {
		weeks = 1
	}
	// Medium goals: week dots, columns sized for a balanced portrait grid.
	if weeks <= 156 { // ~3 years
		return buildCustomWeeks(cfg, start, end, now, weekColumns(weeks))
	}

	months := monthsBetween(start, end) + 1
	if months < 1 {
		months = 1
	}
	return buildCustomMonths(cfg, start, end, now, monthColumns(months))
}

func buildCustomDays(cfg model.Config, start, end, now time.Time, total, cols int) (*Grid, error) {
	ws := weekStartDay(cfg)
	offset := 0
	if cols == 7 {
		offset = (int(start.Weekday()) - int(ws) + 7) % 7
	}
	slots := total + offset
	rows := int(math.Ceil(float64(slots) / float64(cols)))
	cells := make([]Cell, cols*rows)
	for i := range cells {
		cells[i].State = CellEmpty
		cells[i].Opacity = 1
	}
	current := -1
	filled := 0
	for i := 0; i < total; i++ {
		d := start.AddDate(0, 0, i)
		st := CellFuture
		if d.Before(now) {
			st = CellPast
			filled++
		} else if d.Equal(now) {
			st = CellCurrent
			current = offset + i
			filled++
		}
		cells[offset+i] = Cell{State: st, Date: d, Index: i, Opacity: 1, Filled: st != CellFuture}
	}
	remaining := total - filled
	pct := float64(filled) / float64(total) * 100
	g := &Grid{
		Columns: cols, Rows: rows, Cells: cells, Current: current,
		Filled: filled, Total: total, Title: customGoalTitle(cfg, fmt.Sprintf("%d days", total)), Unit: "day",
		StatsLeft: fmt.Sprintf("%dd left", remaining), StatsRight: fmt.Sprintf("%.0f%%", pct), StatsPct: pct,
	}
	if cfg.ShowAxes && cols == 7 {
		top := []string{"M", "T", "W", "T", "F", "S", "S"}
		if cfg.WeekStart == model.WeekStartSunday {
			top = []string{"S", "M", "T", "W", "T", "F", "S"}
		}
		g.Axes = &AxisLabels{Top: top, TopPos: []int{0, 1, 2, 3, 4, 5, 6}}
	}
	applySpecialAndHeat(cfg, cells)
	return g, nil
}

func buildCustomWeeks(cfg model.Config, start, end, now time.Time, cols int) (*Grid, error) {
	ws := weekStartDay(cfg)
	wStart := startOfWeek(start, ws)
	total := weeksBetween(wStart, end) + 1
	if total < 1 {
		total = 1
	}
	rows := int(math.Ceil(float64(total) / float64(cols)))
	cells := make([]Cell, cols*rows)
	for i := range cells {
		cells[i].State = CellEmpty
		cells[i].Opacity = 1
	}
	current := -1
	filled := 0
	lived := weeksBetween(wStart, now)
	for i := 0; i < total; i++ {
		d := wStart.AddDate(0, 0, i*7)
		st := CellFuture
		if i < lived {
			st = CellPast
			filled++
		} else if i == lived {
			st = CellCurrent
			current = i
			filled++
		}
		cells[i] = Cell{State: st, Date: d, Index: i, Opacity: 1, Filled: st != CellFuture}
	}
	remaining := total - filled
	pct := float64(filled) / float64(total) * 100
	g := &Grid{
		Columns: cols, Rows: rows, Cells: cells, Current: current,
		Filled: filled, Total: total, Title: customGoalTitle(cfg, "Your goal"), Unit: "week",
		StatsLeft: fmt.Sprintf("%dw left", remaining), StatsRight: fmt.Sprintf("%.0f%%", pct), StatsPct: pct,
	}
	if cfg.Grouping == model.GroupingMonths || cfg.Grouping == model.GroupingMonth {
		g.Grouping = &GroupingSpec{Columns: &GroupGap{Every: 4, Gap: 1.5}}
	}
	applySpecialAndHeat(cfg, cells)
	return g, nil
}

func monthsBetween(a, b time.Time) int {
	a = time.Date(a.Year(), a.Month(), 1, 0, 0, 0, 0, a.Location())
	b = time.Date(b.Year(), b.Month(), 1, 0, 0, 0, 0, b.Location())
	return (b.Year()-a.Year())*12 + int(b.Month()-a.Month())
}

func buildCustomMonths(cfg model.Config, start, end, now time.Time, cols int) (*Grid, error) {
	mStart := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	total := monthsBetween(start, end) + 1
	if total < 1 {
		total = 1
	}
	rows := int(math.Ceil(float64(total) / float64(cols)))
	cells := make([]Cell, cols*rows)
	for i := range cells {
		cells[i].State = CellEmpty
		cells[i].Opacity = 1
	}
	nowMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	current := -1
	filled := 0
	for i := 0; i < total; i++ {
		d := mStart.AddDate(0, i, 0)
		st := CellFuture
		if d.Before(nowMonth) {
			st = CellPast
			filled++
		} else if d.Equal(nowMonth) {
			st = CellCurrent
			current = i
			filled++
		}
		cells[i] = Cell{State: st, Date: d, Index: i, Opacity: 1, Filled: st != CellFuture}
	}
	remaining := total - filled
	pct := float64(filled) / float64(total) * 100
	g := &Grid{
		Columns: cols, Rows: rows, Cells: cells, Current: current,
		Filled: filled, Total: total, Title: customGoalTitle(cfg, "Your goal"), Unit: "month",
		StatsLeft: fmt.Sprintf("%dm left", remaining), StatsRight: fmt.Sprintf("%.0f%%", pct), StatsPct: pct,
	}
	if cfg.ShowAxes {
		g.Axes = &AxisLabels{
			Top:    []string{"J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"},
			TopPos: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		}
	}
	applySpecialAndHeat(cfg, cells)
	return g, nil
}

func habitsRowsPerHabit(n int) int {
	switch n {
	case 1:
		return 18
	case 2:
		return 8
	case 3:
		return 6
	default:
		return 4
	}
}

func buildHabits(cfg model.Config, now time.Time) (*Grid, error) {
	habits := cfg.Habits
	if len(habits) == 0 {
		habits = []model.Habit{{ID: "1", Name: "Habit", Icon: "check"}}
	}
	if len(habits) > 4 {
		habits = habits[:4]
	}
	n := len(habits)
	rph := habitsRowsPerHabit(n)
	cols := 14
	rows := rph * n
	cells := make([]Cell, cols*rows)
	for i := range cells {
		cells[i].State = CellEmpty
		cells[i].Opacity = 1
	}
	// Each habit shows the last (cols*rph) days ending today
	daysPerHabit := cols * rph
	filled := 0
	for hi, h := range habits {
		base := hi * daysPerHabit
		for d := 0; d < daysPerHabit; d++ {
			dayOffset := daysPerHabit - 1 - d
			dt := now.AddDate(0, 0, -dayOffset)
			key := dt.Format("2006-01-02")
			done := false
			if h.Completions != nil {
				done = h.Completions[key]
			}
			st := CellFuture
			if done {
				st = CellPast
				filled++
			} else if dt.Equal(now) {
				st = CellCurrent
			} else if dt.Before(now) {
				st = CellFuture // missed
			}
			cells[base+d] = Cell{State: st, Date: dt, Index: base + d, Opacity: 1, Filled: done}
		}
	}
	title := cfg.CustomTitle
	g := &Grid{
		Columns: cols, Rows: rows, Cells: cells, Current: -1,
		Filled: filled, Total: daysPerHabit * n, Title: title, Unit: "day",
		StatsLeft: fmt.Sprintf("%d habits", n), StatsRight: fmt.Sprintf("%d weeks", rph),
	}
	icons := make([]string, n)
	iconPos := make([]int, n)
	for i, h := range habits {
		icons[i] = h.Icon
		iconPos[i] = i * rph
	}
	g.Axes = &AxisLabels{LeftIcons: icons, LeftIconPos: iconPos}
	if cfg.Grouping == model.GroupingWeeks || cfg.Grouping == model.GroupingWeek {
		g.Grouping = &GroupingSpec{
			Columns: &GroupGap{Every: 7, Gap: 1.5},
			Rows:    &GroupGap{Every: rph, Gap: 3},
		}
	} else if n > 1 {
		g.Grouping = &GroupingSpec{Rows: &GroupGap{Every: rph, Gap: 3}}
	}
	return g, nil
}
