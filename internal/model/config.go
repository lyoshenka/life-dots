package model

import (
	"time"
)

// Mode is the primary calendar visualization mode.
type Mode string

const (
	ModeLife       Mode = "life"
	ModeYear       Mode = "year"
	ModeYearMonths Mode = "year-months"
	ModeQuarters   Mode = "quarters"
	ModeMonthly    Mode = "monthly"
	ModeWeekly     Mode = "weekly"
	ModeCustom     Mode = "custom"
	ModeGoal       Mode = "goal"
	ModeHabits     Mode = "habits"
)

// Scope narrows yearly views.
type Scope string

const (
	ScopeYearly    Scope = "yearly"
	ScopeQuarterly Scope = "quarterly"
	ScopeMonthly   Scope = "monthly"
	ScopeWeekly    Scope = "weekly"
)

// Grouping controls visual gaps in the grid.
type Grouping string

const (
	GroupingNone      Grouping = "none"
	GroupingDecade    Grouping = "decade"
	GroupingMonth     Grouping = "month"
	GroupingMonths    Grouping = "months"
	GroupingWeek      Grouping = "week"
	GroupingWeeks     Grouping = "weeks"
	GroupingFortnight Grouping = "fortnight"
	GroupingRows      Grouping = "rows"
	GroupingDecades   Grouping = "decades"
	GroupingFull      Grouping = "full"
)

// WeekStart is Monday or Sunday.
type WeekStart string

const (
	WeekStartMonday WeekStart = "monday"
	WeekStartSunday WeekStart = "sunday"
)

// DotShape is the visual shape of each cell.
type DotShape string

const (
	ShapeCircle  DotShape = "circle"
	ShapeRounded DotShape = "rounded"
	ShapeSquare  DotShape = "square"
)

// ShapeStyle is filled vs outlined.
type ShapeStyle string

const (
	StyleFilled   ShapeStyle = "filled"
	StyleOutlined ShapeStyle = "outlined"
)

// FillFocus chooses which dots get the foreground color.
type FillFocus string

const (
	FillPast   FillFocus = "past"
	FillFuture FillFocus = "future"
)

// BgType selects background source (URL only for photos — no stored bytes).
type BgType string

const (
	BgSolid BgType = "solid"
	BgMatch BgType = "match"
	BgURL   BgType = "url"
)

// CardStyle is the card overlay style (flat only).
type CardStyle string

const (
	CardFlat CardStyle = "flat"
)

// ScreenType is lock vs home screen padding profile.
type ScreenType string

const (
	ScreenLock ScreenType = "lock"
	ScreenHome ScreenType = "home"
)

// FooterMode controls bottom text.
type FooterMode string

const (
	FooterNone    FooterMode = "none"
	FooterNumbers FooterMode = "numbers"
)

// WidgetLayout shifts content for lock-screen widgets.
type WidgetLayout string

const (
	WidgetNone   WidgetLayout = "none"
	WidgetTop    WidgetLayout = "top"
	WidgetBottom WidgetLayout = "bottom"
	WidgetCustom WidgetLayout = "custom"
)

// Orientation for quarterly layouts.
type Orientation string

const (
	OrientHorizontal Orientation = "horizontal"
	OrientVertical   Orientation = "vertical"
)

// SpecialDate highlights a date range on the grid.
type SpecialDate struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Color     string `json:"color"`
	Label     string `json:"label,omitempty"`
}

// Habit tracks a single daily habit.
type Habit struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Icon        string            `json:"icon"`
	Mode        string            `json:"mode,omitempty"` // completion mode
	Completions map[string]bool   `json:"completions"`  // YYYY-MM-DD -> done
}

// Config is the full wallpaper / studio configuration.
type Config struct {
	// Identity
	Token string `json:"token,omitempty"`
	Name  string `json:"name,omitempty"`

	// Calendar
	Mode       Mode       `json:"mode"`
	Scope      Scope      `json:"scope,omitempty"`
	Grouping   Grouping   `json:"grouping"`
	WeekStart  WeekStart  `json:"weekStart"`
	Birthday   string     `json:"birthday,omitempty"`   // YYYY-MM-DD
	Lifespan   int        `json:"lifespan"`            // years
	Year       int        `json:"year,omitempty"`
	StartDate   string      `json:"startDate,omitempty"`
	EndDate     string      `json:"endDate,omitempty"`
	GoalTitle   string      `json:"goalTitle,omitempty"`
	CustomTitle string      `json:"customTitle,omitempty"`
	Orientation Orientation `json:"orientation,omitempty"`

	// Habits
	Habits []Habit `json:"habits,omitempty"`

	// Special dates & heatmap
	SpecialDates []SpecialDate  `json:"specialDates,omitempty"`
	DailyCounts  map[string]int `json:"dailyCounts,omitempty"`

	// Appearance
	Fg               string     `json:"fg,omitempty"` // foreground / highlighted dots
	Bg               string     `json:"bg,omitempty"` // background
	CustomColors     *Colors    `json:"customColors,omitempty"`
	DotShape         DotShape   `json:"dotShape"`
	ShapeStyle       ShapeStyle `json:"shapeStyle"`
	OutlinedFillPast bool       `json:"outlinedFillPast"`
	NoGap            bool       `json:"noGap"`
	FillFocus        FillFocus  `json:"fillFocus"` // past|future — which dots use Fg
	HighlightToday   bool       `json:"highlightToday"`
	HighlightColor   string     `json:"highlightColor,omitempty"`
	ShowNumbers      bool       `json:"showNumbers"`
	ShowBoth         bool       `json:"showBoth"` // shape + number

	// Background (URL string only — never image bytes)
	BgType     BgType  `json:"bgType"`
	BgURL      string  `json:"bgUrl,omitempty"`
	BgBlur     float64 `json:"bgBlur"`     // 0–20
	BgDim      float64 `json:"bgDim"`      // 0–80 percent
	BgFocalX   float64 `json:"bgFocalX"`   // 0–100
	BgFocalY   float64 `json:"bgFocalY"`   // 0–100
	GridOpacity float64 `json:"gridOpacity"` // 0–100

	// Card overlay
	ShowCard        bool      `json:"showCard"`
	CardStyle       CardStyle `json:"cardStyle"`
	CardColor       string    `json:"cardColor,omitempty"`
	CardOpacity     float64   `json:"cardOpacity"` // 0–100
	CardRadius      float64   `json:"cardRadius"`  // 0–50
	CardBorderW     float64   `json:"cardBorderW"`
	CardBorderColor string    `json:"cardBorderColor,omitempty"`
	CardPadH        float64   `json:"cardPadH"`
	CardPadV        float64   `json:"cardPadV"`

	// Labels & footer
	ShowLabels   bool       `json:"showLabels"`
	ShowAxes     bool       `json:"showAxes"`
	ShowTitle    bool       `json:"showTitle"` // calendar title (year, custom title, etc.)
	ShowBranding bool       `json:"showBranding"`
	ShowProgress bool       `json:"showProgress"`
	FooterMode   FooterMode `json:"footerMode"`

	// Layout / device
	DeviceID       string       `json:"deviceId"`
	Width          int          `json:"width"`
	Height         int          `json:"height"`
	Screen         ScreenType   `json:"screen"`
	WidgetLayout   WidgetLayout `json:"widgetLayout"`
	PadTop         float64      `json:"padTop"`
	PadBottom      float64      `json:"padBottom"`
	PadSides       float64      `json:"padSides"`
	CustomPadTop   float64      `json:"customPadTop"`
	CustomPadBottom float64     `json:"customPadBottom"`
	Margin         float64      `json:"margin"`     // side inset in px
	MarginTop      float64      `json:"marginTop"`  // top/bottom inset in px
	ShapeScale     float64      `json:"shapeScale"` // 50–150
	StatsScale     float64      `json:"statsScale"` // 50–150

	LockPadTopPercent    float64 `json:"lockPadTopPercent,omitempty"`
	LockPadBottomPercent float64 `json:"lockPadBottomPercent,omitempty"`
	HomePadTopPercent    float64 `json:"homePadTopPercent,omitempty"`
	HomePadBottomPercent float64 `json:"homePadBottomPercent,omitempty"`
}

// Colors is a theme color set.
type Colors struct {
	Background string `json:"background"`
	Past       string `json:"past"`       // filled / accent
	Current    string `json:"current"`
	Future     string `json:"future"`     // muted / empty
	Text       string `json:"text"`
	TextMuted  string `json:"textMuted"`
	Accent     string `json:"accent"`
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	now := time.Now()
	return Config{
		Mode:         ModeYear,
		Grouping:     GroupingFull,
		WeekStart:    WeekStartMonday,
		Birthday:     "1995-01-01",
		Lifespan:     90,
		Year:         now.Year(),
		StartDate:    now.Format("2006-01-02"),
		EndDate:      now.AddDate(0, 3, 0).Format("2006-01-02"),
		GoalTitle:      "YOUR GOAL",
		Fg:             "#FFFFFF",
		Bg:             "#000000",
		DotShape:       ShapeCircle,
		ShapeStyle:     StyleFilled,
		FillFocus:      FillPast,
		HighlightToday: true,
		HighlightColor: "#FF6B4A", // matches studio --accent
		BgType:         BgMatch,
		BgBlur:         0,
		BgDim:          0,
		BgFocalX:       50,
		BgFocalY:       50,
		GridOpacity:    100,
		ShowCard:       false,
		CardStyle:      CardFlat,
		CardOpacity:    28,
		CardRadius:     24,
		CardPadH:       24,
		CardPadV:       32,
		ShowLabels:     false,
		ShowAxes:       false,
		ShowTitle:      true,
		ShowBranding:   false,
		ShowProgress:   true,
		FooterMode:     FooterNumbers,
		DeviceID:     "iphone-16-pro",
		Width:        1206,
		Height:       2622,
		Screen:       ScreenLock,
		WidgetLayout: WidgetNone,
		Margin:       48,
		MarginTop:    48,
		ShapeScale:   100,
		StatsScale:   100,
	}
}

