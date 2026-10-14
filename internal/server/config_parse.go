package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/grin/life-dots/internal/model"
)

func parseConfigFromRequest(r *http.Request, base model.Config) model.Config {
	cfg := base
	applyQuery(&cfg, r.URL.Query())
	return cfg
}

func applyQuery(cfg *model.Config, q map[string][]string) {
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
		cfg.Mode = model.Mode(v)
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
	if v := get("goal"); v != "" {
		cfg.GoalTitle = v
	}
	if v := get("title"); v != "" {
		cfg.CustomTitle = v
	}
	if v := get("fg"); v != "" {
		cfg.Fg = v
	}
	if v := get("bg"); v != "" {
		cfg.Bg = v
	}
	if v := get("highlight_color"); v != "" {
		cfg.HighlightColor = v
	}
	if v := get("device"); v != "" {
		cfg.DeviceID = v
		if d := model.DeviceByID(v); d != nil {
			cfg.Width = d.Width
			cfg.Height = d.Height
		}
	}
	setInt("width", &cfg.Width)
	setInt("height", &cfg.Height)
	if v := get("bg_type"); v != "" {
		cfg.BgType = model.BgType(v)
	}
	if v := get("bg_url"); v != "" {
		cfg.BgType = model.BgURL
		cfg.BgURL = v
	}
	setFloat("bg_blur", &cfg.BgBlur)
	setFloat("bg_dim", &cfg.BgDim)
	setFloat("bg_focal_x", &cfg.BgFocalX)
	setFloat("bg_focal_y", &cfg.BgFocalY)
	if v := get("screen"); v != "" {
		cfg.Screen = model.ScreenType(v)
	}
	if v := get("grouping"); v != "" {
		cfg.Grouping = model.Grouping(v)
	}
	if v := get("week_start"); v != "" {
		cfg.WeekStart = model.WeekStart(v)
	}
	if v := get("shape"); v != "" {
		cfg.DotShape = model.DotShape(v)
	}
	if v := get("shape_style"); v != "" {
		cfg.ShapeStyle = model.ShapeStyle(v)
	}
	setBool("no_gap", &cfg.NoGap)
	if v := get("fill_focus"); v != "" {
		cfg.FillFocus = model.FillFocus(v)
	}
	setBool("highlight_today", &cfg.HighlightToday)
	setBool("show_numbers", &cfg.ShowNumbers)
	setBool("show_both", &cfg.ShowBoth)
	setFloat("grid_opacity", &cfg.GridOpacity)
	setBool("show_card", &cfg.ShowCard)
	if v := get("card_style"); v != "" {
		cfg.CardStyle = model.CardStyle(v)
	}
	setFloat("card_opacity", &cfg.CardOpacity)
	setFloat("card_radius", &cfg.CardRadius)
	setFloat("card_pad_h", &cfg.CardPadH)
	setFloat("card_pad_v", &cfg.CardPadV)
	setBool("show_axes", &cfg.ShowAxes)
	setBool("show_title", &cfg.ShowTitle)
	setBool("show_labels", &cfg.ShowLabels)
	setBool("show_branding", &cfg.ShowBranding)
	setBool("show_progress", &cfg.ShowProgress)
	if v := get("footer_mode"); v != "" {
		cfg.FooterMode = model.FooterMode(v)
	}
	if v := get("widget_layout"); v != "" {
		cfg.WidgetLayout = model.WidgetLayout(v)
	}
	setFloat("margin", &cfg.Margin)
	setFloat("margin_top", &cfg.MarginTop)
	setFloat("shape_scale", &cfg.ShapeScale)
	setFloat("stats_scale", &cfg.StatsScale)
	if v := get("habits"); v != "" {
		var habits []model.Habit
		if err := json.Unmarshal([]byte(v), &habits); err == nil {
			cfg.Habits = habits
		}
	}
}
