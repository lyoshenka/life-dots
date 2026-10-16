package model

import (
	"encoding/json"
	"strconv"
)

// ApplyQuery overlays wallpaper URL query params onto cfg.
// Keys use snake_case (same as the studio /wallpaper.png query string).
func ApplyQuery(cfg *Config, q map[string][]string) {
	get := func(k string) string {
		if v, ok := q[k]; ok && len(v) > 0 {
			return v[0]
		}
		return ""
	}
	setBool := func(key string, dst *bool) {
		v := get(key)
		if v == "" {
			return
		}
		*dst = v == "1" || v == "true"
	}
	setInt := func(key string, dst *int) {
		v := get(key)
		if v == "" {
			return
		}
		if n, err := strconv.Atoi(v); err == nil {
			*dst = n
		}
	}
	setFloat := func(key string, dst *float64) {
		v := get(key)
		if v == "" {
			return
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			*dst = f
		}
	}

	if v := get("mode"); v != "" {
		cfg.Mode = Mode(v)
	}
	if v := get("birthday"); v != "" {
		cfg.Birthday = v
	}
	if v := get("dob"); v != "" {
		cfg.Birthday = v
	}
	setInt("lifespan", &cfg.Lifespan)
	setInt("year", &cfg.Year)
	if v := get("start_date"); v != "" {
		cfg.StartDate = v
	}
	if v := get("end_date"); v != "" {
		cfg.EndDate = v
	}
	if v := get("goal_date"); v != "" {
		cfg.EndDate = v
	}
	if v := get("title"); v != "" {
		cfg.CustomTitle = v
	}
	if v := get("fg"); v != "" {
		cfg.Fg = v
	}
	if v := get("fg_other"); v != "" {
		cfg.FgOther = v
	}
	if v := get("bg"); v != "" {
		cfg.Bg = v
	}
	if v := get("highlight_color"); v != "" {
		cfg.HighlightColor = v
	}
	if v := get("device"); v != "" {
		cfg.DeviceID = v
		if d := DeviceByID(v); d != nil {
			cfg.Width = d.Width
			cfg.Height = d.Height
		}
	}
	setInt("width", &cfg.Width)
	setInt("height", &cfg.Height)
	if v := get("bg_type"); v != "" {
		cfg.BgType = BgType(v)
	}
	if v := get("bg_url"); v != "" {
		cfg.BgType = BgURL
		cfg.BgURL = v
	}
	setInt("bg_rotate", &cfg.BgRotate)
	setFloat("bg_blur", &cfg.BgBlur)
	setFloat("bg_dim", &cfg.BgDim)
	setFloat("bg_focal_x", &cfg.BgFocalX)
	setFloat("bg_focal_y", &cfg.BgFocalY)
	if v := get("screen"); v != "" {
		cfg.Screen = ScreenType(v)
	}
	if v := get("grouping"); v != "" {
		cfg.Grouping = Grouping(v)
	}
	if v := get("week_start"); v != "" {
		cfg.WeekStart = WeekStart(v)
	}
	if v := get("shape"); v != "" {
		cfg.DotShape = DotShape(v)
	}
	if v := get("shape_style"); v != "" {
		cfg.ShapeStyle = ShapeStyle(v)
	}
	setBool("no_gap", &cfg.NoGap)
	setFloat("dot_gap", &cfg.DotGap)
	setFloat("dot_size", &cfg.DotSize)
	if v := get("fill_focus"); v != "" {
		cfg.FillFocus = FillFocus(v)
	}
	setBool("highlight_today", &cfg.HighlightToday)
	setBool("show_numbers", &cfg.ShowNumbers)
	setBool("show_both", &cfg.ShowBoth)
	setFloat("grid_opacity", &cfg.GridOpacity)
	setBool("show_card", &cfg.ShowCard)
	if v := get("card_style"); v != "" {
		cfg.CardStyle = CardStyle(v)
	}
	setFloat("card_opacity", &cfg.CardOpacity)
	setFloat("card_radius", &cfg.CardRadius)
	setFloat("card_fade", &cfg.CardFade)
	setFloat("card_pad_h", &cfg.CardPadH)
	setFloat("card_pad_v", &cfg.CardPadV)
	setBool("show_axes", &cfg.ShowAxes)
	setBool("show_title", &cfg.ShowTitle)
	setBool("show_labels", &cfg.ShowLabels)
	setBool("show_branding", &cfg.ShowBranding)
	setBool("show_progress", &cfg.ShowProgress)
	setBool("mute_text", &cfg.MuteText)
	if get("show_footer") != "" {
		setBool("show_footer", &cfg.ShowFooter)
	} else if v := get("footer_mode"); v != "" {
		// Legacy wallpaper URLs used footer_mode=none|numbers.
		cfg.ShowFooter = v != "none"
	}
	setFloat("title_size", &cfg.TitleSize)
	setFloat("footer_size", &cfg.FooterSize)
	if v := get("widget_layout"); v != "" {
		cfg.WidgetLayout = WidgetLayout(v)
	}
	setFloat("margin", &cfg.Margin)
	setFloat("margin_top", &cfg.MarginTop)
	setFloat("stats_scale", &cfg.StatsScale)
	if v := get("habits"); v != "" {
		var habits []Habit
		if err := json.Unmarshal([]byte(v), &habits); err == nil {
			cfg.Habits = habits
		}
	}
}
